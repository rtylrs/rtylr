package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
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
	"github.com/voxire/onion/internal/modules/shifts"
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
	platformConfig "github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	emailPkg "github.com/voxire/onion/internal/platform/email"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/observability"
	"github.com/voxire/onion/internal/platform/server/http"
	sheetsPkg "github.com/voxire/onion/internal/platform/sheets"
	"github.com/voxire/onion/internal/platform/utils/jwtutils"
	"github.com/voxire/onion/internal/platform/wsbus"
)

// loadDotEnv reads a .env file from the working directory and sets any key=value
// pairs as environment variables. Missing or unreadable .env files are silently
// ignored — in production, variables are injected by the container runtime.
func loadDotEnv() {
	f, err := os.Open(".env")
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		if key == "" {
			continue
		}
		if err := os.Setenv(key, val); err != nil {
			logger.Global().Warn("could not set env var",
				logger.String("key", key),
				logger.Error(err),
			)
		}
	}
}

func main() {
	loadDotEnv()
	logger.SetupLogger()

	if _, err := platformConfig.LoadAndValidate(); err != nil {
		logger.Global().Fatal("Configuration validation failed", logger.Error(err))
	}

	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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
		case <-time.After(10 * time.Second):
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

	ctx, err = jwtutils.InitializeJWT(ctx)
	if err != nil {
		return fmt.Errorf("jwt initialization failed: %w", err)
	}

	emailPkg.Initialize()
	sheetsPkg.Initialize(ctx)

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
	Shifts                *shifts.Module
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
	customersModule := customers.NewModule(ctx, dbclient.Pool())
	inventoryModule := inventory.NewModule(ctx, dbclient.Pool())
	workspacesModule := workspaces.NewModule(ctx, dbclient.Pool())
	expensesModule := expenses.NewModule(ctx, dbclient.Pool())
	posOrdersModule := posorders.NewModule(ctx, dbclient.Pool())
	suppliersModule := suppliers.NewModule(ctx, dbclient.Pool())

	return &apiModules{
		Accounts:              accounts.NewModule(ctx, dbclient.Pool()),
		Activity:              activity.NewModule(ctx, dbclient.Pool()),
		Addons:                addons.NewModule(ctx, dbclient.Pool()),
		APIKeys:               apikeys.NewModule(ctx, dbclient.Pool()),
		Apps:                  apps.NewModule(ctx, dbclient.Pool()),
		Automations:           automations.NewModule(ctx, dbclient.Pool()),
		BankAccounts:          bankaccounts.NewModule(ctx, dbclient.Pool()),
		Benefits:              benefits.NewModule(ctx, dbclient.Pool()),
		Billing:               billing.NewModule(ctx, dbclient.Pool()),
		Subscriptions:         subscriptions.NewModule(ctx, dbclient.Pool()),
		Bills:                 bills.NewModule(ctx, dbclient.Pool()),
		BookingAppointments:   bookingappointments.NewModule(ctx, dbclient.Pool()),
		BookingCalendar:       bookingcalendar.NewModule(ctx, dbclient.Pool()),
		BookingDashboard:      bookingdashboard.NewModule(ctx, dbclient.Pool()),
		BookingReminders:      bookingreminders.NewModule(ctx, dbclient.Pool()),
		BookingResources:      bookingresources.NewModule(ctx, dbclient.Pool()),
		BookingServices:       bookingservices.NewModule(ctx, dbclient.Pool()),
		BookingWaitlist:       bookingwaitlist.NewModule(ctx, dbclient.Pool()),
		Brands:                brands.NewModule(ctx, dbclient.Pool()),
		Business:              business.NewModule(ctx, dbclient.Pool()),
		Campaigns:             campaigns.NewModule(ctx, dbclient.Pool()),
		Careers:               careers.NewModule(ctx, dbclient.Pool()),
		Categories:            categories.NewModule(ctx, dbclient.Pool()),
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
		Echo:                  echo.NewModule(ctx, memoryModule, aiQuotaModule, aiQueryModule, workspacesModule, customersModule, inventoryModule, expensesModule, posOrdersModule, hrCoreModule, suppliersModule),
		EmailTest:             emailtest.NewModule(ctx, dbclient.Pool()),
		Expenses:              expensesModule,
		Feedback:              feedback.NewModule(ctx, dbclient.Pool()),
		FinanceBankFeeds:      financebankfeeds.NewModule(ctx, dbclient.Pool()),
		FinanceReconciliation: financereconciliation.NewModule(ctx, dbclient.Pool()),
		FinanceReports:        financereports.NewModule(ctx, dbclient.Pool()),
		FlowInsights:          flowinsights.NewModule(ctx, dbclient.Pool()),
		HRAttendance:          hrattendance.NewModule(ctx, dbclient.Pool()),
		HRCore:                hrCoreModule,
		HRDashboard:           hrdashboard.NewModule(ctx, dbclient.Pool()),
		HRDocuments:           hrdocuments.NewModule(ctx, dbclient.Pool()),
		HREquipment:           hrequipment.NewModule(ctx, dbclient.Pool(), hrCoreModule),
		HRLeave:               hrleave.NewModule(ctx, dbclient.Pool(), hrCoreModule),
		HROnboarding:          hronboarding.NewModule(ctx, dbclient.Pool()),
		HROvertime:            hrovertime.NewModule(ctx, dbclient.Pool(), hrCoreModule),
		HRPayroll:             hrpayroll.NewModule(ctx, dbclient.Pool(), hrCoreModule),
		HRPerformance:         hrperformance.NewModule(ctx, dbclient.Pool()),
		HRReports:             hrreports.NewModule(ctx, dbclient.Pool()),
		HRSetup:               hrsetup.NewModule(ctx, dbclient.Pool()),
		HRShifts:              hrshifts.NewModule(ctx, dbclient.Pool()),
		HRTime:                hrtime.NewModule(ctx, dbclient.Pool(), hrCoreModule),
		Insights:              insights.NewModule(ctx, dbclient.Pool()),
		Integrations:          integrations.NewModule(ctx, dbclient.Pool()),
		Inventory:             inventoryModule,
		Invoices:              invoices.NewModule(ctx, dbclient.Pool()),
		Journal:               journal.NewModule(ctx, dbclient.Pool()),
		KPIs:                  kpis.NewModule(ctx, dbclient.Pool()),
		Loyalty:               loyalty.NewModule(ctx, dbclient.Pool()),
		Memory:                memoryModule,
		Notifications:         notifications.NewModule(ctx, dbclient.Pool()),
		OAuth:                 oauth.NewModule(ctx, dbclient.Pool(), platformConfig.GlobalConfig.AuthAppURL),
		OneOnOnes:             oneonones.NewModule(ctx, dbclient.Pool()),
		Payroll:               payroll.NewModule(ctx, dbclient.Pool()),
		PosAudit:              posaudit.NewModule(ctx, dbclient.Pool()),
		PosCashMovements:      poscashmovements.NewModule(ctx, dbclient.Pool()),
		PosCombos:             poscombos.NewModule(ctx, dbclient.Pool()),
		PosCustomers:          poscustomers.NewModule(ctx, dbclient.Pool()),
		PosDaySummary:         posdaysummary.NewModule(ctx, dbclient.Pool()),
		PosFloors:             posfloors.NewModule(ctx, dbclient.Pool()),
		PosGiftCards:          posgiftcards.NewModule(ctx, dbclient.Pool()),
		PosKDS:                poskds.NewModule(ctx, dbclient.Pool()),
		PosLocations:          poslocations.NewModule(ctx, dbclient.Pool()),
		PosLoyalty:            posloyalty.NewModule(ctx, dbclient.Pool()),
		PosManager:            posmanager.NewModule(ctx, dbclient.Pool()),
		PosMemberLocations:    posmemberlocations.NewModule(ctx, dbclient.Pool()),
		PosMenu:               posmenu.NewModule(ctx, dbclient.Pool()),
		PosModifiers:          posmodifiers.NewModule(ctx, dbclient.Pool()),
		PosOrders:             posOrdersModule,
		PosReceipts:           posreceipts.NewModule(ctx, dbclient.Pool()),
		PosSessions:           possessions.NewModule(ctx, dbclient.Pool()),
		PosSettings:           possettings.NewModule(ctx, dbclient.Pool()),
		PosStockMovements:     posstockmovements.NewModule(ctx, dbclient.Pool()),
		PosTables:             postables.NewModule(ctx, dbclient.Pool()),
		PosTerminals:          posterminals.NewModule(ctx, dbclient.Pool()),
		PosUsers:              posusers.NewModule(ctx, dbclient.Pool()),
		PriceLists:            pricelists.NewModule(ctx, dbclient.Pool()),
		Products:              products.NewModule(ctx, dbclient.Pool()),
		PurchaseOrders:        purchaseorders.NewModule(ctx, dbclient.Pool()),
		Quotes:                quotes.NewModule(ctx, dbclient.Pool()),
		RecurringInvoices:     recurringinvoices.NewModule(ctx, dbclient.Pool()),
		Reviews:               reviews.NewModule(ctx, dbclient.Pool()),
		Roles:                 roles.NewModule(ctx, dbclient.Pool()),
		Security:              security.NewModule(ctx, dbclient.Pool()),
		Segments:              segments.NewModule(ctx, dbclient.Pool()),
		Shifts:                shifts.NewModule(ctx, dbclient.Pool()),
		Staff:                 staff.NewModule(ctx, dbclient.Pool()),
		StockTakes:            stocktakes.NewModule(ctx, dbclient.Pool()),
		StockTransfers:        stocktransfers.NewModule(ctx, dbclient.Pool()),
		StoreCategories:       storecategories.NewModule(ctx, dbclient.Pool()),
		StoreCheckout:         storecheckout.NewModule(ctx, dbclient.Pool()),
		Storefronts:           storefronts.NewModule(ctx, dbclient.Pool()),
		StoreOrders:           storeorders.NewModule(ctx, dbclient.Pool()),
		StorePayments:         storepayments.NewModule(ctx, dbclient.Pool()),
		StoreProducts:         storeproducts.NewModule(ctx, dbclient.Pool()),
		StoreShipping:         storeshipping.NewModule(ctx, dbclient.Pool()),
		StoreThemes:           storethemes.NewModule(ctx, dbclient.Pool()),
		SupportCanned:         supportcanned.NewModule(ctx, dbclient.Pool()),
		SupportCustomers:      supportcustomers.NewModule(ctx, dbclient.Pool()),
		SupportInbox:          supportinbox.NewModule(ctx, dbclient.Pool()),
		SupportKnowledge:      supportknowledge.NewModule(ctx, dbclient.Pool()),
		SupportSettings:       supportsettings.NewModule(ctx, dbclient.Pool()),
		SupportTickets:        supporttickets.NewModule(ctx, dbclient.Pool()),
		Suppliers:             suppliersModule,
		Tax:                   tax.NewModule(ctx, dbclient.Pool()),
		Team:                  team.NewModule(ctx, dbclient.Pool()),
		Templates:             templates.NewModule(ctx, dbclient.Pool()),
		TimeOff:               timeoff.NewModule(ctx, dbclient.Pool()),
		Timesheets:            timesheets.NewModule(ctx, dbclient.Pool()),
		Training:              training.NewModule(ctx, dbclient.Pool()),
		User: func() *user.Module {
			m := user.NewModule(ctx, dbclient.Pool(), platformConfig.GlobalConfig.GoogleOAuthClientID)
			cfg := platformConfig.GlobalConfig
			if cfg.WorkOSClientID != "" {
				m.WithWorkOS(cfg.WorkOSAPIKey, cfg.WorkOSClientID, cfg.WorkOSRedirectURI)
			}
			return m
		}(),
		Variants:    variants.NewModule(ctx, dbclient.Pool()),
		Voxhire:     voxhire.NewModule(ctx, dbclient.Pool()),
		Warehouses:  warehouses.NewModule(ctx, dbclient.Pool()),
		Webhooks:    webhooks.NewModule(ctx, dbclient.Pool()),
		WebsiteLead: websitelead.NewModule(ctx, dbclient.Pool()),
		Workspaces:  workspacesModule,
	}
}

func setupRoutes(ctx context.Context, r *chi.Mux) {
	modules := initializeModules(ctx)
	modules.HRLeave.RegisterScheduledJobs(ctx)
	modules.HROvertime.RegisterScheduledJobs(ctx)
	modules.HRPayroll.RegisterScheduledJobs(ctx)

	// Start the websocket bus with topic-level authorization backed by
	// workspace/project membership checks. The rules:
	//
	//   user:<id>        → caller must be that user
	//   owner:<id>       → caller must own or belong to that owner account
	//   workspace:<id>   → caller must belong to the workspace
	//   project:<id>     → caller must be a project or workspace member
	//   task:<id>        → caller must be able to access the task's project
	//
	// Unknown topic schemes are denied by default.
	workspacesSvc := modules.Workspaces.Service()
	teamSvc := modules.Team.Service()
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
	if platformConfig.IsRedisEnabled() {
		wsbus.EnableRedisBackplane()
	}

	accessPolicy := http.NewRouteAccessPolicy()
	accessPolicy.RegisterAnonymousRoutes("/health")
	accessPolicy.RegisterAnonymousRoutes("/version")
	accessPolicy.RegisterAnonymousRoutes("/flow/ws")

	provisionerResolver := authPkg.NewStaticPlatformProvisionerResolver(
		nil,
	)
	modules.User.ProvisionerResolver = provisionerResolver

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

	middlewares := []http.MiddlewareFunc{
		http.SecurityMiddleware(),
		jwtutils.JWTMiddleware(),
		http.RequestLoggerMiddleware(ctx),
		http.GetAuthMiddlewareFunc(accessPolicy, tokenResolver),
		authPkg.DeviceMiddleware(),
	}

	http.ApplyMiddlewares(r, middlewares)

	http.RegisterCacheRoutes(ctx, r)
	http.RegisterStatusRoutes(ctx, r)
	http.RegisterVersionRoutes(r)

	http.AutoRegisterRoutes(ctx, modules, r)
}

type apiServer struct{}

func (a apiServer) Setup(ctx context.Context, r *chi.Mux) {
	setupRoutes(ctx, r)
}

func startAPIServer(ctx context.Context) error {
	server := apiServer{}
	return http.Start(ctx, server, "9999")
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
