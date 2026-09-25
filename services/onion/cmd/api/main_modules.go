package main

import (
	"context"

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
	platformConfig "github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

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

	s1FormsExecutor, s1FormsErr := echo.NewFormMappingToolExecutor(platformConfig.GlobalConfig.FormMappingCheckpointPath)
	if s1FormsErr != nil {
		logger.Ctx(ctx).Warn("Caret document mapping checkpoint unavailable; using lexical fallback", logger.Error(s1FormsErr))
	}

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
		Echo:                  echo.NewModule(ctx, memoryModule, aiQuotaModule, aiQueryModule, memoryModule, insightsModule, financeReportsModule, flowInsightsModule, bookingDashboardModule, posDaySummaryModule, kpisModule, workspacesModule, customersModule, campaignsModule, segmentsModule, loyaltyModule, reviewsModule, inventoryModule, productsModule, categoriesModule, brandsModule, suppliersModule, warehousesModule, purchaseOrdersModule, stockTransfersModule, expensesModule, posOrdersModule, posLocationsModule, posTerminalsModule, posUsersModule, posFloorsModule, posTablesModule, posMenuModule, posModifiersModule, posSessionsModule, posCashMovementsModule, posStockMovementsModule, hrCoreModule, hrAttendanceModule, hrLeaveModule, hrOvertimeModule, hrPayrollModule, hrTimeModule, billsModule, quotesModule, bookingAppointmentsModule, bookingCalendarModule, bookingRemindersModule, bookingResourcesModule, bookingServicesModule, bookingWaitlistModule, bankAccountsModule, financeBankFeedsModule, financeReconciliationModule, journalModule, taxModule, invoicesModule, storeOrdersModule, storeProductsModule, storeCategoriesModule, supportTicketsModule, supportInboxModule, supportCustomersModule, supportCannedModule, supportKnowledgeModule, s1FormsExecutor),
		EmailTest:             emailtest.NewModule(ctx, dbclient.Pool()),
		Expenses:              expensesModule,
		ERPSettings:           erpsettings.NewModule(ctx, dbclient.Pool()),
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
		Variants:   variants.NewModule(ctx, dbclient.Pool()),
		Voxhire:    voxhire.NewModule(ctx, dbclient.Pool()),
		Warehouses: warehousesModule,
		Webhooks:   webhooks.NewModule(ctx, dbclient.Pool()),
		WebsiteLead: websitelead.NewModule(ctx, dbclient.Pool(), websitelead.Notify{
			LeadEmails:       platformConfig.GlobalConfig.LeadNotifyEmails,
			NewsletterEmails: platformConfig.GlobalConfig.NewsletterNotifyEmails,
		}),
		Workspaces: workspacesModule,
	}
}
