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
	{Name: "analytics.query", Description: "Run a tenant-scoped, read-only analytics query over allowlisted business views.", Module: "analytics", Scope: "read"},
	{Name: "crm.list_customers", Description: "List CRM customers.", Module: "crm", Scope: "read"},
	{Name: "crm.create_customer", Description: "Create a CRM customer.", Module: "crm", Scope: "write", RequiresApproval: true},
	{Name: "pos.list_orders", Description: "List recent POS orders.", Module: "pos", Scope: "read"},
	{Name: "pos.refund_order", Description: "Refund a POS order.", Module: "pos", Scope: "write", RequiresApproval: true},
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
	{Name: "erp.create_stock_transfer", Description: "Create an ERP stock transfer between warehouses.", Module: "erp", Scope: "write", RequiresApproval: true},
	{Name: "erp.receive_stock_transfer", Description: "Receive an ERP stock transfer and move inventory.", Module: "erp", Scope: "write", RequiresApproval: true},
	{Name: "hr.list_employees", Description: "List employees.", Module: "hr", Scope: "read"},
	{Name: "hr.create_employee", Description: "Create an HR employee record.", Module: "hr", Scope: "write", RequiresApproval: true},
	{Name: "flow.list_tasks", Description: "List Flow tasks.", Module: "flow", Scope: "read"},
	{Name: "flow.create_task", Description: "Create a Flow task.", Module: "flow", Scope: "write", RequiresApproval: true},
	{Name: "finance.list_invoices", Description: "List finance invoices.", Module: "finance", Scope: "read"},
	{Name: "finance.list_expenses", Description: "List finance expenses.", Module: "finance", Scope: "read"},
	{Name: "finance.list_expense_categories", Description: "List finance expense categories.", Module: "finance", Scope: "read"},
	{Name: "finance.record_expense", Description: "Record a business expense.", Module: "finance", Scope: "write", RequiresApproval: true},
	{Name: "finance.list_bills", Description: "List supplier bills.", Module: "finance", Scope: "read"},
	{Name: "finance.record_bill", Description: "Record a supplier bill.", Module: "finance", Scope: "write", RequiresApproval: true},
	{Name: "finance.create_quote", Description: "Create a customer quote.", Module: "finance", Scope: "write", RequiresApproval: true},
	{Name: "finance.create_invoice", Description: "Create a customer invoice.", Module: "finance", Scope: "write", RequiresApproval: true},
	{Name: "finance.record_payment", Description: "Record a customer payment against an invoice.", Module: "finance", Scope: "write", RequiresApproval: true},
	{Name: "finance.send_invoice", Description: "Send a customer invoice and mark it sent.", Module: "finance", Scope: "write", RequiresApproval: true},
	{Name: "booking.create_appointment", Description: "Create a booking appointment.", Module: "booking", Scope: "write", RequiresApproval: true},
}

var promptExecutableTools = map[string]struct{}{
	"booking.create_appointment":      {},
	"crm.list_customers":              {},
	"crm.create_customer":             {},
	"erp.list_products":               {},
	"erp.set_stock":                   {},
	"erp.list_categories":             {},
	"erp.create_product":              {},
	"erp.create_category":             {},
	"erp.list_brands":                 {},
	"erp.create_brand":                {},
	"erp.list_suppliers":              {},
	"erp.create_supplier":             {},
	"erp.list_warehouses":             {},
	"erp.create_warehouse":            {},
	"erp.list_purchase_orders":        {},
	"erp.create_purchase_order":       {},
	"erp.approve_purchase_order":      {},
	"erp.receive_purchase_order":      {},
	"erp.create_stock_transfer":       {},
	"erp.receive_stock_transfer":      {},
	"hr.list_employees":               {},
	"flow.list_tasks":                 {},
	"flow.create_task":                {},
	"finance.list_invoices":           {},
	"finance.list_expenses":           {},
	"finance.list_expense_categories": {},
	"finance.record_expense":          {},
	"finance.list_bills":              {},
	"finance.record_bill":             {},
	"finance.create_quote":            {},
	"finance.create_invoice":          {},
	"finance.record_payment":          {},
	"finance.send_invoice":            {},
	"pos.list_orders":                 {},
	"pos.refund_order":                {},
	"hr.create_employee":              {},
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
