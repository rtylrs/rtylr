package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/voxire/onion/internal/modules/accounts"
	"github.com/voxire/onion/internal/modules/activity"
	"github.com/voxire/onion/internal/modules/addons"
	"github.com/voxire/onion/internal/modules/aiquota"
	apikeys "github.com/voxire/onion/internal/modules/api_keys"
	"github.com/voxire/onion/internal/modules/apps"
	"github.com/voxire/onion/internal/modules/automations"
	bankaccounts "github.com/voxire/onion/internal/modules/bank_accounts"
	"github.com/voxire/onion/internal/modules/benefits"
	"github.com/voxire/onion/internal/modules/billing"
	"github.com/voxire/onion/internal/modules/bills"
	bookingappointments "github.com/voxire/onion/internal/modules/booking_appointments"
	bookingcalendar "github.com/voxire/onion/internal/modules/booking_calendar"
	bookingdashboard "github.com/voxire/onion/internal/modules/booking_dashboard"
	bookingreminders "github.com/voxire/onion/internal/modules/booking_reminders"
	bookingresources "github.com/voxire/onion/internal/modules/booking_resources"
	bookingservices "github.com/voxire/onion/internal/modules/booking_services"
	bookingwaitlist "github.com/voxire/onion/internal/modules/booking_waitlist"
	"github.com/voxire/onion/internal/modules/brands"
	"github.com/voxire/onion/internal/modules/business"
	"github.com/voxire/onion/internal/modules/cacheops"
	"github.com/voxire/onion/internal/modules/campaigns"
	"github.com/voxire/onion/internal/modules/careers"
	"github.com/voxire/onion/internal/modules/categories"
	"github.com/voxire/onion/internal/modules/config"
	"github.com/voxire/onion/internal/modules/console"
	crmsales "github.com/voxire/onion/internal/modules/crm_sales"
	crmsettings "github.com/voxire/onion/internal/modules/crm_settings"
	"github.com/voxire/onion/internal/modules/customers"
	"github.com/voxire/onion/internal/modules/departments"
	"github.com/voxire/onion/internal/modules/documents"
	"github.com/voxire/onion/internal/modules/echo"
	"github.com/voxire/onion/internal/modules/emailtest"
	erpsettings "github.com/voxire/onion/internal/modules/erp_settings"
	"github.com/voxire/onion/internal/modules/expenses"
	"github.com/voxire/onion/internal/modules/featuresupport"
	"github.com/voxire/onion/internal/modules/feedback"
	financebankfeeds "github.com/voxire/onion/internal/modules/finance_bank_feeds"
	financereconciliation "github.com/voxire/onion/internal/modules/finance_reconciliation"
	financereports "github.com/voxire/onion/internal/modules/finance_reports"
	flowinsights "github.com/voxire/onion/internal/modules/flow_insights"
	hrattendance "github.com/voxire/onion/internal/modules/hr_attendance"
	hrcore "github.com/voxire/onion/internal/modules/hr_core"
	hrdashboard "github.com/voxire/onion/internal/modules/hr_dashboard"
	hrdocuments "github.com/voxire/onion/internal/modules/hr_documents"
	hrequipment "github.com/voxire/onion/internal/modules/hr_equipment"
	hrleave "github.com/voxire/onion/internal/modules/hr_leave"
	hronboarding "github.com/voxire/onion/internal/modules/hr_onboarding"
	hrovertime "github.com/voxire/onion/internal/modules/hr_overtime"
	hrpayroll "github.com/voxire/onion/internal/modules/hr_payroll"
	hrperformance "github.com/voxire/onion/internal/modules/hr_performance"
	hrreports "github.com/voxire/onion/internal/modules/hr_reports"
	hrsetup "github.com/voxire/onion/internal/modules/hr_setup"
	hrshifts "github.com/voxire/onion/internal/modules/hr_shifts"
	hrtime "github.com/voxire/onion/internal/modules/hr_time"
	hubmenu "github.com/voxire/onion/internal/modules/hub_menu"
	"github.com/voxire/onion/internal/modules/insights"
	"github.com/voxire/onion/internal/modules/integrations"
	"github.com/voxire/onion/internal/modules/inventory"
	"github.com/voxire/onion/internal/modules/invoices"
	"github.com/voxire/onion/internal/modules/journal"
	"github.com/voxire/onion/internal/modules/kpis"
	"github.com/voxire/onion/internal/modules/loyalty"
	"github.com/voxire/onion/internal/modules/memory"
	"github.com/voxire/onion/internal/modules/notifications"
	"github.com/voxire/onion/internal/modules/oauth"
	oneonones "github.com/voxire/onion/internal/modules/one_on_ones"
	"github.com/voxire/onion/internal/modules/payroll"
	posaudit "github.com/voxire/onion/internal/modules/pos_audit"
	poscashmovements "github.com/voxire/onion/internal/modules/pos_cash_movements"
	poscombos "github.com/voxire/onion/internal/modules/pos_combos"
	poscustomers "github.com/voxire/onion/internal/modules/pos_customers"
	posdaysummary "github.com/voxire/onion/internal/modules/pos_day_summary"
	posfloors "github.com/voxire/onion/internal/modules/pos_floors"
	posgiftcards "github.com/voxire/onion/internal/modules/pos_gift_cards"
	poskds "github.com/voxire/onion/internal/modules/pos_kds"
	poslocations "github.com/voxire/onion/internal/modules/pos_locations"
	posloyalty "github.com/voxire/onion/internal/modules/pos_loyalty"
	posmanager "github.com/voxire/onion/internal/modules/pos_manager"
	posmemberlocations "github.com/voxire/onion/internal/modules/pos_member_locations"
	posmenu "github.com/voxire/onion/internal/modules/pos_menu"
	posmodifiers "github.com/voxire/onion/internal/modules/pos_modifiers"
	posorders "github.com/voxire/onion/internal/modules/pos_orders"
	posreceipts "github.com/voxire/onion/internal/modules/pos_receipts"
	possessions "github.com/voxire/onion/internal/modules/pos_sessions"
	possettings "github.com/voxire/onion/internal/modules/pos_settings"
	posstockmovements "github.com/voxire/onion/internal/modules/pos_stock_movements"
	postables "github.com/voxire/onion/internal/modules/pos_tables"
	posterminals "github.com/voxire/onion/internal/modules/pos_terminals"
	posusers "github.com/voxire/onion/internal/modules/pos_users"
	pricelists "github.com/voxire/onion/internal/modules/price_lists"
	"github.com/voxire/onion/internal/modules/products"
	purchaseorders "github.com/voxire/onion/internal/modules/purchase_orders"
	"github.com/voxire/onion/internal/modules/quotes"
	recurringinvoices "github.com/voxire/onion/internal/modules/recurring_invoices"
	"github.com/voxire/onion/internal/modules/reviews"
	"github.com/voxire/onion/internal/modules/roles"
	"github.com/voxire/onion/internal/modules/security"
	"github.com/voxire/onion/internal/modules/segments"
	"github.com/voxire/onion/internal/modules/staff"
	stocktakes "github.com/voxire/onion/internal/modules/stock_takes"
	stocktransfers "github.com/voxire/onion/internal/modules/stock_transfers"
	storecategories "github.com/voxire/onion/internal/modules/store_categories"
	storecheckout "github.com/voxire/onion/internal/modules/store_checkout"
	storeorders "github.com/voxire/onion/internal/modules/store_orders"
	storepayments "github.com/voxire/onion/internal/modules/store_payments"
	storeproducts "github.com/voxire/onion/internal/modules/store_products"
	storeshipping "github.com/voxire/onion/internal/modules/store_shipping"
	storethemes "github.com/voxire/onion/internal/modules/store_themes"
	"github.com/voxire/onion/internal/modules/storefronts"
	"github.com/voxire/onion/internal/modules/subscriptions"
	"github.com/voxire/onion/internal/modules/suppliers"
	supportcanned "github.com/voxire/onion/internal/modules/support_canned"
	supportcustomers "github.com/voxire/onion/internal/modules/support_customers"
	supportinbox "github.com/voxire/onion/internal/modules/support_inbox"
	supportknowledge "github.com/voxire/onion/internal/modules/support_knowledge"
	supportsettings "github.com/voxire/onion/internal/modules/support_settings"
	supporttickets "github.com/voxire/onion/internal/modules/support_tickets"
	"github.com/voxire/onion/internal/modules/tax"
	"github.com/voxire/onion/internal/modules/team"
	"github.com/voxire/onion/internal/modules/templates"
	timeoff "github.com/voxire/onion/internal/modules/time_off"
	"github.com/voxire/onion/internal/modules/timesheets"
	"github.com/voxire/onion/internal/modules/training"
	"github.com/voxire/onion/internal/modules/user"
	"github.com/voxire/onion/internal/modules/variants"
	"github.com/voxire/onion/internal/modules/voxhire"
	"github.com/voxire/onion/internal/modules/warehouses"
	"github.com/voxire/onion/internal/modules/webhooks"
	"github.com/voxire/onion/internal/modules/websitelead"
	"github.com/voxire/onion/internal/modules/workspaces"
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
	BookingAppointments   *bookingappointments.Module
	BookingCalendar       *bookingcalendar.Module
	BookingDashboard      *bookingdashboard.Module
	BookingReminders      *bookingreminders.Module
	BookingResources      *bookingresources.Module
	BookingServices       *bookingservices.Module
	BookingWaitlist       *bookingwaitlist.Module
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
	FinanceBankFeeds      *financebankfeeds.Module
	FinanceReconciliation *financereconciliation.Module
	FinanceReports        *financereports.Module
	FlowInsights          *flowinsights.Module
	HRAttendance          *hrattendance.Module
	HRCore                *hrcore.Module
	HRDashboard           *hrdashboard.Module
	HRDocuments           *hrdocuments.Module
	HubMenu               *hubmenu.Module
	HREquipment           *hrequipment.Module
	HRLeave               *hrleave.Module
	HROnboarding          *hronboarding.Module
	HROvertime            *hrovertime.Module
	HRPayroll             *hrpayroll.Module
	HRPerformance         *hrperformance.Module
	HRReports             *hrreports.Module
	HRSetup               *hrsetup.Module
	HRShifts              *hrshifts.Module
	HRTime                *hrtime.Module
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
	PosAudit              *posaudit.Module
	PosCashMovements      *poscashmovements.Module
	PosCombos             *poscombos.Module
	PosCustomers          *poscustomers.Module
	PosDaySummary         *posdaysummary.Module
	PosFloors             *posfloors.Module
	PosGiftCards          *posgiftcards.Module
	PosKDS                *poskds.Module
	PosLocations          *poslocations.Module
	PosLoyalty            *posloyalty.Module
	PosManager            *posmanager.Module
	PosMemberLocations    *posmemberlocations.Module
	PosMenu               *posmenu.Module
	PosModifiers          *posmodifiers.Module
	PosOrders             *posorders.Module
	PosReceipts           *posreceipts.Module
	PosSessions           *possessions.Module
	PosSettings           *possettings.Module
	PosStockMovements     *posstockmovements.Module
	PosTables             *postables.Module
	PosTerminals          *posterminals.Module
	PosUsers              *posusers.Module
	PriceLists            *pricelists.Module
	Products              *products.Module
	PurchaseOrders        *purchaseorders.Module
	Quotes                *quotes.Module
	RecurringInvoices     *recurringinvoices.Module
	Reviews               *reviews.Module
	Roles                 *roles.Module
	Security              *security.Module
	Segments              *segments.Module
	Staff                 *staff.Module
	StockTakes            *stocktakes.Module
	StockTransfers        *stocktransfers.Module
	StoreCategories       *storecategories.Module
	StoreCheckout         *storecheckout.Module
	Storefronts           *storefronts.Module
	StoreOrders           *storeorders.Module
	StorePayments         *storepayments.Module
	StoreProducts         *storeproducts.Module
	StoreShipping         *storeshipping.Module
	StoreThemes           *storethemes.Module
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
	TimeOff               *timeoff.Module
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
