package application

import (
	"fmt"
	"strings"
)

// ToolDef describes a callable tool exposed to Caret. The registry is used in
// prompts and stream events; execution starts with server-owned read-only tools
// and can grow toward approval-backed write tools.
type ToolDef struct {
	Name             string `json:"name"`
	Description      string `json:"description"`
	Module           string `json:"module"`
	Scope            string `json:"scope"` // "read" or "write"
	RequiresApproval bool   `json:"requires_approval"`
	InputHint        string `json:"input_hint,omitempty"`
	OutputHint       string `json:"output_hint,omitempty"`
	Freshness        string `json:"freshness,omitempty"`
	SideEffect       string `json:"side_effect,omitempty"`
	ApprovalReason   string `json:"approval_reason,omitempty"`
	// Personal writes change only the caller's own view of the Portal, show
	// at once and are undone the same way, so they run without approval.
	Personal bool `json:"personal,omitempty"`
}

// CoreTools is every tool Caret can call: the business seed registry plus
// the general tools (web, reference data, utilities) that need no business
// data.
var CoreTools = append(append([]ToolDef(nil), seedTools...), generalToolDefs...)

// seedTools is the business registry. The long-term plan is to source this
// list from each module's MCP descriptor.
var seedTools = []ToolDef{
	{Name: "web.search", Description: "Search the public web or the news for current information. Returns results with snippets and source links (and publish dates for news). Query must contain public information only.", Module: "web", Scope: "read", InputHint: `{"query":"public search question","news":false,"recency":"optional day|week|month|year"}`},
	{Name: "tools.search", Description: "Find more tools by describing what you need, e.g. \"list team members\" or \"unpaid invoices\". Covers HR and the team, CRM customers, POS and store orders, products and stock, finance, bookings, support, loyalty, Flow tasks, and insights. Returns matching tools with their inputs; call one next.", Module: "tools", Scope: "read", InputHint: `{"query":"what you need"}`},
	{Name: "agents.run", Description: "Run up to four sub-agents in parallel, one per independent part of a bigger task (e.g. research each competitor, compare each option, check each question). Each works on its own with read-only tools (web search and pages, reference data, business reads) and returns a report with sources. Use it when parts don't depend on each other; give each a self-contained task.", Module: "agents", Scope: "read", InputHint: `{"tasks":[{"task":"self-contained instruction","context":"optional details it needs"}],"context":"optional background shared by all tasks"}`, OutputHint: "One report per task (status, findings, steps taken) and the combined web sources.", Freshness: "live at request time", SideEffect: "read-only; sub-agents can't change anything"},
	{Name: "account.profile", Description: "Read the authenticated person's name and current business identity.", Module: "account", Scope: "read", InputHint: `{}`},
	{Name: "utilities.calculate", Description: "Evaluate arithmetic exactly using numbers supplied by the user or verified results. Supports + - * / parentheses, percentages (\"15% of 200\"), pi, and sqrt, pow, round(x, places), min, max, abs, floor, ceil, ln, log10, exp.", Module: "utilities", Scope: "read", InputHint: `{"expression":"round((120 * 0.2) + pow(1.05, 3), 2)"}`},
	{Name: "utilities.time", Description: "Get the current date and time in an IANA timezone; with no timezone, in the person's own.", Module: "utilities", Scope: "read", InputHint: `{"timezone":"optional, e.g. Asia/Beirut"}`},
	{Name: "business.find", Description: "Find a person, customer, product, supplier, invoice, ticket, task or appointment by name across every app at once (\"Rana\", \"Acme\", \"latte\"). Use it when you don't know which app holds the record; it returns matches with the tool to open each one.", Module: "business", Scope: "read", InputHint: `{"query":"name or word"}`, OutputHint: "Matching records with kind, name, id and key fields, and the get tool for details.", Freshness: "live at request time", SideEffect: "read-only"},
	{Name: "business.snapshot", Description: "Build a tenant-scoped owner snapshot across sales, finance, inventory, CRM, HR, and Flow.", Module: "business", Scope: "read"},
	{Name: "business.get_kpis", Description: "Get headline business KPIs across revenue, orders, staff, customers, stock, and apps.", Module: "business", Scope: "read"},
	{Name: "memory.list", Description: "List what Caret remembers about the person and their business, with ids. Check it before proposing a memory, when asked what Caret knows about them, or to find a memory to forget.", Module: "memory", Scope: "read", InputHint: `{}`, OutputHint: "Each memory's id, type, text and last update.", Freshness: "live", SideEffect: "read-only"},
	{Name: "memory.forget", Description: "Forget one saved memory by id (from memory.list) when the person asks Caret to stop using it or it is wrong.", Module: "memory", Scope: "write", RequiresApproval: true, InputHint: `{"id":123}`, OutputHint: "The forgotten memory's id.", Freshness: "persistent after approval", SideEffect: "deletes one saved memory", ApprovalReason: "This removes something Caret remembers."},
	{Name: "chats.search", Description: "Search the person's past Caret chats in this business for a topic, decision, number or name they discussed before (\"what did we decide about pricing\", \"the supplier I mentioned last week\").", Module: "chats", Scope: "read", InputHint: `{"query":"words to look for","limit":10}`, OutputHint: "Matching excerpts with chat id, chat title, who said it and the date.", Freshness: "live", SideEffect: "read-only"},
	{Name: "chats.read", Description: "Read the latest messages of one past chat found with chats.search.", Module: "chats", Scope: "read", InputHint: `{"chat_id":"id from chats.search","limit":20}`, OutputHint: "The chat title and its recent messages.", Freshness: "live", SideEffect: "read-only"},
	{Name: "buddy.routines", Description: "List the routines Caret Buddy runs for this person (recurring check-ins, reminders, reports), with ids, schedule, next run and whether each is on.", Module: "buddy", Scope: "read", InputHint: `{}`, OutputHint: "Each routine's id, name, instructions, every_minutes, enabled and next_run_at.", Freshness: "live", SideEffect: "read-only"},
	{Name: "buddy.create_routine", Description: "Set up a recurring routine: Caret Buddy runs the instructions on a schedule on its own and posts the result in the person's Buddy chat (e.g. a nightly gym check-in, a morning summary of new tickets, a weekly unpaid-invoices nudge).", Module: "buddy", Scope: "write", RequiresApproval: true, InputHint: `{"name":"short title","instructions":"what to do and what to say each time","every_minutes":1440,"first_run_at":"2026-10-31T20:30:00+03:00"}`, OutputHint: "The new routine with its id and next run.", Freshness: "persistent", SideEffect: "starts a recurring background job", ApprovalReason: "This sets up something that runs on its own."},
	{Name: "buddy.update_routine", Description: "Change one of the person's routines by id: pause or resume it (enabled), rename it, change its instructions, how often it runs, or when it runs next.", Module: "buddy", Scope: "write", RequiresApproval: true, InputHint: `{"id":12,"enabled":false,"name":"optional","instructions":"optional","every_minutes":1440,"next_run_at":"optional RFC3339 time"}`, OutputHint: "The routine after the change.", Freshness: "persistent", SideEffect: "changes a recurring background job", ApprovalReason: "This changes something that runs on its own."},
	{Name: "buddy.delete_routine", Description: "Delete one of the person's routines by id so it never runs again.", Module: "buddy", Scope: "write", RequiresApproval: true, InputHint: `{"id":12}`, OutputHint: "The deleted routine's id.", Freshness: "persistent", SideEffect: "deletes a recurring background job", ApprovalReason: "This removes a routine."},
	{Name: "portal.find_page", Description: "Find the rtylr Portal page for something the person wants to do or see (\"add team members\", \"unpaid invoices\", \"change my password\"). Returns matching pages with their path, what they are for, their main buttons, and a page_block to show the page as a card they can click.", Module: "portal", Scope: "read", InputHint: `{"query":"what they want to do, in English"}`, OutputHint: "Up to four pages with path, app, title, summary, tasks, buttons and page_block.", Freshness: "matches the current Portal", SideEffect: "read-only"},
	{Name: "portal.layout", Description: "Read how the person has arranged their Portal: the app bar (the icon column on the far left: Home, Insights, Work, CRM, Operations, Finance, HR, POS, Automations, and Caret pinned at the bottom) and the workspace sidebar (Home, Team, Groups, Permissions, Audit Log, Menus, QR Codes, Design, Extensions), with item ids, order and what is hidden.", Module: "portal", Scope: "read", InputHint: `{}`, OutputHint: "Each area with its items (id, label, group, position, hidden) and whether it differs from the default.", Freshness: "live", SideEffect: "read-only"},
	{Name: "portal.customize", Description: "Rearrange the person's own Portal: move items in the app bar (area rail) or the workspace sidebar (area workspace), hide them or show them again. Item ids: rail home, insights, flow (shown as Work), crm, erp (shown as Operations), finance, hr, pos, automations, caret (Caret is you); workspace dashboard, team, groups, roles, auditLog, menus, qrCodes, design, extensions. The rail has a top group and a bottom group; \"move yourself to the top\" is {\"area\":\"rail\",\"id\":\"caret\",\"group\":\"top\",\"position\":\"first\"}. Home can move but not be hidden. People call the whole left side \"the sidebar\": when they don't say which part, leave area out: each item goes to the area it is in, and one call can change both. Put every move of one request in a single call's moves list; they apply in order. The result's order field is what the person now sees. It changes only this person's view and applies at once.", Module: "portal", Scope: "write", Personal: true, InputHint: `{"area":"optional rail|workspace; leave out for the whole sidebar","moves":[{"id":"flow","group":"optional top|bottom","position":"first|top|last|end|before:<id>|after:<id>|2"}],"hide":["pos"],"show":["pos"]}`, OutputHint: "The area after the change: order (as the person sees it) and each item.", Freshness: "applies immediately", SideEffect: "changes only how this person's Portal is arranged"},
	{Name: "portal.restore_default", Description: "Put the person's Portal navigation back to the rtylr default. People call the whole left side \"the sidebar\", so \"reset the sidebar\" or \"restore the default\" means everything (all, or leave area out). Use rail or workspace only when they name the app bar or the Home page list.", Module: "portal", Scope: "write", Personal: true, InputHint: `{"area":"optional all|rail|workspace; leave out for everything"}`, OutputHint: "The restored areas.", Freshness: "applies immediately", SideEffect: "changes only how this person's Portal is arranged"},
	{Name: "memory.remember", Description: "Remember a user-approved business fact, preference, reporting default, or operating rule for future Caret answers.", Module: "memory", Scope: "write", RequiresApproval: true},
	{Name: "analytics.query", Description: "Run a tenant-scoped, read-only analytics query over allowlisted business views.", Module: "analytics", Scope: "read"},
	{Name: "documents.map_form_fields", Description: "Map values from a document (scanned form, invoice, onboarding sheet, intake note) onto form fields. Returns a per-field plan with fill, check, click, or skip decisions, values, confidence, and review flags.", Module: "documents", Scope: "read"},
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
	{Name: "insights.schedule_report", Description: "Schedule a recurring business insight or report. Requires template_id, cadence, channel, and recipient.", Module: "insights", Scope: "write", RequiresApproval: true},
	{Name: "crm.list_customers", Description: "List CRM customers.", Module: "crm", Scope: "read"},
	{Name: "crm.get_customer", Description: "Get a CRM customer profile by ID, including contact, loyalty, spend, visit, tag, and note details.", Module: "crm", Scope: "read"},
	{Name: "crm.create_customer", Description: "Create a CRM customer.", Module: "crm", Scope: "write", RequiresApproval: true},
	{Name: "crm.list_campaigns", Description: "List CRM marketing campaigns and engagement counts.", Module: "crm", Scope: "read"},
	{Name: "crm.get_campaign", Description: "Get a CRM marketing campaign by ID.", Module: "crm", Scope: "read"},
	{Name: "crm.list_segments", Description: "List CRM customer segments.", Module: "crm", Scope: "read"},
	{Name: "loyalty.get_config", Description: "Get loyalty program configuration.", Module: "loyalty", Scope: "read"},
	{Name: "loyalty.list_members", Description: "List loyalty members by points and visits.", Module: "loyalty", Scope: "read"},
	{Name: "loyalty.list_rewards", Description: "List loyalty rewards.", Module: "loyalty", Scope: "read"},
	{Name: "loyalty.list_transactions", Description: "List loyalty point transactions.", Module: "loyalty", Scope: "read"},
	{Name: "pos.list_orders", Description: "List recent POS orders.", Module: "pos", Scope: "read"},
	{Name: "pos.get_order", Description: "Get a POS order by ID, including payment details, notes, and line items.", Module: "pos", Scope: "read"},
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
	{Name: "store.get_order", Description: "Get an ecommerce store order by ID, with line items.", Module: "store", Scope: "read"},
	{Name: "store.list_products", Description: "List ecommerce store products.", Module: "store", Scope: "read"},
	{Name: "store.get_product", Description: "Get an ecommerce store product by ID, inventory product ID, slug, or title.", Module: "store", Scope: "read"},
	{Name: "store.list_categories", Description: "List ecommerce store categories.", Module: "store", Scope: "read"},
	{Name: "support.list_tickets", Description: "List customer support tickets.", Module: "support", Scope: "read"},
	{Name: "support.get_ticket", Description: "Get a support ticket by ID, with its messages.", Module: "support", Scope: "read"},
	{Name: "support.list_inbox", Description: "List active support inbox items.", Module: "support", Scope: "read"},
	{Name: "support.get_unread_count", Description: "Get the count of open support tickets.", Module: "support", Scope: "read"},
	{Name: "support.list_customers", Description: "List support customers.", Module: "support", Scope: "read"},
	{Name: "support.get_customer", Description: "Get a support customer summary by ID.", Module: "support", Scope: "read"},
	{Name: "support.list_canned_responses", Description: "List reusable support canned responses.", Module: "support", Scope: "read"},
	{Name: "support.list_articles", Description: "List support knowledge-base articles.", Module: "support", Scope: "read"},
	{Name: "support.get_article", Description: "Get a support knowledge-base article by ID, slug, or title.", Module: "support", Scope: "read"},
	{Name: "erp.list_products", Description: "List inventory products.", Module: "erp", Scope: "read"},
	{Name: "erp.get_product", Description: "Get an inventory product by ID or barcode, including stock, price, margin, reorder, and attribute details.", Module: "erp", Scope: "read"},
	{Name: "erp.list_reorder_suggestions", Description: "List inventory products at or below their reorder point with suggested replenishment quantities.", Module: "erp", Scope: "read"},
	{Name: "erp.set_stock", Description: "Set product stock quantity.", Module: "erp", Scope: "write", RequiresApproval: true},
	{Name: "erp.list_categories", Description: "List ERP product categories.", Module: "erp", Scope: "read"},
	{Name: "erp.create_product", Description: "Add a product to inventory from the user-provided product details. Requires approval before anything changes.", Module: "erp", Scope: "write", RequiresApproval: true},
	{Name: "erp.update_product", Description: "Update selected fields on an ERP product without changing fields that were not requested.", Module: "erp", Scope: "write", RequiresApproval: true},
	{Name: "erp.delete_product", Description: "Delete an ERP product.", Module: "erp", Scope: "write", RequiresApproval: true},
	{Name: "erp.create_category", Description: "Create an ERP product category.", Module: "erp", Scope: "write", RequiresApproval: true},
	{Name: "erp.list_brands", Description: "List ERP product brands.", Module: "erp", Scope: "read"},
	{Name: "erp.create_brand", Description: "Create an ERP product brand.", Module: "erp", Scope: "write", RequiresApproval: true},
	{Name: "erp.list_suppliers", Description: "List ERP suppliers.", Module: "erp", Scope: "read"},
	{Name: "erp.get_supplier", Description: "Get an ERP supplier by ID, including contact, address, tax, payment terms, notes, and active status.", Module: "erp", Scope: "read"},
	{Name: "erp.create_supplier", Description: "Create an ERP supplier.", Module: "erp", Scope: "write", RequiresApproval: true},
	{Name: "erp.list_warehouses", Description: "List ERP warehouses.", Module: "erp", Scope: "read"},
	{Name: "erp.get_warehouse", Description: "Get an ERP warehouse by ID, including address, manager, capacity, default, and active status.", Module: "erp", Scope: "read"},
	{Name: "erp.create_warehouse", Description: "Create an ERP warehouse.", Module: "erp", Scope: "write", RequiresApproval: true},
	{Name: "erp.list_purchase_orders", Description: "List ERP purchase orders.", Module: "erp", Scope: "read"},
	{Name: "erp.get_purchase_order", Description: "Get an ERP purchase order by ID, including supplier, status, totals, delivery, received time, and line items.", Module: "erp", Scope: "read"},
	{Name: "erp.create_purchase_order", Description: "Create an ERP purchase order.", Module: "erp", Scope: "write", RequiresApproval: true},
	{Name: "erp.approve_purchase_order", Description: "Approve an ERP purchase order.", Module: "erp", Scope: "write", RequiresApproval: true},
	{Name: "erp.receive_purchase_order", Description: "Receive an ERP purchase order into inventory.", Module: "erp", Scope: "write", RequiresApproval: true},
	{Name: "erp.list_stock_transfers", Description: "List ERP stock transfers.", Module: "erp", Scope: "read"},
	{Name: "erp.get_stock_transfer", Description: "Get an ERP stock transfer by ID, including source, destination, status, notes, and line items.", Module: "erp", Scope: "read"},
	{Name: "erp.create_stock_transfer", Description: "Create an ERP stock transfer between warehouses.", Module: "erp", Scope: "write", RequiresApproval: true},
	{Name: "erp.receive_stock_transfer", Description: "Receive an ERP stock transfer and move inventory.", Module: "erp", Scope: "write", RequiresApproval: true},
	{Name: "hr.list_employees", Description: "List employees.", Module: "hr", Scope: "read"},
	{Name: "hr.get_employee", Description: "Get an employee profile by ID, including role, contact, pay, schedule, emergency contact, and notes.", Module: "hr", Scope: "read"},
	{Name: "hr.list_locations", Description: "List HR locations.", Module: "hr", Scope: "read"},
	{Name: "hr.list_departments", Description: "List HR departments.", Module: "hr", Scope: "read"},
	{Name: "hr.list_work_schedules", Description: "List HR work schedules.", Module: "hr", Scope: "read"},
	{Name: "hr.list_roles", Description: "List HR roles and permissions.", Module: "hr", Scope: "read"},
	{Name: "hr.get_setup_status", Description: "Get HR setup completion status.", Module: "hr", Scope: "read"},
	{Name: "hr.get_config", Description: "Get HR company configuration.", Module: "hr", Scope: "read"},
	{Name: "hr.list_reviews", Description: "List HR performance reviews.", Module: "hr", Scope: "read"},
	{Name: "hr.list_leave_types", Description: "List HR leave types and accrual settings.", Module: "hr", Scope: "read"},
	{Name: "hr.list_leave_requests", Description: "List HR leave requests and approval status.", Module: "hr", Scope: "read"},
	{Name: "hr.get_leave_balances", Description: "Get leave balances for an employee and year.", Module: "hr", Scope: "read", InputHint: `{"employee_id":"id from hr.list_employees","year":2026}`},
	{Name: "hr.get_leave_calendar", Description: "Get approved or pending leave on the HR calendar.", Module: "hr", Scope: "read"},
	{Name: "hr.list_overtime_requests", Description: "List overtime requests and approval status.", Module: "hr", Scope: "read"},
	{Name: "hr.get_overtime_config", Description: "Get HR overtime rules and auto-detection settings.", Module: "hr", Scope: "read"},
	{Name: "hr.get_payroll_config", Description: "Get payroll period and overtime configuration.", Module: "hr", Scope: "read"},
	{Name: "hr.list_payroll_runs", Description: "List payroll runs and totals.", Module: "hr", Scope: "read"},
	{Name: "hr.get_payroll_run", Description: "Get a payroll run by ID, with employee entries.", Module: "hr", Scope: "read"},
	{Name: "hr.list_time_entries", Description: "List employee time entries for a period.", Module: "hr", Scope: "read"},
	{Name: "hr.list_timesheets", Description: "List employee timesheets.", Module: "hr", Scope: "read"},
	{Name: "hr.get_timesheet", Description: "Get a timesheet by ID or employee period.", Module: "hr", Scope: "read"},
	{Name: "hr.list_attendance", Description: "List attendance records for an employee.", Module: "hr", Scope: "read"},
	{Name: "hr.get_location_attendance", Description: "Get branch attendance records for a day.", Module: "hr", Scope: "read", InputHint: `{"location_id":"id from hr.list_locations","date":"YYYY-MM-DD"}`},
	{Name: "hr.create_employee", Description: "Create an HR employee record.", Module: "hr", Scope: "write", RequiresApproval: true},
	{Name: "flow.list_tasks", Description: "List Flow tasks.", Module: "flow", Scope: "read"},
	{Name: "flow.get_task", Description: "Get a Flow task by ID, including subtasks, comments, and recent activity.", Module: "flow", Scope: "read"},
	{Name: "flow.get_insights", Description: "Get Flow task velocity, completion, overdue, status, and priority metrics.", Module: "flow", Scope: "read"},
	{Name: "flow.get_team_workload", Description: "Get Flow workload distribution by team member.", Module: "flow", Scope: "read"},
	{Name: "flow.get_roadmap", Description: "Get Flow roadmap projects and milestone progress.", Module: "flow", Scope: "read"},
	{Name: "flow.create_task", Description: "Create a Flow task.", Module: "flow", Scope: "write", RequiresApproval: true},
	{Name: "finance.list_invoices", Description: "List finance invoices.", Module: "finance", Scope: "read"},
	{Name: "finance.get_invoice", Description: "Get a finance invoice by ID, including customer, line item, payment, balance, and due-date details.", Module: "finance", Scope: "read"},
	{Name: "finance.get_balance_sheet", Description: "Get a finance balance sheet summary.", Module: "finance", Scope: "read"},
	{Name: "finance.get_cash_flow_report", Description: "Get a finance cash-flow report for month, quarter, or year.", Module: "finance", Scope: "read", InputHint: `{"period":"month|quarter|year"}`},
	{Name: "finance.list_expenses", Description: "List finance expenses.", Module: "finance", Scope: "read"},
	{Name: "finance.get_expense", Description: "Get a finance expense by ID, including category, amount, vendor, receipt, date, and approval status.", Module: "finance", Scope: "read"},
	{Name: "finance.list_expense_categories", Description: "List finance expense categories.", Module: "finance", Scope: "read"},
	{Name: "finance.list_bank_accounts", Description: "List bank accounts with balances and connection status.", Module: "finance", Scope: "read"},
	{Name: "finance.get_bank_account", Description: "Get a finance bank account by ID, including institution, type, balance, currency, connection, and sync details.", Module: "finance", Scope: "read"},
	{Name: "finance.get_bank_summary", Description: "Get aggregate bank balance and connected account count.", Module: "finance", Scope: "read"},
	{Name: "finance.list_bank_feed_transactions", Description: "List bank feed transactions and match state.", Module: "finance", Scope: "read"},
	{Name: "finance.get_reconciliation_overview", Description: "Get unmatched bank transactions, invoices, and reconciliation suggestions.", Module: "finance", Scope: "read"},
	{Name: "finance.list_journal_entries", Description: "List finance journal entries with debit and credit totals.", Module: "finance", Scope: "read"},
	{Name: "finance.get_tax_config", Description: "Get active tax configuration.", Module: "finance", Scope: "read"},
	{Name: "finance.get_pl_report", Description: "Get profit and loss totals.", Module: "finance", Scope: "read"},
	{Name: "finance.record_expense", Description: "Record a business expense.", Module: "finance", Scope: "write", RequiresApproval: true},
	{Name: "finance.list_bills", Description: "List supplier bills.", Module: "finance", Scope: "read"},
	{Name: "finance.record_bill", Description: "Record a supplier bill.", Module: "finance", Scope: "write", RequiresApproval: true},
	{Name: "finance.create_quote", Description: "Create a customer quote.", Module: "finance", Scope: "write", RequiresApproval: true},
	{Name: "finance.create_invoice", Description: "Create a customer invoice.", Module: "finance", Scope: "write", RequiresApproval: true},
	{Name: "finance.record_payment", Description: "Record a customer payment against an invoice.", Module: "finance", Scope: "write", RequiresApproval: true},
	{Name: "finance.send_invoice", Description: "Send a customer invoice and mark it sent.", Module: "finance", Scope: "write", RequiresApproval: true},
	{Name: "booking.list_appointments", Description: "List booking appointments, optionally today's appointments.", Module: "booking", Scope: "read"},
	{Name: "booking.get_appointment", Description: "Get a booking appointment by ID, including customer, timing, service, resource, staff, status, and notes.", Module: "booking", Scope: "read"},
	{Name: "booking.get_dashboard", Description: "Get booking dashboard counts, cancellation rate, and revenue.", Module: "booking", Scope: "read"},
	{Name: "booking.get_calendar", Description: "Get booking calendar appointments over a date range.", Module: "booking", Scope: "read"},
	{Name: "booking.list_services", Description: "List bookable services with duration and price.", Module: "booking", Scope: "read"},
	{Name: "booking.list_resources", Description: "List bookable resources, rooms, and capacity.", Module: "booking", Scope: "read"},
	{Name: "booking.list_waitlist", Description: "List booking waitlist entries.", Module: "booking", Scope: "read"},
	{Name: "booking.list_reminders", Description: "List booking reminders and delivery state.", Module: "booking", Scope: "read"},
	{Name: "booking.create_appointment", Description: "Create a booking appointment.", Module: "booking", Scope: "write", RequiresApproval: true},
	{Name: "pos.get_day_summary", Description: "Get POS sales, tax, tips, refunds, orders, and payments for a day.", Module: "pos", Scope: "read", InputHint: `{"date":"YYYY-MM-DD"}`},
}

var promptExecutableTools = map[string]struct{}{
	"web.search": {}, "tools.search": {}, "agents.run": {}, "account.profile": {}, "utilities.calculate": {}, "utilities.time": {},
	"business.snapshot":                   {},
	"business.find":                       {},
	"business.get_kpis":                   {},
	"memory.remember":                     {},
	"memory.list":                         {},
	"buddy.routines":                      {},
	"buddy.create_routine":                {},
	"buddy.update_routine":                {},
	"buddy.delete_routine":                {},
	"memory.forget":                       {},
	"portal.find_page":                    {},
	"portal.layout":                       {},
	"portal.customize":                    {},
	"portal.restore_default":              {},
	"chats.search":                        {},
	"chats.read":                          {},
	"analytics.query":                     {},
	"documents.map_form_fields":           {},
	"insights.get_dashboard":              {},
	"insights.get_revenue_series":         {},
	"insights.get_revenue_comparison":     {},
	"insights.get_revenue_by_app":         {},
	"insights.get_customer_kpis":          {},
	"insights.list_top_customers":         {},
	"insights.get_customer_churn":         {},
	"insights.get_ltv_distribution":       {},
	"insights.get_product_kpis":           {},
	"insights.list_bestsellers":           {},
	"insights.list_slow_movers":           {},
	"insights.get_sales_by_category":      {},
	"insights.get_margin_distribution":    {},
	"insights.get_cash_flow":              {},
	"insights.get_expenses_by_category":   {},
	"insights.get_invoice_aging":          {},
	"insights.get_operations_kpis":        {},
	"insights.get_headcount":              {},
	"insights.get_shift_metrics":          {},
	"insights.get_time_off":               {},
	"insights.get_utilization":            {},
	"insights.get_settings":               {},
	"insights.schedule_report":            {},
	"booking.list_appointments":           {},
	"booking.get_appointment":             {},
	"booking.create_appointment":          {},
	"crm.list_customers":                  {},
	"crm.get_customer":                    {},
	"crm.create_customer":                 {},
	"crm.list_campaigns":                  {},
	"crm.get_campaign":                    {},
	"crm.list_segments":                   {},
	"loyalty.get_config":                  {},
	"loyalty.list_members":                {},
	"loyalty.list_rewards":                {},
	"loyalty.list_transactions":           {},
	"erp.list_products":                   {},
	"erp.get_product":                     {},
	"erp.list_reorder_suggestions":        {},
	"erp.set_stock":                       {},
	"erp.list_categories":                 {},
	"erp.create_product":                  {},
	"erp.update_product":                  {},
	"erp.delete_product":                  {},
	"erp.create_category":                 {},
	"erp.list_brands":                     {},
	"erp.create_brand":                    {},
	"erp.list_suppliers":                  {},
	"erp.get_supplier":                    {},
	"erp.create_supplier":                 {},
	"erp.list_warehouses":                 {},
	"erp.get_warehouse":                   {},
	"erp.create_warehouse":                {},
	"erp.list_purchase_orders":            {},
	"erp.get_purchase_order":              {},
	"erp.create_purchase_order":           {},
	"erp.approve_purchase_order":          {},
	"erp.receive_purchase_order":          {},
	"erp.list_stock_transfers":            {},
	"erp.get_stock_transfer":              {},
	"erp.create_stock_transfer":           {},
	"erp.receive_stock_transfer":          {},
	"hr.list_employees":                   {},
	"hr.get_employee":                     {},
	"hr.list_locations":                   {},
	"hr.list_departments":                 {},
	"hr.list_work_schedules":              {},
	"hr.list_roles":                       {},
	"hr.get_setup_status":                 {},
	"hr.get_config":                       {},
	"hr.list_reviews":                     {},
	"hr.list_leave_types":                 {},
	"hr.list_leave_requests":              {},
	"hr.get_leave_balances":               {},
	"hr.get_leave_calendar":               {},
	"hr.list_overtime_requests":           {},
	"hr.get_overtime_config":              {},
	"hr.get_payroll_config":               {},
	"hr.list_payroll_runs":                {},
	"hr.get_payroll_run":                  {},
	"hr.list_time_entries":                {},
	"hr.list_timesheets":                  {},
	"hr.get_timesheet":                    {},
	"hr.list_attendance":                  {},
	"hr.get_location_attendance":          {},
	"flow.list_tasks":                     {},
	"flow.get_task":                       {},
	"flow.get_insights":                   {},
	"flow.get_team_workload":              {},
	"flow.get_roadmap":                    {},
	"flow.create_task":                    {},
	"finance.list_invoices":               {},
	"finance.get_invoice":                 {},
	"finance.get_balance_sheet":           {},
	"finance.get_cash_flow_report":        {},
	"finance.list_expenses":               {},
	"finance.get_expense":                 {},
	"finance.list_expense_categories":     {},
	"finance.list_bank_accounts":          {},
	"finance.get_bank_account":            {},
	"finance.get_bank_summary":            {},
	"finance.list_bank_feed_transactions": {},
	"finance.get_reconciliation_overview": {},
	"finance.list_journal_entries":        {},
	"finance.get_tax_config":              {},
	"finance.get_pl_report":               {},
	"finance.record_expense":              {},
	"finance.list_bills":                  {},
	"finance.record_bill":                 {},
	"finance.create_quote":                {},
	"finance.create_invoice":              {},
	"finance.record_payment":              {},
	"finance.send_invoice":                {},
	"pos.list_orders":                     {},
	"pos.get_order":                       {},
	"pos.list_locations":                  {},
	"pos.list_terminals":                  {},
	"pos.list_users":                      {},
	"pos.list_floors":                     {},
	"pos.list_tables":                     {},
	"pos.list_menu_items":                 {},
	"pos.list_modifiers":                  {},
	"pos.list_sessions":                   {},
	"pos.list_cash_movements":             {},
	"pos.list_stock_movements":            {},
	"pos.list_branch_inventory":           {},
	"pos.list_low_stock_items":            {},
	"pos.refund_order":                    {},
	"pos.get_day_summary":                 {},
	"store.list_orders":                   {},
	"store.get_order":                     {},
	"store.list_products":                 {},
	"store.get_product":                   {},
	"store.list_categories":               {},
	"support.list_tickets":                {},
	"support.get_ticket":                  {},
	"support.list_inbox":                  {},
	"support.get_unread_count":            {},
	"support.list_customers":              {},
	"support.get_customer":                {},
	"support.list_canned_responses":       {},
	"support.list_articles":               {},
	"support.get_article":                 {},
	"hr.create_employee":                  {},
	"booking.get_dashboard":               {},
	"booking.get_calendar":                {},
	"booking.list_services":               {},
	"booking.list_resources":              {},
	"booking.list_waitlist":               {},
	"booking.list_reminders":              {},
}

var coreToolIndex = buildCoreToolIndex()

func buildCoreToolIndex() map[string]ToolDef {
	index := make(map[string]ToolDef, len(CoreTools))
	for _, tool := range CoreTools {
		if strings.TrimSpace(tool.Name) == "" {
			continue
		}
		index[tool.Name] = tool
	}
	return index
}

// ValidateToolRegistry is called by startup checks and tests. A tool catalog
// is executable policy, not documentation: duplicate names, invalid scopes,
// and prompt entries without definitions make model behavior nondeterministic.
func ValidateToolRegistry() error {
	seen := make(map[string]struct{}, len(CoreTools))
	for _, tool := range CoreTools {
		name := strings.TrimSpace(tool.Name)
		if name == "" {
			return fmt.Errorf("tool name is empty")
		}
		if _, exists := seen[name]; exists {
			return fmt.Errorf("tool %q is registered more than once", name)
		}
		seen[name] = struct{}{}
		if tool.Scope != "read" && tool.Scope != "write" {
			return fmt.Errorf("tool %q has invalid scope %q", name, tool.Scope)
		}
		if strings.TrimSpace(tool.Module) == "" {
			return fmt.Errorf("tool %q has no module", name)
		}
		if tool.Scope == "write" && !tool.RequiresApproval && !tool.Personal {
			return fmt.Errorf("write tool %q must require approval", name)
		}
		if tool.Personal && (tool.Scope != "write" || tool.RequiresApproval || tool.Module != "portal") {
			return fmt.Errorf("personal tool %q must be a portal write that runs without approval", name)
		}
	}
	for name := range promptExecutableTools {
		if _, ok := seen[name]; !ok {
			return fmt.Errorf("prompt executable %q is not registered", name)
		}
	}
	for _, tool := range generalToolDefs {
		if _, ok := generalTools[tool.Name]; !ok {
			return fmt.Errorf("general tool %q has no handler", tool.Name)
		}
	}
	for name := range generalTools {
		if _, ok := seen[name]; !ok {
			return fmt.Errorf("general tool handler %q is not registered", name)
		}
	}
	return nil
}

// ToolByName returns a registered tool definition by stable tool name.
func ToolByName(name string) (ToolDef, bool) {
	tool, ok := coreToolIndex[strings.TrimSpace(name)]
	if !ok {
		return ToolDef{}, false
	}
	return enrichToolDefinition(tool), true
}

// ToolsForMode filters CoreTools by mode capability.
func ToolsForMode(m CaretMode) []ToolDef {
	out := make([]ToolDef, 0, len(CoreTools))
	for _, t := range CoreTools {
		if ToolAllowedForMode(m, t) {
			out = append(out, enrichToolDefinition(t))
		}
	}
	return out
}

// PromptToolsForMode returns only the tools the model may explicitly emit in an
// answer stream. A tool stays hidden until the stream or approval path has a
// server executor for it.
func PromptToolsForMode(m CaretMode) []ToolDef {
	out := make([]ToolDef, 0, len(CoreTools))
	for _, t := range CoreTools {
		if ToolAllowedForMode(m, t) && promptToolCanExecute(t.Name) {
			out = append(out, enrichToolDefinition(t))
		}
	}
	return out
}

func ToolAllowedForMode(m CaretMode, tool ToolDef) bool {
	if !m.AllowsTools() {
		return false
	}

	switch m {
	case ModeAgent:
		return true
	case ModeAsk, ModeAnalyst:
		return tool.Scope == "read"
	default:
		return false
	}
}

func promptToolCanExecute(name string) bool {
	if _, ok := generalTools[name]; ok {
		return true
	}
	_, ok := promptExecutableTools[name]
	return ok
}
