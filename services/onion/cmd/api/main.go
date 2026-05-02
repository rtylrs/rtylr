package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/accounts"
	"github.com/voxire/onion/internal/modules/activity"
	"github.com/voxire/onion/internal/modules/addons"
	apikeys "github.com/voxire/onion/internal/modules/api_keys"
	"github.com/voxire/onion/internal/modules/apps"
	"github.com/voxire/onion/internal/modules/automations"
	bankaccounts "github.com/voxire/onion/internal/modules/bank_accounts"
	"github.com/voxire/onion/internal/modules/benefits"
	"github.com/voxire/onion/internal/modules/billing"
	"github.com/voxire/onion/internal/modules/subscriptions"
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
	"github.com/voxire/onion/internal/modules/campaigns"
	"github.com/voxire/onion/internal/modules/careers"
	"github.com/voxire/onion/internal/modules/categories"
	"github.com/voxire/onion/internal/modules/config"
	crmsettings "github.com/voxire/onion/internal/modules/crm_settings"
	crmsales "github.com/voxire/onion/internal/modules/crm_sales"
	"github.com/voxire/onion/internal/modules/customers"
	"github.com/voxire/onion/internal/modules/cycles"
	"github.com/voxire/onion/internal/modules/departments"
	"github.com/voxire/onion/internal/modules/documents"
	"github.com/voxire/onion/internal/modules/echo"
	"github.com/voxire/onion/internal/modules/expenses"
	"github.com/voxire/onion/internal/modules/feedback"
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
	financebankfeeds "github.com/voxire/onion/internal/modules/finance_bank_feeds"
	financereconciliation "github.com/voxire/onion/internal/modules/finance_reconciliation"
	financereports "github.com/voxire/onion/internal/modules/finance_reports"
	flowinsights "github.com/voxire/onion/internal/modules/flow_insights"
	"github.com/voxire/onion/internal/modules/insights"
	"github.com/voxire/onion/internal/modules/integrations"
	"github.com/voxire/onion/internal/modules/inventory"
	"github.com/voxire/onion/internal/modules/invoices"
	"github.com/voxire/onion/internal/modules/journal"
	"github.com/voxire/onion/internal/modules/kpis"
	"github.com/voxire/onion/internal/modules/loyalty"
	"github.com/voxire/onion/internal/modules/memory"
	"github.com/voxire/onion/internal/modules/milestones"
	"github.com/voxire/onion/internal/modules/notifications"
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
	"github.com/voxire/onion/internal/modules/views"
	"github.com/voxire/onion/internal/modules/warehouses"
	"github.com/voxire/onion/internal/modules/webhooks"
	"github.com/voxire/onion/internal/modules/websitelead"
	"github.com/voxire/onion/internal/modules/workspaces"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/cache"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	emailPkg "github.com/voxire/onion/internal/platform/email"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/observability"
	"github.com/voxire/onion/internal/platform/server/http"
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
	runtime.GOMAXPROCS(runtime.NumCPU())

	loadDotEnv()

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
	Accounts             *accounts.Module
	Activity             *activity.Module
	Addons               *addons.Module
	APIKeys              *apikeys.Module
	Apps                 *apps.Module
	Automations          *automations.Module
	BankAccounts         *bankaccounts.Module
	Benefits             *benefits.Module
	Billing              *billing.Module
	Subscriptions        *subscriptions.Module
	Bills                *bills.Module
	BookingAppointments  *bookingappointments.Module
	BookingCalendar      *bookingcalendar.Module
	BookingDashboard     *bookingdashboard.Module
	BookingReminders     *bookingreminders.Module
	BookingResources     *bookingresources.Module
	BookingServices      *bookingservices.Module
	BookingWaitlist      *bookingwaitlist.Module
	Brands               *brands.Module
	Business             *business.Module
	Campaigns            *campaigns.Module
	Careers              *careers.Module
	Categories           *categories.Module
	Config               *config.Module
	CRMSettings          *crmsettings.Module
	CRMSales             *crmsales.Module
	Customers            *customers.Module
	Cycles               *cycles.Module
	Departments          *departments.Module
	Documents            *documents.Module
	Echo                 *echo.Module
	Expenses             *expenses.Module
	Feedback             *feedback.Module
	FinanceBankFeeds      *financebankfeeds.Module
	FinanceReconciliation *financereconciliation.Module
	FinanceReports       *financereports.Module
	FlowInsights         *flowinsights.Module
	HRAttendance         *hrattendance.Module
	HRCore               *hrcore.Module
	HRDashboard          *hrdashboard.Module
	HRDocuments          *hrdocuments.Module
	HREquipment          *hrequipment.Module
	HRLeave              *hrleave.Module
	HROnboarding         *hronboarding.Module
	HROvertime           *hrovertime.Module
	HRPayroll            *hrpayroll.Module
	HRPerformance        *hrperformance.Module
	HRReports            *hrreports.Module
	HRSetup              *hrsetup.Module
	HRShifts             *hrshifts.Module
	HRTime               *hrtime.Module
	Insights             *insights.Module
	Integrations         *integrations.Module
	Inventory            *inventory.Module
	Invoices             *invoices.Module
	Journal              *journal.Module
	KPIs                 *kpis.Module
	Loyalty              *loyalty.Module
	Memory               *memory.Module
	Milestones           *milestones.Module
	Notifications        *notifications.Module
	OneOnOnes            *oneonones.Module
	Payroll              *payroll.Module
	PosAudit             *posaudit.Module
	PosCashMovements     *poscashmovements.Module
	PosCombos            *poscombos.Module
	PosCustomers         *poscustomers.Module
	PosDaySummary        *posdaysummary.Module
	PosFloors            *posfloors.Module
	PosGiftCards         *posgiftcards.Module
	PosKDS               *poskds.Module
	PosLocations         *poslocations.Module
	PosLoyalty           *posloyalty.Module
	PosManager           *posmanager.Module
	PosMemberLocations   *posmemberlocations.Module
	PosMenu              *posmenu.Module
	PosModifiers         *posmodifiers.Module
	PosOrders            *posorders.Module
	PosReceipts          *posreceipts.Module
	PosSessions          *possessions.Module
	PosStockMovements    *posstockmovements.Module
	PosTables            *postables.Module
	PosTerminals         *posterminals.Module
	PosUsers             *posusers.Module
	PriceLists           *pricelists.Module
	Products             *products.Module
	PurchaseOrders       *purchaseorders.Module
	Quotes               *quotes.Module
	RecurringInvoices    *recurringinvoices.Module
	Reviews              *reviews.Module
	Roles                *roles.Module
	Security             *security.Module
	Segments             *segments.Module
	Shifts               *shifts.Module
	Staff                *staff.Module
	StockTakes           *stocktakes.Module
	StockTransfers       *stocktransfers.Module
	StoreCategories      *storecategories.Module
	StoreCheckout        *storecheckout.Module
	Storefronts          *storefronts.Module
	StoreOrders          *storeorders.Module
	StorePayments        *storepayments.Module
	StoreProducts        *storeproducts.Module
	StoreShipping        *storeshipping.Module
	StoreThemes          *storethemes.Module
	SupportCanned        *supportcanned.Module
	SupportCustomers     *supportcustomers.Module
	SupportInbox         *supportinbox.Module
	SupportKnowledge     *supportknowledge.Module
	SupportSettings      *supportsettings.Module
	SupportTickets       *supporttickets.Module
	Suppliers            *suppliers.Module
	Tax                  *tax.Module
	Team                 *team.Module
	Templates            *templates.Module
	TimeOff              *timeoff.Module
	Timesheets           *timesheets.Module
	Training             *training.Module
	User                 *user.Module
	Variants             *variants.Module
	Views                *views.Module
	Warehouses           *warehouses.Module
	Webhooks             *webhooks.Module
	WebsiteLead          *websitelead.Module
	Workspaces           *workspaces.Module
}

func initializeModules(ctx context.Context) *apiModules {
	memoryModule := memory.NewModule(ctx)
	hrCoreModule := hrcore.NewModule(ctx)

	return &apiModules{
		Accounts:             accounts.NewModule(ctx),
		Activity:             activity.NewModule(ctx),
		Addons:               addons.NewModule(ctx),
		APIKeys:              apikeys.NewModule(ctx),
		Apps:                 apps.NewModule(ctx),
		Automations:          automations.NewModule(ctx),
		BankAccounts:         bankaccounts.NewModule(ctx),
		Benefits:             benefits.NewModule(ctx),
		Billing:              billing.NewModule(ctx),
		Subscriptions:        subscriptions.NewModule(ctx),
		Bills:                bills.NewModule(ctx),
		BookingAppointments:  bookingappointments.NewModule(ctx),
		BookingCalendar:      bookingcalendar.NewModule(ctx),
		BookingDashboard:     bookingdashboard.NewModule(ctx),
		BookingReminders:     bookingreminders.NewModule(ctx),
		BookingResources:     bookingresources.NewModule(ctx),
		BookingServices:      bookingservices.NewModule(ctx),
		BookingWaitlist:      bookingwaitlist.NewModule(ctx),
		Brands:               brands.NewModule(ctx),
		Business:             business.NewModule(ctx),
		Campaigns:            campaigns.NewModule(ctx),
		Careers:              careers.NewModule(ctx),
		Categories:           categories.NewModule(ctx),
		Config:               config.NewModule(ctx),
		CRMSettings:          crmsettings.NewModule(ctx),
		CRMSales:             crmsales.NewModule(ctx),
		Customers:            customers.NewModule(ctx),
		Cycles:               cycles.NewModule(ctx),
		Departments:          departments.NewModule(ctx),
		Documents:            documents.NewModule(ctx),
		Echo:                 echo.NewModule(ctx, memoryModule),
		Expenses:             expenses.NewModule(ctx),
		Feedback:             feedback.NewModule(ctx),
		FinanceBankFeeds:      financebankfeeds.NewModule(ctx),
		FinanceReconciliation: financereconciliation.NewModule(ctx),
		FinanceReports:       financereports.NewModule(ctx),
		FlowInsights:         flowinsights.NewModule(ctx),
		HRAttendance:         hrattendance.NewModule(ctx),
		HRCore:               hrCoreModule,
		HRDashboard:          hrdashboard.NewModule(ctx),
		HRDocuments:          hrdocuments.NewModule(ctx),
		HREquipment:          hrequipment.NewModule(ctx, hrCoreModule),
		HRLeave:              hrleave.NewModule(ctx, hrCoreModule),
		HROnboarding:         hronboarding.NewModule(ctx),
		HROvertime:           hrovertime.NewModule(ctx, hrCoreModule),
		HRPayroll:            hrpayroll.NewModule(ctx, hrCoreModule),
		HRPerformance:        hrperformance.NewModule(ctx),
		HRReports:            hrreports.NewModule(ctx),
		HRSetup:              hrsetup.NewModule(ctx),
		HRShifts:             hrshifts.NewModule(ctx),
		HRTime:               hrtime.NewModule(ctx, hrCoreModule),
		Insights:             insights.NewModule(ctx),
		Integrations:         integrations.NewModule(ctx),
		Inventory:            inventory.NewModule(ctx),
		Invoices:             invoices.NewModule(ctx),
		Journal:              journal.NewModule(ctx),
		KPIs:                 kpis.NewModule(ctx),
		Loyalty:              loyalty.NewModule(ctx),
		Memory:               memoryModule,
		Milestones:           milestones.NewModule(ctx),
		Notifications:        notifications.NewModule(ctx),
		OneOnOnes:            oneonones.NewModule(ctx),
		Payroll:              payroll.NewModule(ctx),
		PosAudit:             posaudit.NewModule(ctx),
		PosCashMovements:     poscashmovements.NewModule(ctx),
		PosCombos:            poscombos.NewModule(ctx),
		PosCustomers:         poscustomers.NewModule(ctx),
		PosDaySummary:        posdaysummary.NewModule(ctx),
		PosFloors:            posfloors.NewModule(ctx),
		PosGiftCards:         posgiftcards.NewModule(ctx),
		PosKDS:               poskds.NewModule(ctx),
		PosLocations:         poslocations.NewModule(ctx),
		PosLoyalty:           posloyalty.NewModule(ctx),
		PosManager:           posmanager.NewModule(ctx),
		PosMemberLocations:   posmemberlocations.NewModule(ctx),
		PosMenu:              posmenu.NewModule(ctx),
		PosModifiers:         posmodifiers.NewModule(ctx),
		PosOrders:            posorders.NewModule(ctx),
		PosReceipts:          posreceipts.NewModule(ctx),
		PosSessions:          possessions.NewModule(ctx),
		PosStockMovements:    posstockmovements.NewModule(ctx),
		PosTables:            postables.NewModule(ctx),
		PosTerminals:         posterminals.NewModule(ctx),
		PosUsers:             posusers.NewModule(ctx),
		PriceLists:           pricelists.NewModule(ctx),
		Products:             products.NewModule(ctx),
		PurchaseOrders:       purchaseorders.NewModule(ctx),
		Quotes:               quotes.NewModule(ctx),
		RecurringInvoices:    recurringinvoices.NewModule(ctx),
		Reviews:              reviews.NewModule(ctx),
		Roles:                roles.NewModule(ctx),
		Security:             security.NewModule(ctx),
		Segments:             segments.NewModule(ctx),
		Shifts:               shifts.NewModule(ctx),
		Staff:                staff.NewModule(ctx),
		StockTakes:           stocktakes.NewModule(ctx),
		StockTransfers:       stocktransfers.NewModule(ctx),
		StoreCategories:      storecategories.NewModule(ctx),
		StoreCheckout:        storecheckout.NewModule(ctx),
		Storefronts:          storefronts.NewModule(ctx),
		StoreOrders:          storeorders.NewModule(ctx),
		StorePayments:        storepayments.NewModule(ctx),
		StoreProducts:        storeproducts.NewModule(ctx),
		StoreShipping:        storeshipping.NewModule(ctx),
		StoreThemes:          storethemes.NewModule(ctx),
		SupportCanned:        supportcanned.NewModule(ctx),
		SupportCustomers:     supportcustomers.NewModule(ctx),
		SupportInbox:         supportinbox.NewModule(ctx),
		SupportKnowledge:     supportknowledge.NewModule(ctx),
		SupportSettings:      supportsettings.NewModule(ctx),
		SupportTickets:       supporttickets.NewModule(ctx),
		Suppliers:            suppliers.NewModule(ctx),
		Tax:                  tax.NewModule(ctx),
		Team:                 team.NewModule(ctx),
		Templates:            templates.NewModule(ctx),
		TimeOff:              timeoff.NewModule(ctx),
		Timesheets:           timesheets.NewModule(ctx),
		Training:             training.NewModule(ctx),
		User:                 user.NewModule(ctx),
		Variants:             variants.NewModule(ctx),
		Views:                views.NewModule(ctx),
		Warehouses:           warehouses.NewModule(ctx),
		Webhooks:             webhooks.NewModule(ctx),
		WebsiteLead:          websitelead.NewModule(ctx),
		Workspaces:           workspaces.NewModule(ctx),
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
	//   workspace:<id>   → caller must belong to the workspace
	//   project:<id>     → caller must be a project or workspace member
	//   task:<id>        → (MVP) allowed; tightening requires a task→
	//                      project lookup on every subscribe, which we
	//                      skip until task-detail subscriptions land
	//
	// Unknown topic schemes are denied by default.
	workspacesSvc := modules.Workspaces.Service()
	wsbus.SetDefault(ctx, wsbus.New(func(ctx context.Context, userID int64, topic string) bool {
		scope, id, ok := wsbus.ParseTopic(topic)
		if !ok {
			return false
		}
		switch scope {
		case "user":
			return id == userID
		case "workspace":
			return workspacesSvc.IsWorkspaceMember(ctx, userID, id)
		case "project":
			return workspacesSvc.IsProjectMember(ctx, userID, id)
		case "task":
			return true
		default:
			return false
		}
	}))

	accessPolicy := http.NewRouteAccessPolicy()
	accessPolicy.RegisterAnonymousRoutes("/health")
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
		http.RequestLoggerMiddleware(ctx),
		http.GetAuthMiddlewareFunc(accessPolicy, tokenResolver),
		authPkg.DeviceMiddleware(),
	}

	http.ApplyMiddlewares(r, middlewares)

	http.RegisterCacheRoutes(ctx, r)
	http.RegisterStatusRoutes(ctx, r)

	http.AutoRegisterRoutes(ctx, modules, r)

	// Explicit registration guard: if AutoRegisterRoutes skipped HRDashboard via
	// reflection, this call will register it. If it was already registered, chi
	// will panic with "routing duplicate" — remove this line if that happens.
	if modules.HRDashboard != nil {
		modules.HRDashboard.RegisterRoutes(ctx, r)
	}
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
