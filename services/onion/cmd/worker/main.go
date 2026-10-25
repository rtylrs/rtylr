package main

import (
	"github.com/voxire/onion/internal/modules/activity"
	"github.com/voxire/onion/internal/modules/booking"
	"github.com/voxire/onion/internal/modules/crm"
	"github.com/voxire/onion/internal/modules/erp"
	"github.com/voxire/onion/internal/modules/finance"
	flowreferences "github.com/voxire/onion/internal/modules/flow/references"
	identityadapter "github.com/voxire/onion/internal/modules/flow/workspaces/adapters/identity"
	"github.com/voxire/onion/internal/modules/pos"
	posreferences "github.com/voxire/onion/internal/modules/pos/references"
	"github.com/voxire/onion/internal/modules/store"
	"github.com/voxire/onion/internal/modules/support"
	"github.com/voxire/onion/internal/platform/authorization"

	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/voxire/onion/internal/modules/automations"
	crmsales "github.com/voxire/onion/internal/modules/crm/crm_sales"
	"github.com/voxire/onion/internal/modules/crm/customers"
	"github.com/voxire/onion/internal/modules/erp/inventory"
	"github.com/voxire/onion/internal/modules/finance/accounts"
	"github.com/voxire/onion/internal/modules/finance/journal"
	recurringinvoices "github.com/voxire/onion/internal/modules/finance/recurring_invoices"
	"github.com/voxire/onion/internal/modules/flow/workspaces"
	"github.com/voxire/onion/internal/modules/hr/hr_core"
	"github.com/voxire/onion/internal/modules/hr/hr_leave"
	"github.com/voxire/onion/internal/modules/hr/hr_overtime"
	"github.com/voxire/onion/internal/modules/hr/hr_payroll"
	"github.com/voxire/onion/internal/modules/hr/hr_time"
	"github.com/voxire/onion/internal/modules/organization/business"
	"github.com/voxire/onion/internal/modules/organization/scopes"
	"github.com/voxire/onion/internal/modules/organization/subscriptions"
	"github.com/voxire/onion/internal/modules/organization/team"
	teamadapters "github.com/voxire/onion/internal/modules/organization/team/adapters"
	"github.com/voxire/onion/internal/modules/pos/giftcards"
	"github.com/voxire/onion/internal/modules/pos/locations"
	"github.com/voxire/onion/internal/modules/recruit/voxhire"
	"github.com/voxire/onion/internal/modules/webhooks"
	"github.com/voxire/onion/internal/modules/websitelead"
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
	"github.com/voxire/onion/internal/platform/tenant"
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
	eventbus.RegisterJobs(platformjobs.DefaultRegistry(), eventbus.NewDispatcher(
		dbclient.Pool(), modules.Automations, modules.Webhooks, modules.Customers, modules.CRMSales, modules.PosGiftCards, modules.Inventory, activity.NewModule(ctx, dbclient.Pool()).WithUserNameReader(scopes.NewReader(dbclient.Pool()).BusinessUserName),
	).WithSources(booking.Facts(dbclient.Pool()), crm.Facts(dbclient.Pool()), erp.Facts(dbclient.Pool()), finance.Facts(dbclient.Pool()), pos.Facts(dbclient.Pool()), store.Facts(dbclient.Pool()), support.Facts(dbclient.Pool()), team.Facts(dbclient.Pool())))
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
	pool := emailqueue.NewPool(dbclient.Pool(), func(ctx context.Context, job emailqueue.Job) error {
		_, err := emailPkg.Send(ctx, emailPkg.SendParams{
			To:      []string{job.To},
			Subject: job.Subject,
			HTML:    job.HTML,
			Text:    job.Text,
			From:    job.From,
			ReplyTo: job.ReplyTo,
		})
		return err
	})
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
	Customers         *customers.Module
	CRMSales          *crmsales.Module
	PosGiftCards      *giftcards.Module
	Inventory         *inventory.Module
	Webhooks          *webhooks.Module
	Accounts          *accounts.Module
	HRCore            *hr_core.Module
	HRLeave           *hr_leave.Module
	HROvertime        *hr_overtime.Module
	HRPayroll         *hr_payroll.Module
	HRTime            *hr_time.Module
	Workspaces        *workspaces.Module
	Voxhire           *voxhire.Module
	RecurringInvoices *recurringinvoices.Module
	Subscriptions     *subscriptions.Module
	Team              *team.Module
	WebsiteLead       *websitelead.Module
}

func initializeJobModules(ctx context.Context) *jobModules {
	tenant.SetScopeResolver(scopes.NewReader(dbclient.Pool()))
	flowReferenceReader := flowreferences.NewReader(dbclient.Pool())
	posReferenceReader := posreferences.NewReader(dbclient.Pool())
	_ = authorization.SetScopeLookup(func(ctx context.Context, businessID int64, scopeType string, id int64) (bool, error) {
		if scopeType == "flow_project" {
			return flowReferenceReader.ProjectExists(ctx, businessID, id)
		}
		return posReferenceReader.LocationExists(ctx, businessID, id)
	})

	db := dbclient.Pool()
	accountsModule := accounts.NewModule(ctx, db)
	hrCoreModule := hr_core.NewModule(ctx, db).WithAccountEmailReader(scopes.NewReader(db).AccountEmail)
	hrLeaveModule := hr_leave.NewModule(ctx, db, hrCoreModule)
	journalModule := journal.NewModule(ctx, db)
	journalModule.SetPayrollAccountProvider(accountsModule.PayrollAccountProvider())
	hrPayrollModule := hr_payroll.NewModule(ctx, db, hrCoreModule)
	hrPayrollModule.SetFinancePoster(journalModule.PayrollPoster())
	workspacesModule := workspaces.NewModule(ctx, db).WithIdentityReaders(identityadapter.NewIdentityReader(scopes.NewReader(db)), scopes.NewReader(db).AccountByEmail, scopes.NewReader(db).BusinessUserByUsername)
	workspacesModule.Service().SetLabourCostProvider(hrPayrollModule.LabourCostProvider())
	teamModule := team.NewModule(ctx, db)
	teamModule.SetTerminatedEmployeeSource(teamadapters.NewTerminatedEmployeeSource(hrCoreModule))
	businessModule := business.NewModule(ctx, db)
	customersModule := customers.NewModule(ctx, db).WithReportingTimezone(businessModule.ReportingTimezone).WithBusinessOwnerResolver(businessModule.GetBusinessOwnerID)
	automationContacts := func(ctx context.Context, customerID, businessID int64) (automations.CustomerContact, bool, error) {
		recipient, found, err := customersModule.ResolveEmailRecipient(ctx, businessID, customerID)
		return automations.CustomerContact{Email: recipient.Email, Name: recipient.Name}, found, err
	}
	webhooksModule := webhooks.NewModule(ctx, db)
	webhooksModule.SetNotificationContactResolver(teamModule.GetActiveMemberContact)
	return &jobModules{
		Automations:       automations.NewModule(ctx, db, automationContacts),
		Customers:         customersModule,
		CRMSales:          crmsales.NewModule(ctx, db).WithMembershipReader(scopes.NewReader(db).IsBusinessUser),
		PosGiftCards:      giftcards.NewModule(ctx, db),
		Inventory:         inventory.NewModule(ctx, db).WithPOSBranchReader(locations.NewModule(ctx, db).WithLocationLimitReader(scopes.NewReader(db).LocationLimit).StockBranchExists),
		Webhooks:          webhooksModule,
		Accounts:          accountsModule,
		HRCore:            hrCoreModule,
		HRLeave:           hrLeaveModule,
		HROvertime:        hr_overtime.NewModule(ctx, db, hrCoreModule),
		HRPayroll:         hrPayrollModule,
		HRTime:            hr_time.NewModule(ctx, db, hrCoreModule),
		Workspaces:        workspacesModule,
		Voxhire:           voxhire.NewModule(ctx, db),
		RecurringInvoices: recurringinvoices.NewModule(ctx, db),
		Subscriptions:     subscriptions.NewModule(ctx, db),
		Team:              teamModule,
		WebsiteLead: websitelead.NewModule(ctx, db, websitelead.Notify{
			LeadEmails:       platformConfig.GlobalConfig.LeadNotifyEmails,
			NewsletterEmails: platformConfig.GlobalConfig.NewsletterNotifyEmails,
			AudienceID:       platformConfig.GlobalConfig.ResendAudienceID,
		}),
	}
}
