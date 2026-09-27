package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/voxire/onion/internal/modules/accounts"
	"github.com/voxire/onion/internal/modules/automations"
	hrcore "github.com/voxire/onion/internal/modules/hr_core"
	hrleave "github.com/voxire/onion/internal/modules/hr_leave"
	hrovertime "github.com/voxire/onion/internal/modules/hr_overtime"
	hrpayroll "github.com/voxire/onion/internal/modules/hr_payroll"
	hrtime "github.com/voxire/onion/internal/modules/hr_time"
	"github.com/voxire/onion/internal/modules/journal"
	recurringinvoices "github.com/voxire/onion/internal/modules/recurring_invoices"
	"github.com/voxire/onion/internal/modules/subscriptions"
	"github.com/voxire/onion/internal/modules/team"
	"github.com/voxire/onion/internal/modules/voxhire"
	"github.com/voxire/onion/internal/modules/webhooks"
	"github.com/voxire/onion/internal/modules/websitelead"
	"github.com/voxire/onion/internal/modules/workspaces"
	platformcache "github.com/voxire/onion/internal/platform/cache"
	cacheInvalidate "github.com/voxire/onion/internal/platform/cache/invalidate"
	"github.com/voxire/onion/internal/platform/cache/invalidation"
	platformConfig "github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	emailPkg "github.com/voxire/onion/internal/platform/email"
	"github.com/voxire/onion/internal/platform/emailqueue"
	"github.com/voxire/onion/internal/platform/eventbus"
	platformjobs "github.com/voxire/onion/internal/platform/jobs"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/server/probe"
	"github.com/voxire/onion/internal/platform/server/profiling"
	"github.com/voxire/onion/internal/platform/slack"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger.SetupLogger()
	// Keep asynchronous Better Stack delivery alive across both normal returns
	// and the explicit startup failure exits below.
	defer logger.Flush()

	if _, err := platformConfig.LoadAndValidate(); err != nil {
		logger.Global().Error("config load failed", logger.Error(err))
		exitWorker(1)
	}

	// SetupLogger runs before configuration is loaded, so its first Slack
	// initialization is necessarily a no-op. Initialize again after the config
	// is available, and flush buffered alerts on process shutdown.
	slack.Init()
	defer slack.Shutdown()

	emailPkg.Initialize()

	if err := dbclient.Connect(ctx); err != nil {
		logger.Global().Error("failed to connect to database", logger.Error(err))
		exitWorker(1)
	}
	defer dbclient.Close()
	if err := platformcache.Initialize(ctx); err != nil {
		logger.Global().Error("failed to initialize cache", logger.Error(err))
		exitWorker(1)
	}
	if err := probe.Start(ctx, probe.Config{
		ServiceName: "worker",
		Port:        probe.PortFromEnv("8888"),
		Stop:        stop,
		ReadyChecks: []probe.Check{
			func(ctx context.Context) error {
				db := dbclient.Pool()
				if db == nil {
					return fmt.Errorf("database pool is not initialized")
				}
				return db.PingContext(ctx)
			},
		},
	}); err != nil {
		logger.Global().Warn("probe server failed to start", logger.Error(err))
	}
	if err := profiling.Start(ctx, profiling.PortFromEnv("6060")); err != nil {
		logger.Global().Warn("pprof server failed to start", logger.Error(err))
	}

	logger.Global().Info("worker starting")

	modules := initializeJobModules(ctx)
	platformjobs.AutoRegister(ctx, modules)
	// Not a module, so it registers itself rather than through AutoRegister.
	eventbus.RegisterJobs(platformjobs.DefaultRegistry(), eventbus.NewDispatcher(dbclient.Pool()))
	jobActivity := newWorkerJobActivity()
	if err := trackWorkerJobs(platformjobs.DefaultRegistry(), jobActivity); err != nil {
		logger.Ctx(ctx).Error("failed to instrument worker jobs", logger.Error(err))
		exitWorker(1)
	}
	logger.Ctx(ctx).Info("worker jobs registered", logger.Any("jobs", platformjobs.Names()))

	scheduler := platformjobs.NewScheduler(
		platformjobs.DefaultRegistry(),
		platformjobs.WithDefaultTimeout(platformjobs.DurationFromEnv(platformjobs.EnvJobTimeout, platformjobs.DefaultJobTimeout)),
		platformjobs.WithLocker(platformjobs.NewMySQLLocker(dbclient.Pool())),
	)
	if runJob := platformjobs.RunJobFromEnv(); runJob != "" {
		result, err := scheduler.Run(ctx, runJob)
		if err != nil {
			logger.Ctx(ctx).Error("one-shot job failed", logger.String("job_name", runJob), logger.Error(err))
			exitWorker(1)
		}
		logger.Ctx(ctx).Info("one-shot job completed",
			logger.String("job_name", runJob),
			logger.Any("duration", result.Duration.String()),
		)
		return
	}

	if platformjobs.SchedulerEnabledFromEnv() {
		if err := scheduler.Start(ctx); err != nil {
			logger.Ctx(ctx).Error("scheduler failed to start", logger.Error(err))
			exitWorker(1)
		}
	} else {
		logger.Ctx(ctx).Info("scheduler disabled by environment")
	}

	go cacheInvalidate.SubscribeAsyncInvalidations(ctx)
	go invalidation.SubscribeCacheGroupInvalidations(ctx)
	pool := emailqueue.NewPool(dbclient.Pool())
	pool.Run(ctx)
	stop()

	// The scheduler owns its cron stop goroutine, so give every registered
	// handler a chance to finish before the database is closed. The activity
	// gate also rejects jobs that cron races to dispatch after cancellation.
	if !jobActivity.stopAndWait(10 * time.Second) {
		logger.Global().Warn("worker jobs did not stop before shutdown timeout")
	}

	logger.Global().Info("worker stopped")
}

// exitWorker is used on startup failures where os.Exit would otherwise skip
// the deferred asynchronous log and Slack flushes.
func exitWorker(code int) {
	slack.Shutdown()
	logger.Flush()
	os.Exit(code)
}

// workerJobActivity is a shutdown barrier for registry handlers. Scheduler
// exposes no wait handle, so main wraps every registered handler and refuses
// new work once shutdown starts while waiting for handlers already in flight.
type workerJobActivity struct {
	mu       sync.Mutex
	active   int
	stopping bool
	idle     chan struct{}
}

func newWorkerJobActivity() *workerJobActivity {
	idle := make(chan struct{})
	close(idle)
	return &workerJobActivity{idle: idle}
}

func (a *workerJobActivity) wrap(handler platformjobs.Handler) platformjobs.Handler {
	return func(ctx context.Context) error {
		if !a.begin() {
			// A cron callback may race with cancellation. Do not start new
			// database work or turn the shutdown race into a false job error.
			return nil
		}
		defer a.finish()
		return handler(ctx)
	}
}

func (a *workerJobActivity) begin() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.stopping {
		return false
	}
	if a.active == 0 {
		a.idle = make(chan struct{})
	}
	a.active++
	return true
}

func (a *workerJobActivity) finish() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.active--
	if a.active == 0 {
		close(a.idle)
		a.idle = make(chan struct{})
	}
}

func (a *workerJobActivity) stopAndWait(timeout time.Duration) bool {
	a.mu.Lock()
	a.stopping = true
	idle := a.idle
	active := a.active
	a.mu.Unlock()

	if active == 0 {
		return true
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-idle:
		return true
	case <-timer.C:
		return false
	}
}

func trackWorkerJobs(registry *platformjobs.Registry, activity *workerJobActivity) error {
	for _, job := range registry.Jobs() {
		job.Handler = activity.wrap(job.Handler)
		if err := registry.Register(job); err != nil {
			return fmt.Errorf("track worker job %q: %w", job.Name, err)
		}
	}
	return nil
}

type jobModules struct {
	Automations       *automations.Module
	Webhooks          *webhooks.Module
	Accounts          *accounts.Module
	HRCore            *hrcore.Module
	HRLeave           *hrleave.Module
	HROvertime        *hrovertime.Module
	HRPayroll         *hrpayroll.Module
	HRTime            *hrtime.Module
	Workspaces        *workspaces.Module
	Voxhire           *voxhire.Module
	RecurringInvoices *recurringinvoices.Module
	Subscriptions     *subscriptions.Module
	Team              *team.Module
	WebsiteLead       *websitelead.Module
}

func initializeJobModules(ctx context.Context) *jobModules {
	db := dbclient.Pool()
	accountsModule := accounts.NewModule(ctx, db)
	hrCoreModule := hrcore.NewModule(ctx, db)
	hrLeaveModule := hrleave.NewModule(ctx, db, hrCoreModule)
	journalModule := journal.NewModule(ctx, db)
	journalModule.SetPayrollAccountProvider(accountsModule.PayrollAccountProvider())
	hrPayrollModule := hrpayroll.NewModule(ctx, db, hrCoreModule)
	hrPayrollModule.SetFinancePoster(journalModule.PayrollPoster())
	workspacesModule := workspaces.NewModule(ctx, db, hrCoreModule, hrLeaveModule)
	workspacesModule.Service().SetLabourCostProvider(hrPayrollModule.LabourCostProvider())
	return &jobModules{
		Automations:       automations.NewModule(ctx, db),
		Webhooks:          webhooks.NewModule(ctx, db),
		Accounts:          accountsModule,
		HRCore:            hrCoreModule,
		HRLeave:           hrLeaveModule,
		HROvertime:        hrovertime.NewModule(ctx, db, hrCoreModule),
		HRPayroll:         hrPayrollModule,
		HRTime:            hrtime.NewModule(ctx, db, hrCoreModule),
		Workspaces:        workspacesModule,
		Voxhire:           voxhire.NewModule(ctx, db),
		RecurringInvoices: recurringinvoices.NewModule(ctx, db),
		Subscriptions:     subscriptions.NewModule(ctx, db),
		Team:              team.NewModule(ctx, db),
		WebsiteLead: websitelead.NewModule(ctx, db, websitelead.Notify{
			LeadEmails:       platformConfig.GlobalConfig.LeadNotifyEmails,
			NewsletterEmails: platformConfig.GlobalConfig.NewsletterNotifyEmails,
			AudienceID:       platformConfig.GlobalConfig.ResendAudienceID,
		}),
	}
}
