package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

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
	"github.com/voxire/onion/internal/modules/voxhire"
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
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger.SetupLogger()

	if _, err := platformConfig.LoadAndValidate(); err != nil {
		logger.Global().Error("config load failed", logger.Error(err))
		os.Exit(1)
	}

	emailPkg.Initialize()

	if err := dbclient.Connect(ctx); err != nil {
		logger.Global().Error("failed to connect to database", logger.Error(err))
		os.Exit(1)
	}
	defer dbclient.Close()
	if err := platformcache.Initialize(ctx); err != nil {
		logger.Global().Error("failed to initialize cache", logger.Error(err))
		os.Exit(1)
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
			os.Exit(1)
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
			os.Exit(1)
		}
	} else {
		logger.Ctx(ctx).Info("scheduler disabled by environment")
	}

	go cacheInvalidate.SubscribeAsyncInvalidations(ctx)
	go invalidation.SubscribeCacheGroupInvalidations(ctx)
	pool := emailqueue.NewPool(dbclient.Pool())
	pool.Run(ctx)

	logger.Global().Info("worker stopped")
}

type jobModules struct {
	Automations       *automations.Module
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
	}
}
