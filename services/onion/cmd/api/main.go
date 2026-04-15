package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/voxire/onion/internal/modules/accounts"
	"github.com/voxire/onion/internal/modules/activity"
	"github.com/voxire/onion/internal/modules/addons"
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
	"github.com/voxire/onion/internal/modules/campaigns"
	"github.com/voxire/onion/internal/modules/careers"
	"github.com/voxire/onion/internal/modules/categories"
	"github.com/voxire/onion/internal/modules/config"
	crmsettings "github.com/voxire/onion/internal/modules/crm_settings"
	"github.com/voxire/onion/internal/modules/customers"
	"github.com/voxire/onion/internal/modules/cycles"
	"github.com/voxire/onion/internal/modules/departments"
	"github.com/voxire/onion/internal/modules/documents"
	"github.com/voxire/onion/internal/modules/echo"
	"github.com/voxire/onion/internal/modules/expenses"
	"github.com/voxire/onion/internal/modules/feedback"
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
	"github.com/voxire/onion/internal/modules/workspaces"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/cache"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	emailPkg "github.com/voxire/onion/internal/platform/email"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/server/http"
	"github.com/voxire/onion/internal/platform/utils/jwtutils"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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
	Accounts            *accounts.Module
	Activity            *activity.Module
	Addons              *addons.Module
	APIKeys             *apikeys.Module
	Apps                *apps.Module
	Automations         *automations.Module
	BankAccounts        *bankaccounts.Module
	Benefits            *benefits.Module
	Billing             *billing.Module
	Bills               *bills.Module
	BookingAppointments *bookingappointments.Module
	BookingCalendar     *bookingcalendar.Module
	BookingDashboard    *bookingdashboard.Module
	BookingReminders    *bookingreminders.Module
	BookingResources    *bookingresources.Module
	BookingServices     *bookingservices.Module
	BookingWaitlist     *bookingwaitlist.Module
	Brands              *brands.Module
	Business            *business.Module
	Campaigns           *campaigns.Module
	Careers             *careers.Module
	Categories          *categories.Module
	Config              *config.Module
	CRMSettings         *crmsettings.Module
	Customers           *customers.Module
	Cycles              *cycles.Module
	Departments         *departments.Module
	Documents           *documents.Module
	Echo                *echo.Module
	Expenses            *expenses.Module
	Feedback            *feedback.Module
	FlowInsights        *flowinsights.Module
	FinanceReports      *financereports.Module
	Insights            *insights.Module
	Integrations        *integrations.Module
	Inventory           *inventory.Module
	Invoices            *invoices.Module
	Journal             *journal.Module
	KPIs                *kpis.Module
	Loyalty             *loyalty.Module
	Memory              *memory.Module
	Milestones          *milestones.Module
	Notifications       *notifications.Module
	OneOnOnes           *oneonones.Module
	Payroll             *payroll.Module
	PosCombos           *poscombos.Module
	PosCustomers        *poscustomers.Module
	PosDaySummary       *posdaysummary.Module
	PosGiftCards        *posgiftcards.Module
	PosKDS              *poskds.Module
	PosLocations        *poslocations.Module
	PosManager          *posmanager.Module
	PosMenu             *posmenu.Module
	PosModifiers        *posmodifiers.Module
	PosOrders           *posorders.Module
	PosReceipts         *posreceipts.Module
	PosSessions         *possessions.Module
	PosTables           *postables.Module
	PosTerminals        *posterminals.Module
	PosFloors           *posfloors.Module
	PosAudit            *posaudit.Module
	PosCashMovements    *poscashmovements.Module
	PosLoyalty          *posloyalty.Module
	PosStockMovements   *posstockmovements.Module
	PosUsers            *posusers.Module
	PosMemberLocations  *posmemberlocations.Module
	Products            *products.Module
	PurchaseOrders      *purchaseorders.Module
	Quotes              *quotes.Module
	RecurringInvoices   *recurringinvoices.Module
	Reviews             *reviews.Module
	Roles               *roles.Module
	Security            *security.Module
	Segments            *segments.Module
	Shifts              *shifts.Module
	Staff               *staff.Module
	StockTakes          *stocktakes.Module
	StockTransfers      *stocktransfers.Module
	StoreCategories     *storecategories.Module
	StoreCheckout       *storecheckout.Module
	Storefronts         *storefronts.Module
	StoreOrders         *storeorders.Module
	StorePayments       *storepayments.Module
	StoreProducts       *storeproducts.Module
	StoreShipping       *storeshipping.Module
	StoreThemes         *storethemes.Module
	SupportCanned       *supportcanned.Module
	SupportCustomers    *supportcustomers.Module
	SupportInbox        *supportinbox.Module
	SupportKnowledge    *supportknowledge.Module
	SupportSettings     *supportsettings.Module
	SupportTickets      *supporttickets.Module
	Suppliers           *suppliers.Module
	Tax                 *tax.Module
	Team                *team.Module
	Templates           *templates.Module
	TimeOff             *timeoff.Module
	Timesheets          *timesheets.Module
	Training            *training.Module
	User                *user.Module
	Variants            *variants.Module
	Views               *views.Module
	Warehouses          *warehouses.Module
	Webhooks            *webhooks.Module
	Workspaces          *workspaces.Module
}

func initializeModules(ctx context.Context) *apiModules {
	memoryModule := memory.NewModule(ctx)

	return &apiModules{
		Accounts:            accounts.NewModule(ctx),
		Activity:            activity.NewModule(ctx),
		Addons:              addons.NewModule(ctx),
		APIKeys:             apikeys.NewModule(ctx),
		Apps:                apps.NewModule(ctx),
		Automations:         automations.NewModule(ctx),
		BankAccounts:        bankaccounts.NewModule(ctx),
		Benefits:            benefits.NewModule(ctx),
		Billing:             billing.NewModule(ctx),
		Bills:               bills.NewModule(ctx),
		BookingAppointments: bookingappointments.NewModule(ctx),
		BookingCalendar:     bookingcalendar.NewModule(ctx),
		BookingDashboard:    bookingdashboard.NewModule(ctx),
		BookingReminders:    bookingreminders.NewModule(ctx),
		BookingResources:    bookingresources.NewModule(ctx),
		BookingServices:     bookingservices.NewModule(ctx),
		BookingWaitlist:     bookingwaitlist.NewModule(ctx),
		Brands:              brands.NewModule(ctx),
		Business:            business.NewModule(ctx),
		Campaigns:           campaigns.NewModule(ctx),
		Careers:             careers.NewModule(ctx),
		Categories:          categories.NewModule(ctx),
		Config:              config.NewModule(ctx),
		CRMSettings:         crmsettings.NewModule(ctx),
		Customers:           customers.NewModule(ctx),
		Cycles:              cycles.NewModule(ctx),
		Departments:         departments.NewModule(ctx),
		Documents:           documents.NewModule(ctx),
		Echo:                echo.NewModule(ctx, memoryModule),
		Expenses:            expenses.NewModule(ctx),
		Feedback:            feedback.NewModule(ctx),
		FlowInsights:        flowinsights.NewModule(ctx),
		FinanceReports:      financereports.NewModule(ctx),
		Insights:            insights.NewModule(ctx),
		Integrations:        integrations.NewModule(ctx),
		Inventory:           inventory.NewModule(ctx),
		Invoices:            invoices.NewModule(ctx),
		Journal:             journal.NewModule(ctx),
		KPIs:                kpis.NewModule(ctx),
		Loyalty:             loyalty.NewModule(ctx),
		Memory:              memoryModule,
		Milestones:          milestones.NewModule(ctx),
		Notifications:       notifications.NewModule(ctx),
		OneOnOnes:           oneonones.NewModule(ctx),
		Payroll:             payroll.NewModule(ctx),
		PosCombos:           poscombos.NewModule(ctx),
		PosCustomers:        poscustomers.NewModule(ctx),
		PosDaySummary:       posdaysummary.NewModule(ctx),
		PosGiftCards:        posgiftcards.NewModule(ctx),
		PosKDS:              poskds.NewModule(ctx),
		PosLocations:        poslocations.NewModule(ctx),
		PosManager:          posmanager.NewModule(ctx),
		PosMenu:             posmenu.NewModule(ctx),
		PosModifiers:        posmodifiers.NewModule(ctx),
		PosOrders:           posorders.NewModule(ctx),
		PosReceipts:         posreceipts.NewModule(ctx),
		PosSessions:         possessions.NewModule(ctx),
		PosTables:           postables.NewModule(ctx),
		PosTerminals:        posterminals.NewModule(ctx),
		PosFloors:           posfloors.NewModule(ctx),
		PosAudit:            posaudit.NewModule(ctx),
		PosCashMovements:    poscashmovements.NewModule(ctx),
		PosLoyalty:          posloyalty.NewModule(ctx),
		PosStockMovements:   posstockmovements.NewModule(ctx),
		PosUsers:            posusers.NewModule(ctx),
		PosMemberLocations:  posmemberlocations.NewModule(ctx),
		Products:            products.NewModule(ctx),
		PurchaseOrders:      purchaseorders.NewModule(ctx),
		Quotes:              quotes.NewModule(ctx),
		RecurringInvoices:   recurringinvoices.NewModule(ctx),
		Reviews:             reviews.NewModule(ctx),
		Roles:               roles.NewModule(ctx),
		Security:            security.NewModule(ctx),
		Segments:            segments.NewModule(ctx),
		Shifts:              shifts.NewModule(ctx),
		Staff:               staff.NewModule(ctx),
		StockTakes:          stocktakes.NewModule(ctx),
		StockTransfers:      stocktransfers.NewModule(ctx),
		StoreCategories:     storecategories.NewModule(ctx),
		StoreCheckout:       storecheckout.NewModule(ctx),
		Storefronts:         storefronts.NewModule(ctx),
		StoreOrders:         storeorders.NewModule(ctx),
		StorePayments:       storepayments.NewModule(ctx),
		StoreProducts:       storeproducts.NewModule(ctx),
		StoreShipping:       storeshipping.NewModule(ctx),
		StoreThemes:         storethemes.NewModule(ctx),
		SupportCanned:       supportcanned.NewModule(ctx),
		SupportCustomers:    supportcustomers.NewModule(ctx),
		SupportInbox:        supportinbox.NewModule(ctx),
		SupportKnowledge:    supportknowledge.NewModule(ctx),
		SupportSettings:     supportsettings.NewModule(ctx),
		SupportTickets:      supporttickets.NewModule(ctx),
		Suppliers:           suppliers.NewModule(ctx),
		Tax:                 tax.NewModule(ctx),
		Team:                team.NewModule(ctx),
		Templates:           templates.NewModule(ctx),
		TimeOff:             timeoff.NewModule(ctx),
		Timesheets:          timesheets.NewModule(ctx),
		Training:            training.NewModule(ctx),
		User:                user.NewModule(ctx),
		Variants:            variants.NewModule(ctx),
		Views:               views.NewModule(ctx),
		Warehouses:          warehouses.NewModule(ctx),
		Webhooks:            webhooks.NewModule(ctx),
		Workspaces:          workspaces.NewModule(ctx),
	}
}

func setupRoutes(ctx context.Context, r *chi.Mux) {
	modules := initializeModules(ctx)
	accessPolicy := http.NewRouteAccessPolicy()
	accessPolicy.RegisterAnonymousRoutes("/health")

	provisionerResolver := authPkg.NewStaticPlatformProvisionerResolver(
		nil,
	)
	modules.User.ProvisionerResolver = provisionerResolver

	// Wire personal workspace creation into user registration
	user.OnUserCreated = func(ctx context.Context, userID int64) {
		modules.Workspaces.CreatePersonalWorkspace(ctx, userID)
	}

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
