package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/voxire/onion/internal/modules/activity"
	"github.com/voxire/onion/internal/modules/aiquota"
	"github.com/voxire/onion/internal/modules/automations"
	"github.com/voxire/onion/internal/modules/booking/appointments"
	"github.com/voxire/onion/internal/modules/booking/calendar"
	"github.com/voxire/onion/internal/modules/booking/dashboard"
	"github.com/voxire/onion/internal/modules/booking/reminders"
	"github.com/voxire/onion/internal/modules/booking/resources"
	"github.com/voxire/onion/internal/modules/booking/services"
	"github.com/voxire/onion/internal/modules/booking/waitlist"
	"github.com/voxire/onion/internal/modules/cacheops"
	"github.com/voxire/onion/internal/modules/caret"
	"github.com/voxire/onion/internal/modules/caret/memory"
	"github.com/voxire/onion/internal/modules/crm/campaigns"
	"github.com/voxire/onion/internal/modules/crm/crm_sales"
	"github.com/voxire/onion/internal/modules/crm/crm_settings"
	"github.com/voxire/onion/internal/modules/crm/customers"
	"github.com/voxire/onion/internal/modules/crm/feedback"
	"github.com/voxire/onion/internal/modules/crm/loyalty"
	"github.com/voxire/onion/internal/modules/crm/segments"
	"github.com/voxire/onion/internal/modules/crm/templates"
	"github.com/voxire/onion/internal/modules/emailtest"
	"github.com/voxire/onion/internal/modules/erp/brands"
	"github.com/voxire/onion/internal/modules/erp/categories"
	"github.com/voxire/onion/internal/modules/erp/erp_settings"
	"github.com/voxire/onion/internal/modules/erp/inventory"
	"github.com/voxire/onion/internal/modules/erp/onlineorders"
	"github.com/voxire/onion/internal/modules/erp/price_lists"
	"github.com/voxire/onion/internal/modules/erp/products"
	"github.com/voxire/onion/internal/modules/erp/purchase_orders"
	"github.com/voxire/onion/internal/modules/erp/stock_takes"
	"github.com/voxire/onion/internal/modules/erp/stock_transfers"
	"github.com/voxire/onion/internal/modules/erp/store_categories"
	"github.com/voxire/onion/internal/modules/erp/store_products"
	"github.com/voxire/onion/internal/modules/erp/suppliers"
	"github.com/voxire/onion/internal/modules/erp/variants"
	"github.com/voxire/onion/internal/modules/erp/warehouses"
	"github.com/voxire/onion/internal/modules/finance/accounts"
	"github.com/voxire/onion/internal/modules/finance/bank_accounts"
	"github.com/voxire/onion/internal/modules/finance/bills"
	"github.com/voxire/onion/internal/modules/finance/expenses"
	"github.com/voxire/onion/internal/modules/finance/finance_bank_feeds"
	"github.com/voxire/onion/internal/modules/finance/finance_reconciliation"
	"github.com/voxire/onion/internal/modules/finance/finance_reports"
	"github.com/voxire/onion/internal/modules/finance/invoices"
	"github.com/voxire/onion/internal/modules/finance/journal"
	"github.com/voxire/onion/internal/modules/finance/quotes"
	"github.com/voxire/onion/internal/modules/finance/recurring_invoices"
	"github.com/voxire/onion/internal/modules/finance/tax"
	"github.com/voxire/onion/internal/modules/flow/insights"
	"github.com/voxire/onion/internal/modules/flow/workspaces"
	"github.com/voxire/onion/internal/modules/hr/benefits"
	"github.com/voxire/onion/internal/modules/hr/departments"
	"github.com/voxire/onion/internal/modules/hr/documents"
	"github.com/voxire/onion/internal/modules/hr/hr_attendance"
	"github.com/voxire/onion/internal/modules/hr/hr_core"
	"github.com/voxire/onion/internal/modules/hr/hr_dashboard"
	"github.com/voxire/onion/internal/modules/hr/hr_documents"
	"github.com/voxire/onion/internal/modules/hr/hr_equipment"
	"github.com/voxire/onion/internal/modules/hr/hr_leave"
	"github.com/voxire/onion/internal/modules/hr/hr_onboarding"
	"github.com/voxire/onion/internal/modules/hr/hr_overtime"
	"github.com/voxire/onion/internal/modules/hr/hr_payroll"
	"github.com/voxire/onion/internal/modules/hr/hr_performance"
	"github.com/voxire/onion/internal/modules/hr/hr_reports"
	"github.com/voxire/onion/internal/modules/hr/hr_setup"
	"github.com/voxire/onion/internal/modules/hr/hr_shifts"
	"github.com/voxire/onion/internal/modules/hr/hr_time"
	"github.com/voxire/onion/internal/modules/hr/one_on_ones"
	"github.com/voxire/onion/internal/modules/hr/payroll"
	"github.com/voxire/onion/internal/modules/hr/reviews"
	"github.com/voxire/onion/internal/modules/hr/staff"
	"github.com/voxire/onion/internal/modules/hr/time_off"
	"github.com/voxire/onion/internal/modules/hr/timesheets"
	"github.com/voxire/onion/internal/modules/hr/training"
	"github.com/voxire/onion/internal/modules/hub_menu"
	"github.com/voxire/onion/internal/modules/insights"
	"github.com/voxire/onion/internal/modules/insights/kpis"
	"github.com/voxire/onion/internal/modules/integrations"
	"github.com/voxire/onion/internal/modules/notifications"
	"github.com/voxire/onion/internal/modules/organization/addons"
	"github.com/voxire/onion/internal/modules/organization/api_keys"
	"github.com/voxire/onion/internal/modules/organization/apps"
	"github.com/voxire/onion/internal/modules/organization/billing"
	"github.com/voxire/onion/internal/modules/organization/business"
	"github.com/voxire/onion/internal/modules/organization/config"
	"github.com/voxire/onion/internal/modules/organization/console"
	"github.com/voxire/onion/internal/modules/organization/featuresupport"
	"github.com/voxire/onion/internal/modules/organization/oauth"
	"github.com/voxire/onion/internal/modules/organization/roles"
	"github.com/voxire/onion/internal/modules/organization/security"
	"github.com/voxire/onion/internal/modules/organization/subscriptions"
	"github.com/voxire/onion/internal/modules/organization/team"
	"github.com/voxire/onion/internal/modules/organization/user"
	"github.com/voxire/onion/internal/modules/portal"
	"github.com/voxire/onion/internal/modules/pos/audit"
	"github.com/voxire/onion/internal/modules/pos/cashmovements"
	"github.com/voxire/onion/internal/modules/pos/combos"
	"github.com/voxire/onion/internal/modules/pos/daysummary"
	"github.com/voxire/onion/internal/modules/pos/floors"
	"github.com/voxire/onion/internal/modules/pos/giftcards"
	"github.com/voxire/onion/internal/modules/pos/kds"
	"github.com/voxire/onion/internal/modules/pos/locations"
	posloyalty "github.com/voxire/onion/internal/modules/pos/loyalty"
	"github.com/voxire/onion/internal/modules/pos/manager"
	"github.com/voxire/onion/internal/modules/pos/memberlocations"
	"github.com/voxire/onion/internal/modules/pos/modifiers"
	"github.com/voxire/onion/internal/modules/pos/orders"
	"github.com/voxire/onion/internal/modules/pos/receipts"
	"github.com/voxire/onion/internal/modules/pos/sessions"
	"github.com/voxire/onion/internal/modules/pos/settings"
	"github.com/voxire/onion/internal/modules/pos/tables"
	"github.com/voxire/onion/internal/modules/pos/terminals"
	"github.com/voxire/onion/internal/modules/pos/users"
	"github.com/voxire/onion/internal/modules/recruit/careers"
	"github.com/voxire/onion/internal/modules/recruit/voxhire"
	"github.com/voxire/onion/internal/modules/store/checkout"
	"github.com/voxire/onion/internal/modules/store/payments"
	"github.com/voxire/onion/internal/modules/store/shipping"
	"github.com/voxire/onion/internal/modules/store/storefronts"
	"github.com/voxire/onion/internal/modules/store/themes"
	"github.com/voxire/onion/internal/modules/support/canned"
	"github.com/voxire/onion/internal/modules/support/customers"
	"github.com/voxire/onion/internal/modules/support/inbox"
	"github.com/voxire/onion/internal/modules/support/knowledge"
	"github.com/voxire/onion/internal/modules/support/settings"
	"github.com/voxire/onion/internal/modules/support/tickets"
	"github.com/voxire/onion/internal/modules/webhooks"
	"github.com/voxire/onion/internal/modules/websitelead"
	"github.com/voxire/onion/internal/platform/cache"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheInvalidate "github.com/voxire/onion/internal/platform/cache/invalidate"
	invalidation "github.com/voxire/onion/internal/platform/cache/invalidation"
	platformConfig "github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	emailPkg "github.com/voxire/onion/internal/platform/email"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/observability"
	"github.com/voxire/onion/internal/platform/server/probe"
	"github.com/voxire/onion/internal/platform/server/profiling"
	sheetsPkg "github.com/voxire/onion/internal/platform/sheets"
	"github.com/voxire/onion/internal/platform/slack"
	trinoPkg "github.com/voxire/onion/internal/platform/trino"
	"github.com/voxire/onion/internal/platform/utils/jwtutils"
)

// loadDotEnv is kept as a thin wrapper so existing call sites stay stable.
func loadDotEnv() {
	platformConfig.LoadDotEnv()
}

// httpServerDrainWait must exceed the HTTP server's internal drain timeout
// (20s in internal/platform/server/http) or SIGTERM kills requests that are
// still being drained.
const httpServerDrainWait = 30 * time.Second

func main() {
	loadDotEnv()
	logger.SetupLogger()

	if _, err := platformConfig.LoadAndValidate(); err != nil {
		logger.Global().Fatal("Configuration validation failed", logger.Error(err))
	}

	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := probe.Start(shutdownCtx, probe.Config{
		ServiceName: "api",
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
	if err := profiling.Start(shutdownCtx, profiling.PortFromEnv("6060")); err != nil {
		logger.Global().Warn("pprof server failed to start", logger.Error(err))
	}

	// Initialise Sentry as early as possible so even boot-time
	// panics get captured. `flush` ships any queued events before
	// the process exits — deferring here covers the restart loop
	// path too.
	flushSentry := observability.InitSentry(shutdownCtx)
	defer flushSentry()

	// Slack's flush loop is started once by logger.SetupLogger and its
	// notifier lives for the whole process. runServer is re-entered after a
	// crash, so shutting Slack down inside it would kill the only flush loop
	// on the first restart while Post kept buffering, and no periodic flush
	// would ever deliver an alert again. Deferring the shutdown here keeps the
	// loop alive across restarts and flushes exactly once at final exit.
	defer slack.Shutdown()

	for {
		if err := runServer(shutdownCtx); err != nil {
			logger.Global().Error("API server crashed, restarting in 5 seconds..", logger.Error(err))

			select {
			case <-shutdownCtx.Done():
				logger.Global().Info("Shutdown signal received during restart delay, stopping")

				return
			case <-time.After(5 * time.Second):
				continue
			}
		} else {
			return
		}
	}
}

func runServer(shutdownCtx context.Context) error {
	serverCtx, cancel := context.WithCancel(shutdownCtx)
	defer cancel()

	// recoverAndLog may queue Slack alerts while the final logger.Flush runs,
	// so those defers stay ahead of it. Slack itself is only shut down in
	// main, at process exit — see the comment there.
	defer recoverAndLog(serverCtx)
	defer logger.Flush()

	serverErr := make(chan error, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Global().Error("API server panicked, will restart", logger.Any("panic", r))

				serverErr <- &ServerPanicError{panic: r}
			}
		}()

		if err := run(serverCtx); err != nil {
			serverErr <- err

			return
		}

		serverErr <- nil
	}()

	select {
	case <-shutdownCtx.Done():
		cancel()
		logger.Global().Info("Shutdown signal received, waiting for API server to stop...")

		select {
		case <-serverErr:
			logger.Global().Info("API server stopped gracefully")

			return nil
		case <-time.After(httpServerDrainWait):
			logger.Global().Warn("API server shutdown timeout, forcing exit")

			return nil
		}
	case err := <-serverErr:
		return err
	}
}

func run(ctx context.Context) error {
	logger.Ctx(ctx).Info("Starting API server...")

	startTime := time.Now()

	err := dbclient.Connect(ctx)
	if err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}

	err = cache.Initialize(ctx)
	if err != nil {
		return fmt.Errorf("cache initialization failed: %w", err)
	}

	if err := trinoPkg.Connect(ctx); err != nil {
		return fmt.Errorf("trino initialization failed: %w", err)
	}

	ctx, err = jwtutils.InitializeJWT(ctx)
	if err != nil {
		return fmt.Errorf("jwt initialization failed: %w", err)
	}

	emailPkg.Initialize()
	sheetsPkg.Initialize(ctx)

	// This replica serves reads out of the in-memory (BigCache) layer, so it
	// must consume every invalidation channel: pattern-based, cache-group,
	// and direct local-key messages published by any replica. Without these
	// subscribers, invalidated entries keep being served stale from this
	// process's memory until their TTL lapses.
	go cacheInvalidate.SubscribeAsyncInvalidations(ctx)
	go invalidation.SubscribeCacheGroupInvalidations(ctx)
	go cacher.NewPubSub[cacher.LocalCacheInvalidationConsumer]().SubscribeContext(ctx, cacher.LocalCacheInvalidationChannel)

	readyTime := time.Now()
	startupDuration := readyTime.Sub(startTime)

	logger.Ctx(ctx).Info("API server is ready",
		logger.Any("startup_time", startupDuration.String()),
	)

	if err := startAPIServer(ctx); err != nil {
		return fmt.Errorf("api server failed: %w", err)
	}

	return nil
}

type apiModules struct {
	Accounts              *accounts.Module
	Activity              *activity.Module
	Addons                *addons.Module
	APIKeys               *apikeys.Module
	Apps                  *apps.Module
	Automations           *automations.Module
	BankAccounts          *bankaccounts.Module
	Benefits              *benefits.Module
	Billing               *billing.Module
	Subscriptions         *subscriptions.Module
	Bills                 *bills.Module
	BookingAppointments   *booking_appointments.Module
	BookingCalendar       *booking_calendar.Module
	BookingDashboard      *booking_dashboard.Module
	BookingReminders      *booking_reminders.Module
	BookingResources      *booking_resources.Module
	BookingServices       *booking_services.Module
	BookingWaitlist       *booking_waitlist.Module
	Brands                *brands.Module
	FeatureSupport        *featuresupport.Module
	Business              *business.Module
	Campaigns             *campaigns.Module
	Careers               *careers.Module
	Categories            *categories.Module
	Console               *console.Module
	Config                *config.Module
	CRMSettings           *crmsettings.Module
	CRMSales              *crmsales.Module
	Customers             *customers.Module
	Departments           *departments.Module
	Documents             *documents.Module
	CacheOps              *cacheops.Module
	AIQuota               *aiquota.Module
	Echo                  *echo.Module
	EmailTest             *emailtest.Module
	Expenses              *expenses.Module
	ERPSettings           *erpsettings.Module
	Feedback              *feedback.Module
	Portal                *portal.Module
	FinanceBankFeeds      *financebankfeeds.Module
	FinanceReconciliation *financereconciliation.Module
	FinanceReports        *financereports.Module
	FlowInsights          *flow_insights.Module
	HRAttendance          *hr_attendance.Module
	HRCore                *hr_core.Module
	HRDashboard           *hr_dashboard.Module
	HRDocuments           *hr_documents.Module
	HubMenu               *hubmenu.Module
	HREquipment           *hr_equipment.Module
	HRLeave               *hr_leave.Module
	HROnboarding          *hr_onboarding.Module
	HROvertime            *hr_overtime.Module
	HRPayroll             *hr_payroll.Module
	HRPerformance         *hr_performance.Module
	HRReports             *hr_reports.Module
	HRSetup               *hr_setup.Module
	HRShifts              *hr_shifts.Module
	HRTime                *hr_time.Module
	Insights              *insights.Module
	Integrations          *integrations.Module
	Inventory             *inventory.Module
	Invoices              *invoices.Module
	Journal               *journal.Module
	KPIs                  *kpis.Module
	Loyalty               *loyalty.Module
	Memory                *memory.Module
	Notifications         *notifications.Module
	OAuth                 *oauth.Module
	OneOnOnes             *oneonones.Module
	Payroll               *payroll.Module
	PosAudit              *audit.Module
	PosCashMovements      *cashmovements.Module
	PosCombos             *combos.Module
	PosDaySummary         *daysummary.Module
	PosFloors             *floors.Module
	PosGiftCards          *giftcards.Module
	PosKDS                *kds.Module
	PosLocations          *locations.Module
	PosLoyalty            *posloyalty.Module
	PosManager            *manager.Module
	PosMemberLocations    *memberlocations.Module
	PosModifiers          *modifiers.Module
	PosOrders             *orders.Module
	PosReceipts           *receipts.Module
	PosSessions           *sessions.Module
	PosSettings           *settings.Module
	PosTables             *tables.Module
	PosTerminals          *terminals.Module
	PosUsers              *users.Module
	PriceLists            *price_lists.Module
	Products              *products.Module
	PurchaseOrders        *purchase_orders.Module
	Quotes                *quotes.Module
	RecurringInvoices     *recurringinvoices.Module
	Reviews               *reviews.Module
	Roles                 *roles.Module
	Security              *security.Module
	Segments              *segments.Module
	Staff                 *staff.Module
	StockTakes            *stocktakes.Module
	StockTransfers        *stocktransfers.Module
	StoreCategories       *store_categories.Module
	StoreCheckout         *store_checkout.Module
	Storefronts           *storefronts.Module
	StoreOrders           *onlineorders.Module
	StorePayments         *store_payments.Module
	StoreProducts         *store_products.Module
	StoreShipping         *store_shipping.Module
	StoreThemes           *store_themes.Module
	SupportCanned         *supportcanned.Module
	SupportCustomers      *supportcustomers.Module
	SupportInbox          *supportinbox.Module
	SupportKnowledge      *supportknowledge.Module
	SupportSettings       *supportsettings.Module
	SupportTickets        *supporttickets.Module
	Suppliers             *suppliers.Module
	Tax                   *tax.Module
	Team                  *team.Module
	Templates             *templates.Module
	TimeOff               *time_off.Module
	Timesheets            *timesheets.Module
	Training              *training.Module
	User                  *user.Module
	Variants              *variants.Module
	Voxhire               *voxhire.Module
	Warehouses            *warehouses.Module
	Webhooks              *webhooks.Module
	WebsiteLead           *websitelead.Module
	Workspaces            *workspaces.Module
}
