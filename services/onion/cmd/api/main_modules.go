package main

import (
	flowreferences "github.com/voxire/onion/internal/modules/flow/references"
	identityadapter "github.com/voxire/onion/internal/modules/flow/workspaces/adapters/identity"
	posreferences "github.com/voxire/onion/internal/modules/pos/references"
	"github.com/voxire/onion/internal/platform/authorization"

	"context"

	"github.com/voxire/onion/internal/modules/activity"
	"github.com/voxire/onion/internal/modules/aiquery"
	"github.com/voxire/onion/internal/modules/aiquota"
	"github.com/voxire/onion/internal/modules/automations"
	booking_appointments "github.com/voxire/onion/internal/modules/booking/appointments"
	booking_calendar "github.com/voxire/onion/internal/modules/booking/calendar"
	booking_dashboard "github.com/voxire/onion/internal/modules/booking/dashboard"
	booking_reminders "github.com/voxire/onion/internal/modules/booking/reminders"
	booking_resources "github.com/voxire/onion/internal/modules/booking/resources"
	booking_services "github.com/voxire/onion/internal/modules/booking/services"
	booking_waitlist "github.com/voxire/onion/internal/modules/booking/waitlist"
	"github.com/voxire/onion/internal/modules/cacheops"
	echo "github.com/voxire/onion/internal/modules/caret"
	"github.com/voxire/onion/internal/modules/caret/memory"
	"github.com/voxire/onion/internal/modules/crm/campaigns"
	crmsales "github.com/voxire/onion/internal/modules/crm/crm_sales"
	crmsettings "github.com/voxire/onion/internal/modules/crm/crm_settings"
	"github.com/voxire/onion/internal/modules/crm/customers"
	"github.com/voxire/onion/internal/modules/crm/feedback"
	"github.com/voxire/onion/internal/modules/crm/loyalty"
	"github.com/voxire/onion/internal/modules/crm/segments"
	"github.com/voxire/onion/internal/modules/crm/templates"
	"github.com/voxire/onion/internal/modules/emailtest"
	"github.com/voxire/onion/internal/modules/erp/brands"
	"github.com/voxire/onion/internal/modules/erp/categories"
	erpsettings "github.com/voxire/onion/internal/modules/erp/erp_settings"
	"github.com/voxire/onion/internal/modules/erp/inventory"
	"github.com/voxire/onion/internal/modules/erp/onlineorders"
	"github.com/voxire/onion/internal/modules/erp/price_lists"
	"github.com/voxire/onion/internal/modules/erp/products"
	"github.com/voxire/onion/internal/modules/erp/purchase_orders"
	stocktakes "github.com/voxire/onion/internal/modules/erp/stock_takes"
	stocktransfers "github.com/voxire/onion/internal/modules/erp/stock_transfers"
	"github.com/voxire/onion/internal/modules/erp/store_categories"
	"github.com/voxire/onion/internal/modules/erp/store_products"
	"github.com/voxire/onion/internal/modules/erp/suppliers"
	"github.com/voxire/onion/internal/modules/erp/variants"
	"github.com/voxire/onion/internal/modules/erp/warehouses"
	"github.com/voxire/onion/internal/modules/finance/accounts"
	bankaccounts "github.com/voxire/onion/internal/modules/finance/bank_accounts"
	"github.com/voxire/onion/internal/modules/finance/bills"
	"github.com/voxire/onion/internal/modules/finance/expenses"
	financebankfeeds "github.com/voxire/onion/internal/modules/finance/finance_bank_feeds"
	financereconciliation "github.com/voxire/onion/internal/modules/finance/finance_reconciliation"
	financereports "github.com/voxire/onion/internal/modules/finance/finance_reports"
	"github.com/voxire/onion/internal/modules/finance/invoices"
	invoiceadapters "github.com/voxire/onion/internal/modules/finance/invoices/adapters"
	"github.com/voxire/onion/internal/modules/finance/journal"
	"github.com/voxire/onion/internal/modules/finance/quotes"
	recurringinvoices "github.com/voxire/onion/internal/modules/finance/recurring_invoices"
	"github.com/voxire/onion/internal/modules/finance/tax"
	flow_insights "github.com/voxire/onion/internal/modules/flow/insights"
	"github.com/voxire/onion/internal/modules/flow/workspaces"
	workspaceadapters "github.com/voxire/onion/internal/modules/flow/workspaces/adapters"
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
	oneonones "github.com/voxire/onion/internal/modules/hr/one_on_ones"
	"github.com/voxire/onion/internal/modules/hr/payroll"
	"github.com/voxire/onion/internal/modules/hr/reviews"
	"github.com/voxire/onion/internal/modules/hr/staff"
	"github.com/voxire/onion/internal/modules/hr/time_off"
	"github.com/voxire/onion/internal/modules/hr/timesheets"
	"github.com/voxire/onion/internal/modules/hr/training"
	hubmenu "github.com/voxire/onion/internal/modules/hub_menu"
	"github.com/voxire/onion/internal/modules/insights"
	insightsadapters "github.com/voxire/onion/internal/modules/insights/adapters"
	"github.com/voxire/onion/internal/modules/insights/kpis"
	"github.com/voxire/onion/internal/modules/integrations"
	"github.com/voxire/onion/internal/modules/notifications"
	"github.com/voxire/onion/internal/modules/organization/addons"
	apikeys "github.com/voxire/onion/internal/modules/organization/api_keys"
	"github.com/voxire/onion/internal/modules/organization/apps"
	"github.com/voxire/onion/internal/modules/organization/billing"
	"github.com/voxire/onion/internal/modules/organization/business"
	businessadapters "github.com/voxire/onion/internal/modules/organization/business/adapters"
	"github.com/voxire/onion/internal/modules/organization/config"
	"github.com/voxire/onion/internal/modules/organization/console"
	"github.com/voxire/onion/internal/modules/organization/featuresupport"
	"github.com/voxire/onion/internal/modules/organization/oauth"
	"github.com/voxire/onion/internal/modules/organization/roles"
	"github.com/voxire/onion/internal/modules/organization/scopes"
	"github.com/voxire/onion/internal/modules/organization/security"
	"github.com/voxire/onion/internal/modules/organization/subscriptions"
	"github.com/voxire/onion/internal/modules/organization/team"
	teamadapters "github.com/voxire/onion/internal/modules/organization/team/adapters"
	"github.com/voxire/onion/internal/modules/organization/user"
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
	store_checkout "github.com/voxire/onion/internal/modules/store/checkout"
	store_payments "github.com/voxire/onion/internal/modules/store/payments"
	store_shipping "github.com/voxire/onion/internal/modules/store/shipping"
	"github.com/voxire/onion/internal/modules/store/storefronts"
	store_themes "github.com/voxire/onion/internal/modules/store/themes"
	supportcanned "github.com/voxire/onion/internal/modules/support/canned"
	supportcustomers "github.com/voxire/onion/internal/modules/support/customers"
	supportinbox "github.com/voxire/onion/internal/modules/support/inbox"
	supportknowledge "github.com/voxire/onion/internal/modules/support/knowledge"
	supportsettings "github.com/voxire/onion/internal/modules/support/settings"
	supporttickets "github.com/voxire/onion/internal/modules/support/tickets"
	"github.com/voxire/onion/internal/modules/webhooks"
	"github.com/voxire/onion/internal/modules/websitelead"
	platformConfig "github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/tenant"
)

func initializeModules(ctx context.Context) *apiModules {
	tenant.SetScopeResolver(scopes.NewReader(dbclient.Pool()))
	flowReferenceReader := flowreferences.NewReader(dbclient.Pool())
	posReferenceReader := posreferences.NewReader(dbclient.Pool())
	_ = authorization.SetScopeLookup(func(ctx context.Context, businessID int64, scopeType string, id int64) (bool, error) {
		if scopeType == "flow_project" {
			return flowReferenceReader.ProjectExists(ctx, businessID, id)
		}
		return posReferenceReader.LocationExists(ctx, businessID, id)
	})

	memoryModule := memory.NewModule(ctx, dbclient.Pool()).WithIdentityReaders(scopes.NewReader(dbclient.Pool()).DefaultBusiness, scopes.NewReader(dbclient.Pool()).AssistantIdentity)
	subscriptionsModule := subscriptions.NewModule(ctx, dbclient.Pool())
	aiQuotaModule := aiquota.NewModule(ctx, dbclient.Pool(), subscriptionsModule)
	aiQueryModule := aiquery.NewModule(ctx)
	businessModule := business.NewModule(ctx, dbclient.Pool())
	addonsModule := addons.NewModule(ctx, dbclient.Pool())
	hrCoreModule := hr_core.NewModule(ctx, dbclient.Pool()).
		WithAccountEmailReader(scopes.NewReader(dbclient.Pool()).AccountEmail).
		WithBusinessOwnerResolver(businessModule).
		WithBusinessIdentityResolver(businessModule)
	businessModule.WithPublicLocationReader(businessadapters.NewHRPublicLocationReader(hrCoreModule))
	hrStaffModule := staff.NewModule(ctx, dbclient.Pool(), hrCoreModule)
	hrAttendanceModule := hr_attendance.NewModule(ctx, dbclient.Pool(), hrCoreModule)
	hrLeaveModule := hr_leave.NewModule(ctx, dbclient.Pool(), hrCoreModule)
	hrOvertimeModule := hr_overtime.NewModule(ctx, dbclient.Pool(), hrCoreModule)
	hrPayrollModule := hr_payroll.NewModule(ctx, dbclient.Pool(), hrCoreModule)
	hrTimeModule := hr_time.NewModule(ctx, dbclient.Pool(), hrCoreModule)
	customersModule := customers.NewModule(ctx, dbclient.Pool()).WithReportingTimezone(businessModule.ReportingTimezone)
	automationContacts := func(ctx context.Context, customerID, businessID int64) (automations.CustomerContact, bool, error) {
		recipient, found, err := customersModule.ResolveEmailRecipient(ctx, businessID, customerID)
		return automations.CustomerContact{Email: recipient.Email, Name: recipient.Name}, found, err
	}
	campaignsModule := campaigns.NewModule(ctx, dbclient.Pool())
	inventoryModule := inventory.NewModule(ctx, dbclient.Pool())
	productsModule := products.NewModule(ctx, dbclient.Pool())
	productsModule.SetCatalogBusinessContext(businessModule.GetBusinessOwnerID, addonsModule.OnlineOrderingEnabled)
	categoriesModule := categories.NewModule(ctx, dbclient.Pool())
	brandsModule := brands.NewModule(ctx, dbclient.Pool())
	loyaltyModule := loyalty.NewModule(ctx, dbclient.Pool())
	reviewsModule := reviews.NewModule(ctx, dbclient.Pool())
	segmentsModule := segments.NewModule(ctx, dbclient.Pool())
	// flowInsights is constructed first: insights.dashboard.v1's Flow-shaped
	// projections delegate to it rather than querying flow_* directly.
	flowInsightsModule := flow_insights.NewModule(ctx, dbclient.Pool())
	teamModule := team.NewModule(ctx, dbclient.Pool())
	workspacesModule := workspaces.NewModule(ctx, dbclient.Pool()).WithIdentityReaders(identityadapter.NewIdentityReader(scopes.NewReader(dbclient.Pool())), scopes.NewReader(dbclient.Pool()).AccountByEmail, scopes.NewReader(dbclient.Pool()).BusinessUserByUsername)
	userModule := func() *user.Module {
		cfg := platformConfig.GlobalConfig
		m := user.NewModule(ctx, dbclient.Pool(), cfg.GoogleOAuthClientID)
		m.WithBusinessProvisioner(hrCoreModule)
		m.SetEmployeeProvisioner(hrCoreModule)
		m.WithGoogleOAuth(cfg.GoogleOAuthClientID, cfg.GoogleOAuthClientSecret, cfg.GoogleOAuthRedirectURI)
		return m
	}()
	workspacesModule.SetWorkDayLineStore(workspaceadapters.NewHRTimeWorkDayLineStore(hrTimeModule, hrCoreModule))
	workspacesModule.SetWorkdayRosterSource(workspaceadapters.NewWorkdayRosterSource(hrCoreModule, teamModule))
	workspacesModule.SetInsightsMemberProfileSource(workspaceadapters.NewInsightsMemberProfileSource(userModule))
	workspacesModule.SetPresenceSource(workspaceadapters.NewPresenceSource(hrCoreModule, hrLeaveModule))
	workspacesModule.Service().SetLabourCostProvider(hrPayrollModule.LabourCostProvider())
	notificationsModule := notifications.NewModule(ctx, dbclient.Pool())
	teamModule.SetNotificationPublisher(teamadapters.NewNotificationPublisher(notificationsModule))
	teamModule.SetMemberProvisioner(teamadapters.NewHRMemberProvisioner(hrCoreModule, hrStaffModule))
	teamModule.SetTerminatedEmployeeSource(teamadapters.NewTerminatedEmployeeSource(hrCoreModule))
	expensesModule := expenses.NewModule(ctx, dbclient.Pool())
	posOrdersModule := orders.NewModule(ctx, dbclient.Pool()).
		WithSalesProductReader(productsModule.SalesProduct).
		WithOnlineOrderingReader(addonsModule.OnlineOrderingEnabled)
	posLocationsModule := locations.NewModule(ctx, dbclient.Pool()).WithLocationLimitReader(scopes.NewReader(dbclient.Pool()).LocationLimit)
	inventoryModule.WithBusinessOwnerResolver(businessModule.GetBusinessOwnerID).WithPOSBranchOwnerReader(posLocationsModule.StockBranchOwner)
	posTerminalsModule := terminals.NewModule(ctx, dbclient.Pool())
	posUsersModule := users.NewModule(ctx, dbclient.Pool()).WithAuthorizationVersionReader(scopes.NewReader(dbclient.Pool()).AuthorizationVersion)
	posFloorsModule := floors.NewModule(ctx, dbclient.Pool())
	posTablesModule := tables.NewModule(ctx, dbclient.Pool())
	posModifiersModule := modifiers.NewModule(ctx, dbclient.Pool())
	posSessionsModule := sessions.NewModule(ctx, dbclient.Pool())
	posCashMovementsModule := cashmovements.NewModule(ctx, dbclient.Pool())
	suppliersModule := suppliers.NewModule(ctx, dbclient.Pool())
	warehousesModule := warehouses.NewModule(ctx, dbclient.Pool())
	purchaseOrdersModule := purchase_orders.NewModule(ctx, dbclient.Pool())
	stockTransfersModule := stocktransfers.NewModule(ctx, dbclient.Pool())
	bankAccountsModule := bankaccounts.NewModule(ctx, dbclient.Pool()).WithReportingCurrency(businessModule.ReportingCurrency)
	billsModule := bills.NewModule(ctx, dbclient.Pool()).WithPurchaseOrderReader(purchaseOrdersModule.Exists)
	quotesModule := quotes.NewModule(ctx, dbclient.Pool())
	bookingDashboardModule := booking_dashboard.NewModule(ctx, dbclient.Pool())
	bookingAppointmentsModule := booking_appointments.NewModule(ctx, dbclient.Pool())
	bookingCalendarModule := booking_calendar.NewModule(ctx, dbclient.Pool())
	bookingRemindersModule := booking_reminders.NewModule(ctx, dbclient.Pool())
	bookingResourcesModule := booking_resources.NewModule(ctx, dbclient.Pool())
	bookingServicesModule := booking_services.NewModule(ctx, dbclient.Pool())
	bookingWaitlistModule := booking_waitlist.NewModule(ctx, dbclient.Pool())
	financeBankFeedsModule := financebankfeeds.NewModule(ctx, dbclient.Pool())
	financeReconciliationModule := financereconciliation.NewModule(ctx, dbclient.Pool())
	financeReportsModule := financereports.NewModule(ctx, dbclient.Pool()).WithReportingCurrency(businessModule.ReportingCurrency)
	insightsModule := insights.NewModule(ctx, dbclient.Pool(), insightsadapters.NewFlowReader(flowInsightsModule), insights.DashboardReaders{
		Finance: financeReportsModule,
		POS:     posOrdersModule,
		CRM:     customersModule,
	})
	accountsModule := accounts.NewModule(ctx, dbclient.Pool())
	journalModule := journal.NewModule(ctx, dbclient.Pool())
	journalModule.SetPayrollAccountProvider(accountsModule.PayrollAccountProvider())
	hrPayrollModule.SetFinancePoster(journalModule.PayrollPoster())
	kpisModule := kpis.NewModule(ctx, dbclient.Pool())
	posDaySummaryModule := daysummary.NewModule(ctx, dbclient.Pool())
	invoicesModule := invoices.NewModule(ctx, dbclient.Pool())
	invoicesModule.SetBillableTimeReader(invoiceadapters.NewBillableTimeReader(hrTimeModule))
	hrTimeModule.SetInvoiceExistenceProvider(invoicesModule)
	// CRM->FIN-02 (#840): invoices verify a supplied customer_id against the
	// customers module and snapshot its name/email/address onto the invoice.
	// Port lives in invoices; customers implements it; no package imports.
	invoicesModule.SetCustomerSnapshotProvider(customersModule.SnapshotProvider())

	storeCategoriesModule := store_categories.NewModule(ctx, dbclient.Pool())
	storeOrdersModule := onlineorders.NewModule(ctx, dbclient.Pool())
	storeProductsModule := store_products.NewModule(ctx, dbclient.Pool())
	supportCannedModule := supportcanned.NewModule(ctx, dbclient.Pool())
	supportCustomersModule := supportcustomers.NewModule(ctx, dbclient.Pool())
	supportInboxModule := supportinbox.NewModule(ctx, dbclient.Pool())
	supportKnowledgeModule := supportknowledge.NewModule(ctx, dbclient.Pool())
	supportTicketsModule := supporttickets.NewModule(ctx, dbclient.Pool())
	taxModule := tax.NewModule(ctx, dbclient.Pool())
	consoleModule := console.NewModule(ctx, dbclient.Pool())
	consoleModule.SetEmployeeProvisioner(hrCoreModule)

	s1FormsExecutor, s1FormsErr := echo.NewFormMappingToolExecutor(platformConfig.GlobalConfig.FormMappingCheckpointPath)
	if s1FormsErr != nil {
		logger.Ctx(ctx).Warn("Caret document mapping checkpoint unavailable; using lexical fallback", logger.Error(s1FormsErr))
	}

	return &apiModules{
		Accounts:              accountsModule,
		Activity:              activity.NewModule(ctx, dbclient.Pool()).WithUserNameReader(scopes.NewReader(dbclient.Pool()).BusinessUserName),
		Addons:                addonsModule,
		APIKeys:               apikeys.NewModule(ctx, dbclient.Pool()),
		Apps:                  apps.NewModule(ctx, dbclient.Pool()),
		Automations:           automations.NewModule(ctx, dbclient.Pool(), automationContacts),
		BankAccounts:          bankAccountsModule,
		Benefits:              benefits.NewModule(ctx, dbclient.Pool(), hrCoreModule),
		Billing:               billing.NewModule(ctx, dbclient.Pool()),
		Subscriptions:         subscriptionsModule,
		Bills:                 billsModule,
		BookingAppointments:   bookingAppointmentsModule,
		BookingCalendar:       bookingCalendarModule,
		BookingDashboard:      bookingDashboardModule,
		BookingReminders:      bookingRemindersModule,
		BookingResources:      bookingResourcesModule,
		BookingServices:       bookingServicesModule,
		BookingWaitlist:       bookingWaitlistModule,
		Brands:                brandsModule,
		Business:              businessModule,
		Campaigns:             campaignsModule,
		Careers:               careers.NewModule(ctx, dbclient.Pool()),
		Categories:            categoriesModule,
		Console:               consoleModule,
		Config:                config.NewModule(ctx, dbclient.Pool()),
		CRMSettings:           crmsettings.NewModule(ctx, dbclient.Pool()),
		CRMSales:              crmsales.NewModule(ctx, dbclient.Pool()).WithMembershipReader(scopes.NewReader(dbclient.Pool()).IsBusinessUser).WithRepReader(crmSalesReps(scopes.NewReader(dbclient.Pool()))).WithBusinessOwnerResolver(businessModule),
		Customers:             customersModule,
		FeatureSupport:        featuresupport.NewModule(ctx, dbclient.Pool()),
		Departments:           departments.NewModule(ctx, dbclient.Pool()),
		Documents:             documents.NewModule(ctx, dbclient.Pool()),
		CacheOps:              cacheops.NewModule(),
		AIQuota:               aiQuotaModule,
		Echo:                  echo.NewModule(ctx, memoryModule, aiQuotaModule, aiQueryModule, memoryModule, insightsModule, financeReportsModule, flowInsightsModule, bookingDashboardModule, posDaySummaryModule, kpisModule, workspacesModule, customersModule, campaignsModule, segmentsModule, loyaltyModule, reviewsModule, inventoryModule, productsModule, categoriesModule, brandsModule, suppliersModule, warehousesModule, purchaseOrdersModule, stockTransfersModule, expensesModule, posOrdersModule, posLocationsModule, posTerminalsModule, posUsersModule, posFloorsModule, posTablesModule, posModifiersModule, posSessionsModule, posCashMovementsModule, hrCoreModule, hrAttendanceModule, hrLeaveModule, hrOvertimeModule, hrPayrollModule, hrTimeModule, billsModule, quotesModule, bookingAppointmentsModule, bookingCalendarModule, bookingRemindersModule, bookingResourcesModule, bookingServicesModule, bookingWaitlistModule, bankAccountsModule, financeBankFeedsModule, financeReconciliationModule, journalModule, taxModule, invoicesModule, storeOrdersModule, storeProductsModule, storeCategoriesModule, supportTicketsModule, supportInboxModule, supportCustomersModule, supportCannedModule, supportKnowledgeModule, s1FormsExecutor),
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
		HRDashboard:           hr_dashboard.NewModule(ctx, dbclient.Pool()),
		HRDocuments:           hr_documents.NewModule(ctx, dbclient.Pool()),
		HubMenu:               hubmenu.NewModule(ctx, dbclient.Pool()),
		HREquipment:           hr_equipment.NewModule(ctx, dbclient.Pool(), hrCoreModule),
		HRLeave:               hrLeaveModule,
		HROnboarding:          hr_onboarding.NewModule(ctx, dbclient.Pool()),
		HROvertime:            hrOvertimeModule,
		HRPayroll:             hrPayrollModule,
		HRPerformance:         hr_performance.NewModule(ctx, dbclient.Pool()),
		HRReports:             hr_reports.NewModule(ctx, dbclient.Pool()),
		HRSetup:               hr_setup.NewModule(ctx, dbclient.Pool(), hrCoreModule),
		HRShifts:              hr_shifts.NewModule(ctx, dbclient.Pool(), hrCoreModule),
		HRTime:                hrTimeModule,
		Insights:              insightsModule,
		Integrations:          integrations.NewModule(ctx, dbclient.Pool()),
		Inventory:             inventoryModule,
		Invoices:              invoicesModule,
		Journal:               journalModule,
		KPIs:                  kpisModule,
		Loyalty:               loyaltyModule,
		Memory:                memoryModule,
		Notifications:         notificationsModule,
		OAuth:                 oauth.NewModule(ctx, dbclient.Pool(), platformConfig.GlobalConfig.AuthAppURL),
		OneOnOnes:             oneonones.NewModule(ctx, dbclient.Pool()),
		Payroll:               payroll.NewModule(ctx, dbclient.Pool(), hrCoreModule),
		PosAudit:              audit.NewModule(ctx, dbclient.Pool()),
		PosCashMovements:      posCashMovementsModule,
		PosCombos:             combos.NewModule(ctx, dbclient.Pool()),
		PosDaySummary:         posDaySummaryModule,
		PosFloors:             posFloorsModule,
		PosGiftCards:          giftcards.NewModule(ctx, dbclient.Pool()),
		PosKDS:                kds.NewModule(ctx, dbclient.Pool()),
		PosLocations:          posLocationsModule,
		PosLoyalty:            posloyalty.NewModule(ctx, dbclient.Pool()),
		PosManager:            manager.NewModule(ctx, dbclient.Pool()).WithPasswordVerifier(scopes.NewReader(dbclient.Pool()).VerifyAccountPassword),
		PosMemberLocations:    memberlocations.NewModule(ctx, dbclient.Pool()).WithIdentityReaders(scopes.NewReader(dbclient.Pool()).BusinessUserName, scopes.NewReader(dbclient.Pool()).IsBusinessUser),
		PosModifiers:          posModifiersModule,
		PosOrders:             posOrdersModule,
		PosReceipts:           receipts.NewModule(ctx, dbclient.Pool()),
		PosSessions:           posSessionsModule,
		PosSettings:           settings.NewModule(ctx, dbclient.Pool()),
		PosTables:             posTablesModule,
		PosTerminals:          posTerminalsModule,
		PosUsers:              posUsersModule,
		PriceLists:            price_lists.NewModule(ctx, dbclient.Pool()),
		Products:              productsModule,
		PurchaseOrders:        purchaseOrdersModule,
		Quotes:                quotesModule,
		RecurringInvoices:     recurringinvoices.NewModule(ctx, dbclient.Pool()),
		Reviews:               reviewsModule,
		Roles:                 roles.NewModule(ctx, dbclient.Pool()),
		Security:              security.NewModule(ctx, dbclient.Pool()),
		Segments:              segmentsModule,
		Staff:                 hrStaffModule,
		StockTakes:            stocktakes.NewModule(ctx, dbclient.Pool()),
		StockTransfers:        stockTransfersModule,
		StoreCategories:       storeCategoriesModule,
		StoreCheckout:         store_checkout.NewModule(ctx, dbclient.Pool()).WithSalesProductReader(productsModule.SalesProduct),
		Storefronts:           storefronts.NewModule(ctx, dbclient.Pool()),
		StoreOrders:           storeOrdersModule,
		StorePayments:         store_payments.NewModule(ctx, dbclient.Pool()),
		StoreProducts:         storeProductsModule,
		StoreShipping:         store_shipping.NewModule(ctx, dbclient.Pool()),
		StoreThemes:           store_themes.NewModule(ctx, dbclient.Pool()),
		SupportCanned:         supportCannedModule,
		SupportCustomers:      supportCustomersModule,
		SupportInbox:          supportInboxModule,
		SupportKnowledge:      supportKnowledgeModule,
		SupportSettings:       supportsettings.NewModule(ctx, dbclient.Pool()),
		SupportTickets:        supportTicketsModule,
		Suppliers:             suppliersModule,
		Tax:                   taxModule,
		Team:                  teamModule,
		Templates:             templates.NewModule(ctx, dbclient.Pool()),
		TimeOff:               time_off.NewModule(ctx, dbclient.Pool()),
		Timesheets:            timesheets.NewModule(ctx, dbclient.Pool()),
		Training:              training.NewModule(ctx, dbclient.Pool()),
		User:                  userModule,
		Variants:              variants.NewModule(ctx, dbclient.Pool()),
		Voxhire:               voxhire.NewModule(ctx, dbclient.Pool()),
		Warehouses:            warehousesModule,
		Webhooks:              webhooks.NewModule(ctx, dbclient.Pool()),
		WebsiteLead: websitelead.NewModule(ctx, dbclient.Pool(), websitelead.Notify{
			LeadEmails:       platformConfig.GlobalConfig.LeadNotifyEmails,
			NewsletterEmails: platformConfig.GlobalConfig.NewsletterNotifyEmails,
		}),
		Workspaces: workspacesModule,
	}
}

// crmSalesReps lists the teammates a CRM lead can be assigned to: id and
// display name only. Any CRM viewer can read it, so a member with no name on
// file is not shown by email; the client labels them by id instead.
func crmSalesReps(reader *scopes.Reader) func(context.Context, int64) ([]crmsales.Rep, error) {
	return func(ctx context.Context, businessID int64) ([]crmsales.Rep, error) {
		users, err := reader.ActiveBusinessUsers(ctx, businessID)
		if err != nil {
			return nil, err
		}
		reps := make([]crmsales.Rep, 0, len(users))
		for _, user := range users {
			reps = append(reps, crmsales.Rep{ID: user.ID, Name: user.Name})
		}
		return reps, nil
	}
}
