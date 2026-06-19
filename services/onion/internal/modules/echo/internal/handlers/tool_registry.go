package handlers

// ToolDef describes a callable tool exposed to Caret. The registry is used in
// prompts and stream events; execution starts with server-owned read-only tools
// and can grow toward approval-backed write tools.
type ToolDef struct {
	Name             string `json:"name"`
	Description      string `json:"description"`
	Module           string `json:"module"`
	Scope            string `json:"scope"` // "read" or "write"
	RequiresApproval bool   `json:"requiresApproval"`
}

// CoreTools is the seed registry — the long-term plan is to source this
// list from each module's MCP descriptor.
var CoreTools = []ToolDef{
	{Name: "business.snapshot", Description: "Build a tenant-scoped owner snapshot across sales, finance, inventory, CRM, HR, and Flow.", Module: "business", Scope: "read"},
	{Name: "business.get_kpis", Description: "Get headline business KPIs across revenue, orders, staff, customers, stock, and apps.", Module: "business", Scope: "read"},
	{Name: "analytics.query", Description: "Run a tenant-scoped, read-only analytics query over allowlisted business views.", Module: "analytics", Scope: "read"},
	{Name: "insights.get_dashboard", Description: "Get cross-domain business dashboard headline KPIs.", Module: "insights", Scope: "read"},
	{Name: "insights.get_revenue_series", Description: "Get daily revenue over a date range.", Module: "insights", Scope: "read"},
	{Name: "insights.get_revenue_comparison", Description: "Compare revenue for a period against the previous period.", Module: "insights", Scope: "read"},
	{Name: "insights.get_revenue_by_app", Description: "Get revenue split by app surface.", Module: "insights", Scope: "read"},
	{Name: "insights.get_customer_kpis", Description: "Get customer acquisition, active, and churn counters.", Module: "insights", Scope: "read"},
	{Name: "insights.list_top_customers", Description: "List top customers by lifetime spend.", Module: "insights", Scope: "read"},
	{Name: "insights.get_customer_churn", Description: "Get customer churn rates across rolling windows.", Module: "insights", Scope: "read"},
	{Name: "insights.get_ltv_distribution", Description: "Get customer lifetime-value distribution buckets.", Module: "insights", Scope: "read"},
	{Name: "insights.get_product_kpis", Description: "Get product catalog and margin KPIs.", Module: "insights", Scope: "read"},
	{Name: "insights.list_bestsellers", Description: "List best-selling products.", Module: "insights", Scope: "read"},
	{Name: "insights.list_slow_movers", Description: "List slow-moving products.", Module: "insights", Scope: "read"},
	{Name: "insights.get_sales_by_category", Description: "Get product sales grouped by category.", Module: "insights", Scope: "read"},
	{Name: "insights.get_margin_distribution", Description: "Get product margin distribution buckets.", Module: "insights", Scope: "read"},
	{Name: "insights.get_cash_flow", Description: "Get cash inflow and outflow over a date range.", Module: "insights", Scope: "read"},
	{Name: "insights.get_expenses_by_category", Description: "Get expenses grouped by category over a date range.", Module: "insights", Scope: "read"},
	{Name: "insights.get_invoice_aging", Description: "Get accounts receivable invoice aging buckets.", Module: "insights", Scope: "read"},
	{Name: "insights.get_operations_kpis", Description: "Get HR operations KPI counters.", Module: "insights", Scope: "read"},
	{Name: "insights.get_headcount", Description: "Get active headcount and month-over-month change.", Module: "insights", Scope: "read"},
	{Name: "insights.get_shift_metrics", Description: "Get scheduled, completed, and no-show shifts over a date range.", Module: "insights", Scope: "read"},
	{Name: "insights.get_time_off", Description: "Get pending and approved time-off counters.", Module: "insights", Scope: "read"},
	{Name: "insights.get_utilization", Description: "Get scheduled-vs-available staff utilization.", Module: "insights", Scope: "read"},
	{Name: "insights.get_settings", Description: "Get insights currency, fiscal-year, and timezone settings.", Module: "insights", Scope: "read"},
	{Name: "crm.list_customers", Description: "List CRM customers.", Module: "crm", Scope: "read"},
	{Name: "crm.create_customer", Description: "Create a CRM customer.", Module: "crm", Scope: "write", RequiresApproval: true},
	{Name: "crm.list_campaigns", Description: "List CRM marketing campaigns and engagement counts.", Module: "crm", Scope: "read"},
	{Name: "crm.get_campaign", Description: "Get a CRM marketing campaign.", Module: "crm", Scope: "read"},
	{Name: "crm.list_segments", Description: "List CRM customer segments.", Module: "crm", Scope: "read"},
	{Name: "loyalty.get_config", Description: "Get loyalty program configuration.", Module: "loyalty", Scope: "read"},
	{Name: "loyalty.list_members", Description: "List loyalty members by points and visits.", Module: "loyalty", Scope: "read"},
	{Name: "loyalty.list_rewards", Description: "List loyalty rewards.", Module: "loyalty", Scope: "read"},
	{Name: "loyalty.list_transactions", Description: "List loyalty point transactions.", Module: "loyalty", Scope: "read"},
	{Name: "pos.list_orders", Description: "List recent POS orders.", Module: "pos", Scope: "read"},
	{Name: "pos.list_locations", Description: "List POS locations and branches.", Module: "pos", Scope: "read"},
	{Name: "pos.list_terminals", Description: "List POS terminals and register status.", Module: "pos", Scope: "read"},
	{Name: "pos.list_users", Description: "List POS users, roles, and operational permissions.", Module: "pos", Scope: "read"},
	{Name: "pos.list_floors", Description: "List POS floors for a branch.", Module: "pos", Scope: "read"},
	{Name: "pos.list_tables", Description: "List POS tables and seating status.", Module: "pos", Scope: "read"},
	{Name: "pos.list_menu_items", Description: "List POS menu items and sale prices.", Module: "pos", Scope: "read"},
	{Name: "pos.list_modifiers", Description: "List POS menu modifiers and upcharges.", Module: "pos", Scope: "read"},
	{Name: "pos.list_sessions", Description: "List POS register sessions.", Module: "pos", Scope: "read"},
	{Name: "pos.list_cash_movements", Description: "List POS cash movements for a session.", Module: "pos", Scope: "read"},
	{Name: "pos.list_stock_movements", Description: "List POS stock movement history for a branch product.", Module: "pos", Scope: "read"},
	{Name: "pos.list_branch_inventory", Description: "List POS branch inventory.", Module: "pos", Scope: "read"},
	{Name: "pos.list_low_stock_items", Description: "List low-stock POS branch inventory items.", Module: "pos", Scope: "read"},
	{Name: "pos.refund_order", Description: "Refund a POS order.", Module: "pos", Scope: "write", RequiresApproval: true},
	{Name: "store.list_orders", Description: "List ecommerce store orders.", Module: "store", Scope: "read"},
	{Name: "store.get_order", Description: "Get an ecommerce store order with line items.", Module: "store", Scope: "read"},
	{Name: "store.list_products", Description: "List ecommerce store products.", Module: "store", Scope: "read"},
	{Name: "store.list_categories", Description: "List ecommerce store categories.", Module: "store", Scope: "read"},
	{Name: "support.list_tickets", Description: "List customer support tickets.", Module: "support", Scope: "read"},
	{Name: "support.get_ticket", Description: "Get a support ticket with its messages.", Module: "support", Scope: "read"},
	{Name: "support.list_inbox", Description: "List active support inbox items.", Module: "support", Scope: "read"},
	{Name: "support.get_unread_count", Description: "Get the count of open support tickets.", Module: "support", Scope: "read"},
	{Name: "support.list_customers", Description: "List support customers.", Module: "support", Scope: "read"},
	{Name: "support.get_customer", Description: "Get a support customer summary.", Module: "support", Scope: "read"},
	{Name: "support.list_canned_responses", Description: "List reusable support canned responses.", Module: "support", Scope: "read"},
	{Name: "support.list_articles", Description: "List support knowledge-base articles.", Module: "support", Scope: "read"},
	{Name: "erp.list_products", Description: "List inventory products.", Module: "erp", Scope: "read"},
	{Name: "erp.set_stock", Description: "Set product stock quantity.", Module: "erp", Scope: "write", RequiresApproval: true},
	{Name: "erp.list_categories", Description: "List ERP product categories.", Module: "erp", Scope: "read"},
	{Name: "erp.create_product", Description: "Create an ERP product.", Module: "erp", Scope: "write", RequiresApproval: true},
	{Name: "erp.create_category", Description: "Create an ERP product category.", Module: "erp", Scope: "write", RequiresApproval: true},
	{Name: "erp.list_brands", Description: "List ERP product brands.", Module: "erp", Scope: "read"},
	{Name: "erp.create_brand", Description: "Create an ERP product brand.", Module: "erp", Scope: "write", RequiresApproval: true},
	{Name: "erp.list_suppliers", Description: "List ERP suppliers.", Module: "erp", Scope: "read"},
	{Name: "erp.create_supplier", Description: "Create an ERP supplier.", Module: "erp", Scope: "write", RequiresApproval: true},
	{Name: "erp.list_warehouses", Description: "List ERP warehouses.", Module: "erp", Scope: "read"},
	{Name: "erp.create_warehouse", Description: "Create an ERP warehouse.", Module: "erp", Scope: "write", RequiresApproval: true},
	{Name: "erp.list_purchase_orders", Description: "List ERP purchase orders.", Module: "erp", Scope: "read"},
	{Name: "erp.create_purchase_order", Description: "Create an ERP purchase order.", Module: "erp", Scope: "write", RequiresApproval: true},
	{Name: "erp.approve_purchase_order", Description: "Approve an ERP purchase order.", Module: "erp", Scope: "write", RequiresApproval: true},
	{Name: "erp.receive_purchase_order", Description: "Receive an ERP purchase order into inventory.", Module: "erp", Scope: "write", RequiresApproval: true},
	{Name: "erp.list_stock_transfers", Description: "List ERP stock transfers.", Module: "erp", Scope: "read"},
	{Name: "erp.create_stock_transfer", Description: "Create an ERP stock transfer between warehouses.", Module: "erp", Scope: "write", RequiresApproval: true},
	{Name: "erp.receive_stock_transfer", Description: "Receive an ERP stock transfer and move inventory.", Module: "erp", Scope: "write", RequiresApproval: true},
	{Name: "hr.list_employees", Description: "List employees.", Module: "hr", Scope: "read"},
	{Name: "hr.list_locations", Description: "List HR locations.", Module: "hr", Scope: "read"},
	{Name: "hr.list_departments", Description: "List HR departments.", Module: "hr", Scope: "read"},
	{Name: "hr.list_work_schedules", Description: "List HR work schedules.", Module: "hr", Scope: "read"},
	{Name: "hr.list_roles", Description: "List HR roles and permissions.", Module: "hr", Scope: "read"},
	{Name: "hr.get_setup_status", Description: "Get HR setup completion status.", Module: "hr", Scope: "read"},
	{Name: "hr.get_config", Description: "Get HR company configuration.", Module: "hr", Scope: "read"},
	{Name: "hr.list_reviews", Description: "List HR performance reviews.", Module: "hr", Scope: "read"},
	{Name: "hr.list_leave_types", Description: "List HR leave types and accrual settings.", Module: "hr", Scope: "read"},
	{Name: "hr.list_leave_requests", Description: "List HR leave requests and approval status.", Module: "hr", Scope: "read"},
	{Name: "hr.get_leave_balances", Description: "Get leave balances for an employee and year.", Module: "hr", Scope: "read"},
	{Name: "hr.get_leave_calendar", Description: "Get approved or pending leave on the HR calendar.", Module: "hr", Scope: "read"},
	{Name: "hr.list_overtime_requests", Description: "List overtime requests and approval status.", Module: "hr", Scope: "read"},
	{Name: "hr.get_overtime_config", Description: "Get HR overtime rules and auto-detection settings.", Module: "hr", Scope: "read"},
	{Name: "hr.get_payroll_config", Description: "Get payroll period and overtime configuration.", Module: "hr", Scope: "read"},
	{Name: "hr.list_payroll_runs", Description: "List payroll runs and totals.", Module: "hr", Scope: "read"},
	{Name: "hr.get_payroll_run", Description: "Get a payroll run with employee entries.", Module: "hr", Scope: "read"},
	{Name: "hr.list_time_entries", Description: "List employee time entries for a period.", Module: "hr", Scope: "read"},
	{Name: "hr.list_timesheets", Description: "List employee timesheets.", Module: "hr", Scope: "read"},
	{Name: "hr.get_timesheet", Description: "Get a timesheet by ID or employee period.", Module: "hr", Scope: "read"},
	{Name: "hr.list_attendance", Description: "List attendance records for an employee.", Module: "hr", Scope: "read"},
	{Name: "hr.get_location_attendance", Description: "Get branch attendance records for a day.", Module: "hr", Scope: "read"},
	{Name: "hr.create_employee", Description: "Create an HR employee record.", Module: "hr", Scope: "write", RequiresApproval: true},
	{Name: "flow.list_tasks", Description: "List Flow tasks.", Module: "flow", Scope: "read"},
	{Name: "flow.get_insights", Description: "Get Flow task velocity, completion, overdue, status, and priority metrics.", Module: "flow", Scope: "read"},
	{Name: "flow.get_team_workload", Description: "Get Flow workload distribution by team member.", Module: "flow", Scope: "read"},
	{Name: "flow.get_roadmap", Description: "Get Flow roadmap projects and milestone progress.", Module: "flow", Scope: "read"},
	{Name: "flow.create_task", Description: "Create a Flow task.", Module: "flow", Scope: "write", RequiresApproval: true},
	{Name: "finance.list_invoices", Description: "List finance invoices.", Module: "finance", Scope: "read"},
	{Name: "finance.get_balance_sheet", Description: "Get a finance balance sheet summary.", Module: "finance", Scope: "read"},
	{Name: "finance.get_cash_flow_report", Description: "Get a finance cash-flow report for month, quarter, or year.", Module: "finance", Scope: "read"},
	{Name: "finance.list_expenses", Description: "List finance expenses.", Module: "finance", Scope: "read"},
	{Name: "finance.list_expense_categories", Description: "List finance expense categories.", Module: "finance", Scope: "read"},
	{Name: "finance.record_expense", Description: "Record a business expense.", Module: "finance", Scope: "write", RequiresApproval: true},
	{Name: "finance.list_bills", Description: "List supplier bills.", Module: "finance", Scope: "read"},
	{Name: "finance.record_bill", Description: "Record a supplier bill.", Module: "finance", Scope: "write", RequiresApproval: true},
	{Name: "finance.create_quote", Description: "Create a customer quote.", Module: "finance", Scope: "write", RequiresApproval: true},
	{Name: "finance.create_invoice", Description: "Create a customer invoice.", Module: "finance", Scope: "write", RequiresApproval: true},
	{Name: "finance.record_payment", Description: "Record a customer payment against an invoice.", Module: "finance", Scope: "write", RequiresApproval: true},
	{Name: "finance.send_invoice", Description: "Send a customer invoice and mark it sent.", Module: "finance", Scope: "write", RequiresApproval: true},
	{Name: "booking.list_appointments", Description: "List booking appointments, optionally today's appointments.", Module: "booking", Scope: "read"},
	{Name: "booking.get_dashboard", Description: "Get booking dashboard counts, cancellation rate, and revenue.", Module: "booking", Scope: "read"},
	{Name: "booking.get_calendar", Description: "Get booking calendar appointments over a date range.", Module: "booking", Scope: "read"},
	{Name: "booking.list_services", Description: "List bookable services with duration and price.", Module: "booking", Scope: "read"},
	{Name: "booking.list_resources", Description: "List bookable resources, rooms, and capacity.", Module: "booking", Scope: "read"},
	{Name: "booking.list_waitlist", Description: "List booking waitlist entries.", Module: "booking", Scope: "read"},
	{Name: "booking.list_reminders", Description: "List booking reminders and delivery state.", Module: "booking", Scope: "read"},
	{Name: "booking.create_appointment", Description: "Create a booking appointment.", Module: "booking", Scope: "write", RequiresApproval: true},
	{Name: "pos.get_day_summary", Description: "Get POS sales, tax, tips, refunds, orders, and payments for a day.", Module: "pos", Scope: "read"},
}

var promptExecutableTools = map[string]struct{}{
	"business.snapshot":                 {},
	"business.get_kpis":                 {},
	"insights.get_dashboard":            {},
	"insights.get_revenue_series":       {},
	"insights.get_revenue_comparison":   {},
	"insights.get_revenue_by_app":       {},
	"insights.get_customer_kpis":        {},
	"insights.list_top_customers":       {},
	"insights.get_customer_churn":       {},
	"insights.get_ltv_distribution":     {},
	"insights.get_product_kpis":         {},
	"insights.list_bestsellers":         {},
	"insights.list_slow_movers":         {},
	"insights.get_sales_by_category":    {},
	"insights.get_margin_distribution":  {},
	"insights.get_cash_flow":            {},
	"insights.get_expenses_by_category": {},
	"insights.get_invoice_aging":        {},
	"insights.get_operations_kpis":      {},
	"insights.get_headcount":            {},
	"insights.get_shift_metrics":        {},
	"insights.get_time_off":             {},
	"insights.get_utilization":          {},
	"insights.get_settings":             {},
	"booking.list_appointments":         {},
	"booking.create_appointment":        {},
	"crm.list_customers":                {},
	"crm.create_customer":               {},
	"crm.list_campaigns":                {},
	"crm.get_campaign":                  {},
	"crm.list_segments":                 {},
	"loyalty.get_config":                {},
	"loyalty.list_members":              {},
	"loyalty.list_rewards":              {},
	"loyalty.list_transactions":         {},
	"erp.list_products":                 {},
	"erp.set_stock":                     {},
	"erp.list_categories":               {},
	"erp.create_product":                {},
	"erp.create_category":               {},
	"erp.list_brands":                   {},
	"erp.create_brand":                  {},
	"erp.list_suppliers":                {},
	"erp.create_supplier":               {},
	"erp.list_warehouses":               {},
	"erp.create_warehouse":              {},
	"erp.list_purchase_orders":          {},
	"erp.create_purchase_order":         {},
	"erp.approve_purchase_order":        {},
	"erp.receive_purchase_order":        {},
	"erp.list_stock_transfers":          {},
	"erp.create_stock_transfer":         {},
	"erp.receive_stock_transfer":        {},
	"hr.list_employees":                 {},
	"hr.list_locations":                 {},
	"hr.list_departments":               {},
	"hr.list_work_schedules":            {},
	"hr.list_roles":                     {},
	"hr.get_setup_status":               {},
	"hr.get_config":                     {},
	"hr.list_reviews":                   {},
	"hr.list_leave_types":               {},
	"hr.list_leave_requests":            {},
	"hr.get_leave_balances":             {},
	"hr.get_leave_calendar":             {},
	"hr.list_overtime_requests":         {},
	"hr.get_overtime_config":            {},
	"hr.get_payroll_config":             {},
	"hr.list_payroll_runs":              {},
	"hr.get_payroll_run":                {},
	"hr.list_time_entries":              {},
	"hr.list_timesheets":                {},
	"hr.get_timesheet":                  {},
	"hr.list_attendance":                {},
	"hr.get_location_attendance":        {},
	"flow.list_tasks":                   {},
	"flow.get_insights":                 {},
	"flow.get_team_workload":            {},
	"flow.get_roadmap":                  {},
	"flow.create_task":                  {},
	"finance.list_invoices":             {},
	"finance.get_balance_sheet":         {},
	"finance.get_cash_flow_report":      {},
	"finance.list_expenses":             {},
	"finance.list_expense_categories":   {},
	"finance.record_expense":            {},
	"finance.list_bills":                {},
	"finance.record_bill":               {},
	"finance.create_quote":              {},
	"finance.create_invoice":            {},
	"finance.record_payment":            {},
	"finance.send_invoice":              {},
	"pos.list_orders":                   {},
	"pos.list_locations":                {},
	"pos.list_terminals":                {},
	"pos.list_users":                    {},
	"pos.list_floors":                   {},
	"pos.list_tables":                   {},
	"pos.list_menu_items":               {},
	"pos.list_modifiers":                {},
	"pos.list_sessions":                 {},
	"pos.list_cash_movements":           {},
	"pos.list_stock_movements":          {},
	"pos.list_branch_inventory":         {},
	"pos.list_low_stock_items":          {},
	"pos.refund_order":                  {},
	"pos.get_day_summary":               {},
	"store.list_orders":                 {},
	"store.get_order":                   {},
	"store.list_products":               {},
	"store.list_categories":             {},
	"support.list_tickets":              {},
	"support.get_ticket":                {},
	"support.list_inbox":                {},
	"support.get_unread_count":          {},
	"support.list_customers":            {},
	"support.get_customer":              {},
	"support.list_canned_responses":     {},
	"support.list_articles":             {},
	"hr.create_employee":                {},
	"booking.get_dashboard":             {},
	"booking.get_calendar":              {},
	"booking.list_services":             {},
	"booking.list_resources":            {},
	"booking.list_waitlist":             {},
	"booking.list_reminders":            {},
}

// ToolByName returns a registered tool definition by stable tool name.
func ToolByName(name string) (ToolDef, bool) {
	for _, tool := range CoreTools {
		if tool.Name == name {
			return tool, true
		}
	}
	return ToolDef{}, false
}

// ToolsForMode filters CoreTools by mode capability.
func ToolsForMode(m CaretMode) []ToolDef {
	if !m.AllowsTools() {
		return nil
	}
	if m == ModeAsk {
		out := make([]ToolDef, 0, len(CoreTools))
		for _, t := range CoreTools {
			if t.Scope == "read" {
				out = append(out, t)
			}
		}
		return out
	}
	return CoreTools
}

// PromptToolsForMode returns only the tools the model may explicitly emit in an
// answer stream. A tool stays hidden until the stream or approval path has a
// server executor for it.
func PromptToolsForMode(m CaretMode) []ToolDef {
	if m != ModeAgent {
		return nil
	}
	out := make([]ToolDef, 0, len(CoreTools))
	for _, t := range CoreTools {
		if promptToolCanExecute(t.Name) {
			out = append(out, t)
		}
	}
	return out
}

func promptToolCanExecute(name string) bool {
	_, ok := promptExecutableTools[name]
	return ok
}
