package main

import (
	"context"
	"fmt"
	stdhttp "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/accounts"
	"github.com/voxire/onion/internal/modules/activity"
	"github.com/voxire/onion/internal/modules/addons"
	"github.com/voxire/onion/internal/modules/aiquery"
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
	hrdashboardstats "github.com/voxire/onion/internal/modules/hr_dashboard/stats"
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
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/cache"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheInvalidate "github.com/voxire/onion/internal/platform/cache/invalidate"
	invalidation "github.com/voxire/onion/internal/platform/cache/invalidation"
	platformConfig "github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	emailPkg "github.com/voxire/onion/internal/platform/email"
	"github.com/voxire/onion/internal/platform/license"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/observability"
	"github.com/voxire/onion/internal/platform/server/http"
	"github.com/voxire/onion/internal/platform/server/probe"
	"github.com/voxire/onion/internal/platform/server/profiling"
	sheetsPkg "github.com/voxire/onion/internal/platform/sheets"
	"github.com/voxire/onion/internal/platform/slack"
	trinoPkg "github.com/voxire/onion/internal/platform/trino"
	"github.com/voxire/onion/internal/platform/utils/jwtutils"
	"github.com/voxire/onion/internal/platform/wsbus"
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

	// Declared first so it runs last: recoverAndLog may still queue Slack
	// alerts that the final flush must deliver.
	defer slack.Shutdown()
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

func initializeModules(ctx context.Context) *apiModules {
	memoryModule := memory.NewModule(ctx, dbclient.Pool())
	aiQuotaModule := aiquota.NewModule(ctx, dbclient.Pool())
	aiQueryModule := aiquery.NewModule(ctx)
	hrCoreModule := hrcore.NewModule(ctx, dbclient.Pool())
	hrAttendanceModule := hrattendance.NewModule(ctx, dbclient.Pool(), hrCoreModule)
	hrLeaveModule := hrleave.NewModule(ctx, dbclient.Pool(), hrCoreModule)
	hrOvertimeModule := hrovertime.NewModule(ctx, dbclient.Pool(), hrCoreModule)
	hrPayrollModule := hrpayroll.NewModule(ctx, dbclient.Pool(), hrCoreModule)
	hrTimeModule := hrtime.NewModule(ctx, dbclient.Pool(), hrCoreModule)
	customersModule := customers.NewModule(ctx, dbclient.Pool())
	campaignsModule := campaigns.NewModule(ctx, dbclient.Pool())
	inventoryModule := inventory.NewModule(ctx, dbclient.Pool())
	productsModule := products.NewModule(ctx, dbclient.Pool())
	categoriesModule := categories.NewModule(ctx, dbclient.Pool())
	brandsModule := brands.NewModule(ctx, dbclient.Pool())
	loyaltyModule := loyalty.NewModule(ctx, dbclient.Pool())
	reviewsModule := reviews.NewModule(ctx, dbclient.Pool())
	segmentsModule := segments.NewModule(ctx, dbclient.Pool())
	// flowInsights is constructed first: insights.dashboard.v1's Flow-shaped
	// projections delegate to it rather than querying flow_* directly.
	flowInsightsModule := flowinsights.NewModule(ctx, dbclient.Pool())
	insightsModule := insights.NewModule(ctx, dbclient.Pool(), flowInsightsModule)
	workspacesModule := workspaces.NewModule(ctx, dbclient.Pool(), hrCoreModule, hrLeaveModule)
	workspacesModule.Service().SetLabourCostProvider(hrPayrollModule.LabourCostProvider())
	expensesModule := expenses.NewModule(ctx, dbclient.Pool())
	posOrdersModule := posorders.NewModule(ctx, dbclient.Pool())
	posLocationsModule := poslocations.NewModule(ctx, dbclient.Pool())
	posTerminalsModule := posterminals.NewModule(ctx, dbclient.Pool())
	posUsersModule := posusers.NewModule(ctx, dbclient.Pool())
	posFloorsModule := posfloors.NewModule(ctx, dbclient.Pool())
	posTablesModule := postables.NewModule(ctx, dbclient.Pool())
	posMenuModule := posmenu.NewModule(ctx, dbclient.Pool())
	posModifiersModule := posmodifiers.NewModule(ctx, dbclient.Pool())
	posSessionsModule := possessions.NewModule(ctx, dbclient.Pool())
	posCashMovementsModule := poscashmovements.NewModule(ctx, dbclient.Pool())
	posStockMovementsModule := posstockmovements.NewModule(ctx, dbclient.Pool())
	suppliersModule := suppliers.NewModule(ctx, dbclient.Pool())
	warehousesModule := warehouses.NewModule(ctx, dbclient.Pool())
	purchaseOrdersModule := purchaseorders.NewModule(ctx, dbclient.Pool())
	stockTransfersModule := stocktransfers.NewModule(ctx, dbclient.Pool())
	bankAccountsModule := bankaccounts.NewModule(ctx, dbclient.Pool())
	billsModule := bills.NewModule(ctx, dbclient.Pool())
	quotesModule := quotes.NewModule(ctx, dbclient.Pool())
	bookingDashboardModule := bookingdashboard.NewModule(ctx, dbclient.Pool())
	bookingAppointmentsModule := bookingappointments.NewModule(ctx, dbclient.Pool())
	bookingCalendarModule := bookingcalendar.NewModule(ctx, dbclient.Pool())
	bookingRemindersModule := bookingreminders.NewModule(ctx, dbclient.Pool())
	bookingResourcesModule := bookingresources.NewModule(ctx, dbclient.Pool())
	bookingServicesModule := bookingservices.NewModule(ctx, dbclient.Pool())
	bookingWaitlistModule := bookingwaitlist.NewModule(ctx, dbclient.Pool())
	financeBankFeedsModule := financebankfeeds.NewModule(ctx, dbclient.Pool())
	financeReconciliationModule := financereconciliation.NewModule(ctx, dbclient.Pool())
	financeReportsModule := financereports.NewModule(ctx, dbclient.Pool())
	accountsModule := accounts.NewModule(ctx, dbclient.Pool())
	journalModule := journal.NewModule(ctx, dbclient.Pool())
	journalModule.SetPayrollAccountProvider(accountsModule.PayrollAccountProvider())
	hrPayrollModule.SetFinancePoster(journalModule.PayrollPoster())
	kpisModule := kpis.NewModule(ctx, dbclient.Pool())
	posDaySummaryModule := posdaysummary.NewModule(ctx, dbclient.Pool())
	invoicesModule := invoices.NewModule(ctx, dbclient.Pool())
	// CRM->FIN-02 (#840): invoices verify a supplied customer_id against the
	// customers module and snapshot its name/email/address onto the invoice.
	// Port lives in invoices; customers implements it; no package imports.
	invoicesModule.SetCustomerSnapshotProvider(customersModule.SnapshotProvider())

	storeCategoriesModule := storecategories.NewModule(ctx, dbclient.Pool())
	storeOrdersModule := storeorders.NewModule(ctx, dbclient.Pool())
	storeProductsModule := storeproducts.NewModule(ctx, dbclient.Pool())
	supportCannedModule := supportcanned.NewModule(ctx, dbclient.Pool())
	supportCustomersModule := supportcustomers.NewModule(ctx, dbclient.Pool())
	supportInboxModule := supportinbox.NewModule(ctx, dbclient.Pool())
	supportKnowledgeModule := supportknowledge.NewModule(ctx, dbclient.Pool())
	supportTicketsModule := supporttickets.NewModule(ctx, dbclient.Pool())
	taxModule := tax.NewModule(ctx, dbclient.Pool())

	return &apiModules{
		Accounts:              accountsModule,
		Activity:              activity.NewModule(ctx, dbclient.Pool()),
		Addons:                addons.NewModule(ctx, dbclient.Pool()),
		APIKeys:               apikeys.NewModule(ctx, dbclient.Pool()),
		Apps:                  apps.NewModule(ctx, dbclient.Pool()),
		Automations:           automations.NewModule(ctx, dbclient.Pool()),
		BankAccounts:          bankAccountsModule,
		Benefits:              benefits.NewModule(ctx, dbclient.Pool(), hrCoreModule),
		Billing:               billing.NewModule(ctx, dbclient.Pool()),
		Subscriptions:         subscriptions.NewModule(ctx, dbclient.Pool()),
		Bills:                 billsModule,
		BookingAppointments:   bookingAppointmentsModule,
		BookingCalendar:       bookingCalendarModule,
		BookingDashboard:      bookingDashboardModule,
		BookingReminders:      bookingRemindersModule,
		BookingResources:      bookingResourcesModule,
		BookingServices:       bookingServicesModule,
		BookingWaitlist:       bookingWaitlistModule,
		Brands:                brandsModule,
		Business:              business.NewModule(ctx, dbclient.Pool()),
		Campaigns:             campaignsModule,
		Careers:               careers.NewModule(ctx, dbclient.Pool()),
		Categories:            categoriesModule,
		Console:               console.NewModule(ctx, dbclient.Pool()),
		Config:                config.NewModule(ctx, dbclient.Pool()),
		CRMSettings:           crmsettings.NewModule(ctx, dbclient.Pool()),
		CRMSales:              crmsales.NewModule(ctx, dbclient.Pool()),
		Customers:             customersModule,
		FeatureSupport:        featuresupport.NewModule(ctx, dbclient.Pool()),
		Departments:           departments.NewModule(ctx, dbclient.Pool()),
		Documents:             documents.NewModule(ctx, dbclient.Pool()),
		CacheOps:              cacheops.NewModule(),
		AIQuota:               aiQuotaModule,
		Echo:                  echo.NewModule(ctx, memoryModule, aiQuotaModule, aiQueryModule, memoryModule, insightsModule, financeReportsModule, flowInsightsModule, bookingDashboardModule, posDaySummaryModule, kpisModule, workspacesModule, customersModule, campaignsModule, segmentsModule, loyaltyModule, reviewsModule, inventoryModule, productsModule, categoriesModule, brandsModule, suppliersModule, warehousesModule, purchaseOrdersModule, stockTransfersModule, expensesModule, posOrdersModule, posLocationsModule, posTerminalsModule, posUsersModule, posFloorsModule, posTablesModule, posMenuModule, posModifiersModule, posSessionsModule, posCashMovementsModule, posStockMovementsModule, hrCoreModule, hrAttendanceModule, hrLeaveModule, hrOvertimeModule, hrPayrollModule, hrTimeModule, billsModule, quotesModule, bookingAppointmentsModule, bookingCalendarModule, bookingRemindersModule, bookingResourcesModule, bookingServicesModule, bookingWaitlistModule, bankAccountsModule, financeBankFeedsModule, financeReconciliationModule, journalModule, taxModule, invoicesModule, storeOrdersModule, storeProductsModule, storeCategoriesModule, supportTicketsModule, supportInboxModule, supportCustomersModule, supportCannedModule, supportKnowledgeModule),
		EmailTest:             emailtest.NewModule(ctx, dbclient.Pool()),
		Expenses:              expensesModule,
		Feedback:              feedback.NewModule(ctx, dbclient.Pool()),
		FinanceBankFeeds:      financeBankFeedsModule,
		FinanceReconciliation: financeReconciliationModule,
		FinanceReports:        financeReportsModule,
		FlowInsights:          flowInsightsModule,
		HRAttendance:          hrAttendanceModule,
		HRCore:                hrCoreModule,
		HRDashboard:           hrdashboard.NewModule(ctx, dbclient.Pool()),
		HRDocuments:           hrdocuments.NewModule(ctx, dbclient.Pool()),
		HubMenu:               hubmenu.NewModule(ctx, dbclient.Pool()),
		HREquipment:           hrequipment.NewModule(ctx, dbclient.Pool(), hrCoreModule),
		HRLeave:               hrLeaveModule,
		HROnboarding:          hronboarding.NewModule(ctx, dbclient.Pool()),
		HROvertime:            hrOvertimeModule,
		HRPayroll:             hrPayrollModule,
		HRPerformance:         hrperformance.NewModule(ctx, dbclient.Pool()),
		HRReports:             hrreports.NewModule(ctx, dbclient.Pool()),
		HRSetup:               hrsetup.NewModule(ctx, dbclient.Pool(), hrCoreModule),
		HRShifts:              hrshifts.NewModule(ctx, dbclient.Pool(), hrCoreModule),
		HRTime:                hrTimeModule,
		Insights:              insightsModule,
		Integrations:          integrations.NewModule(ctx, dbclient.Pool()),
		Inventory:             inventoryModule,
		Invoices:              invoicesModule,
		Journal:               journalModule,
		KPIs:                  kpisModule,
		Loyalty:               loyaltyModule,
		Memory:                memoryModule,
		Notifications:         notifications.NewModule(ctx, dbclient.Pool()),
		OAuth:                 oauth.NewModule(ctx, dbclient.Pool(), platformConfig.GlobalConfig.AuthAppURL),
		OneOnOnes:             oneonones.NewModule(ctx, dbclient.Pool()),
		Payroll:               payroll.NewModule(ctx, dbclient.Pool(), hrCoreModule),
		PosAudit:              posaudit.NewModule(ctx, dbclient.Pool()),
		PosCashMovements:      posCashMovementsModule,
		PosCombos:             poscombos.NewModule(ctx, dbclient.Pool()),
		PosCustomers:          poscustomers.NewModule(ctx, dbclient.Pool()),
		PosDaySummary:         posDaySummaryModule,
		PosFloors:             posFloorsModule,
		PosGiftCards:          posgiftcards.NewModule(ctx, dbclient.Pool()),
		PosKDS:                poskds.NewModule(ctx, dbclient.Pool()),
		PosLocations:          posLocationsModule,
		PosLoyalty:            posloyalty.NewModule(ctx, dbclient.Pool()),
		PosManager:            posmanager.NewModule(ctx, dbclient.Pool()),
		PosMemberLocations:    posmemberlocations.NewModule(ctx, dbclient.Pool()),
		PosMenu:               posMenuModule,
		PosModifiers:          posModifiersModule,
		PosOrders:             posOrdersModule,
		PosReceipts:           posreceipts.NewModule(ctx, dbclient.Pool()),
		PosSessions:           posSessionsModule,
		PosSettings:           possettings.NewModule(ctx, dbclient.Pool()),
		PosStockMovements:     posStockMovementsModule,
		PosTables:             posTablesModule,
		PosTerminals:          posTerminalsModule,
		PosUsers:              posUsersModule,
		PriceLists:            pricelists.NewModule(ctx, dbclient.Pool()),
		Products:              productsModule,
		PurchaseOrders:        purchaseOrdersModule,
		Quotes:                quotesModule,
		RecurringInvoices:     recurringinvoices.NewModule(ctx, dbclient.Pool()),
		Reviews:               reviewsModule,
		Roles:                 roles.NewModule(ctx, dbclient.Pool()),
		Security:              security.NewModule(ctx, dbclient.Pool()),
		Segments:              segmentsModule,
		Staff:                 staff.NewModule(ctx, dbclient.Pool(), hrCoreModule),
		StockTakes:            stocktakes.NewModule(ctx, dbclient.Pool()),
		StockTransfers:        stockTransfersModule,
		StoreCategories:       storeCategoriesModule,
		StoreCheckout:         storecheckout.NewModule(ctx, dbclient.Pool()),
		Storefronts:           storefronts.NewModule(ctx, dbclient.Pool()),
		StoreOrders:           storeOrdersModule,
		StorePayments:         storepayments.NewModule(ctx, dbclient.Pool()),
		StoreProducts:         storeProductsModule,
		StoreShipping:         storeshipping.NewModule(ctx, dbclient.Pool()),
		StoreThemes:           storethemes.NewModule(ctx, dbclient.Pool()),
		SupportCanned:         supportCannedModule,
		SupportCustomers:      supportCustomersModule,
		SupportInbox:          supportInboxModule,
		SupportKnowledge:      supportKnowledgeModule,
		SupportSettings:       supportsettings.NewModule(ctx, dbclient.Pool()),
		SupportTickets:        supportTicketsModule,
		Suppliers:             suppliersModule,
		Tax:                   taxModule,
		Team:                  team.NewModule(ctx, dbclient.Pool()),
		Templates:             templates.NewModule(ctx, dbclient.Pool()),
		TimeOff:               timeoff.NewModule(ctx, dbclient.Pool()),
		Timesheets:            timesheets.NewModule(ctx, dbclient.Pool()),
		Training:              training.NewModule(ctx, dbclient.Pool()),
		User: func() *user.Module {
			m := user.NewModule(ctx, dbclient.Pool(), platformConfig.GlobalConfig.GoogleOAuthClientID)
			m.WithBusinessProvisioner(hrCoreModule)
			cfg := platformConfig.GlobalConfig
			if cfg.WorkOSClientID != "" {
				m.WithWorkOS(cfg.WorkOSAPIKey, cfg.WorkOSClientID, cfg.WorkOSRedirectURI)
			}
			return m
		}(),
		Variants:    variants.NewModule(ctx, dbclient.Pool()),
		Voxhire:     voxhire.NewModule(ctx, dbclient.Pool()),
		Warehouses:  warehousesModule,
		Webhooks:    webhooks.NewModule(ctx, dbclient.Pool()),
		WebsiteLead: websitelead.NewModule(ctx, dbclient.Pool()),
		Workspaces:  workspacesModule,
	}
}

func setupRoutes(ctx context.Context, r *chi.Mux) {
	modules := initializeModules(ctx)

	// Start the websocket bus with topic-level authorization backed by
	// workspace/project membership checks. The rules:
	//
	//   user:<id>        → caller must be that user
	//   owner:<id>       → caller must own or belong to that owner account
	//   business:<id>    → caller must belong to that business
	//   workspace:<id>   → caller must belong to the workspace
	//   project:<id>     → caller must be a project or workspace member
	//   task:<id>        → caller must be able to access the task's project
	//
	// Unknown topic schemes are denied by default.
	workspacesSvc := modules.Workspaces.Service()
	teamSvc := modules.Team.Service()
	// Employment-driven access revocation (#834): when hr_core terminates an
	// employee, team (the entitlement owner) drops that user's permission rows
	// for the business. team's daily reconciliation job is the durability net.
	modules.HRCore.SetTerminationRevoker(teamSvc)
	hubResolver := team.NewHubProfileResolver(modules.Team.Service())
	wsbus.SetDefault(ctx, wsbus.New(func(ctx context.Context, userID int64, topic string) bool {
		scope, id, ok := wsbus.ParseTopic(topic)
		if !ok {
			return false
		}
		switch scope {
		case "user":
			return id == userID
		case "owner":
			return teamSvc.CanAccessOwner(ctx, userID, id)
		case "business":
			// Workspace-removal topic: project lifecycle events fan out here.
			// The authorizer only gets a user id, so membership resolves via
			// the hub profile: a resolvable profile means the caller belongs
			// to that business.
			_, profileErr := hubResolver.GetProfile(ctx, userID, id)
			return profileErr == nil
		case "workspace":
			return workspacesSvc.IsWorkspaceMember(ctx, userID, id)
		case "project":
			return workspacesSvc.IsProjectMember(ctx, userID, id)
		case "task":
			return workspacesSvc.IsTaskMember(ctx, userID, id)
		default:
			return false
		}
	}))

	// Fan realtime events across replicas when Redis is available; otherwise the
	// hub stays in-process, which is correct for a single replica.
	if platformConfig.IsRedisEnabled() && cache.RedisAvailable() {
		wsbus.EnableRedisBackplane()
	} else if platformConfig.IsRedisEnabled() {
		logger.Ctx(ctx).Warn("Redis backplane disabled; using in-process websocket fan-out")
	}

	accessPolicy := http.NewRouteAccessPolicy()
	accessPolicy.RegisterAnonymousRoutes("/health")
	accessPolicy.RegisterAnonymousRoutes("/version")
	accessPolicy.RegisterAnonymousRoutes("/license/state")
	accessPolicy.RegisterAnonymousRoutes("/runtime-config")
	accessPolicy.RegisterAnonymousRoutes("/.well-known/rtylr.json")
	accessPolicy.RegisterAnonymousRoutes("/flow/ws")
	accessPolicy.RegisterAnonymousRoutes("/invalidate-cache")

	provisionerResolver := authPkg.NewStaticPlatformProvisionerResolver(
		nil,
	)
	modules.User.ProvisionerResolver = provisionerResolver
	modules.User.ProfileResolver = team.NewHubProfileResolver(modules.Team.Service())
	modules.User.WithProvisioningStore(teamSvc)
	modules.User.WithInviteAccepter(modules.Team.Service())
	// A business-wide Owner/Admin/Operator has no path through Flow's own
	// project/workspace membership tables, so RequireProjectRole/
	// RequireTaskRole/RequireWorkspaceRole consult this resolver first and
	// admit them unconditionally before falling back to Flow's own roles.
	modules.Workspaces.Service().SetHubProfileResolver(team.NewHubProfileResolver(modules.Team.Service()))
	// Dept-manager visibility proxy (M3): Flow asks HR who reports to a caller
	// and grants them member-equivalent access to those people's projects. Until
	// this is wired the proxy is inert, so a worker or test without it sees
	// exactly the pre-proxy behaviour.
	modules.Workspaces.Service().SetOrgchartResolver(modules.Staff.OrgchartResolver())
	modules.HRDashboard.WithStatsLoader(func(ctx context.Context, businessID int64) (hrdashboardstats.DashboardStats, error) {
		totalStaff, onShiftToday, pendingTimeOff, payrollDraft, err := modules.Staff.LoadDashboardStats(ctx, businessID)
		if err != nil {
			return hrdashboardstats.DashboardStats{}, err
		}
		return hrdashboardstats.DashboardStats{
			TotalStaff:     totalStaff,
			OnShiftToday:   onShiftToday,
			PendingTimeOff: pendingTimeOff,
			PayrollDraft:   payrollDraft,
		}, nil
	})
	modules.Business.WithUserService(modules.User.UserService)
	modules.Apps.SetAddonProvider(addons.NewHubAddonProvider(modules.Addons.GetService()))
	modules.APIKeys.SetWebhooksProvider(webhooks.NewHubWebhooksProvider(modules.Webhooks.GetService()))

	// Wire personal workspace creation into user registration and login (idempotent).
	ensurePersonalWorkspace := func(ctx context.Context, userID int64) {
		modules.Workspaces.CreatePersonalWorkspace(ctx, userID)
	}
	user.OnUserCreated = ensurePersonalWorkspace
	user.OnUserLogin = ensurePersonalWorkspace

	http.AutoRegisterAnonymousRoutes(ctx, modules, accessPolicy)
	http.AutoConfigureAccessPolicy(ctx, modules, accessPolicy)

	tokenResolver := authPkg.NewStaticTokenValidatorResolver(
		modules.User.GetTokenValidator(),
		map[authPkg.ProductPlatform]authPkg.TokenValidator{
			authPkg.ProductPlatformRtylr:  modules.User.GetTokenValidator(),
			authPkg.ProductPlatformVoxire: modules.User.GetTokenValidator(),
		},
	)

	// Self-host license validator. A no-op in hosted mode (always valid, no
	// network); in self-host mode it validates the key against the license
	// authority with a grace window. Built once and shared by /license/state,
	// /runtime-config, and the enforcement middleware below.
	licenseValidator := license.NewCachedValidator(platformConfig.GlobalConfig)

	middlewares := []http.MiddlewareFunc{
		http.SecurityMiddleware(),
		jwtutils.JWTMiddleware(),
		http.RequestLoggerMiddleware(ctx),
		http.GetAuthMiddlewareFunc(accessPolicy, tokenResolver, teamSvc),
		authPkg.DeviceMiddleware(),
		// Keep subscription state authoritative for every authenticated tenant
		// request. The middleware leaves billing recovery and console routes open.
		http.SubscriptionEnforcementMiddleware(dbclient.Pool(), accessPolicy),
		// After auth so idempotency keys are scoped per authenticated user.
		http.IdempotencyMiddleware,
		// AFTER auth: gate protected API routes when a self-host license is
		// invalid. Strict no-op in hosted mode (never consults the validator).
		http.LicenseEnforcementMiddleware(licenseValidator, accessPolicy),
	}

	http.ApplyMiddlewares(r, middlewares)

	http.RegisterCacheRoutes(ctx, r)
	http.RegisterStatusRoutes(ctx, r)
	http.RegisterVersionRoutes(r)

	// /license/state lets frontends render an admin-only notice. Prime the
	// cache in the background so the first protected request isn't blocked on
	// the authority round-trip.
	http.RegisterLicenseStateRoutes(r, licenseValidator)
	http.RegisterRuntimeConfigRoutes(r, licenseValidator)
	if license.IsSelfHosted(platformConfig.GlobalConfig) {
		go licenseValidator.Refresh(ctx)
	}

	// oneKit contract services: every module's generated service servers
	// register on one shared mux; the generated code owns every contract path.
	// Runs before RegisterRoutes so module escape hatches (rate-limited or
	// special routes) can reference the mux captured in RegisterAPI.
	oneKitMux := stdhttp.NewServeMux()
	http.AutoRegisterAPIs(ctx, modules, oneKitMux)
	http.MountOneKit(r, oneKitMux)

	http.AutoRegisterRoutes(ctx, modules, r)
}

type apiServer struct{}

func (a apiServer) Setup(ctx context.Context, r *chi.Mux) {
	setupRoutes(ctx, r)
}

func startAPIServer(ctx context.Context) error {
	server := apiServer{}
	return http.Start(ctx, server, platformConfig.ListenPort("9999"))
}

func recoverAndLog(ctx context.Context) {
	if r := recover(); r != nil {
		logger.Ctx(ctx).Error("Recovered panic", logger.Any("error", r))
	}
}

type ServerPanicError struct {
	panic interface{}
}

func (e *ServerPanicError) Error() string {
	return "server panic occurred"
}

func (e *ServerPanicError) Unwrap() error {
	return nil
}
