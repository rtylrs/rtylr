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
	{Name: "erp.create_supplier", Description: "Create an ERP supplier.", Module: "erp", Scope: "write", RequiresApproval: true},
	{Name: "hr.list_employees", Description: "List employees.", Module: "hr", Scope: "read"},
	{Name: "hr.create_employee", Description: "Create an HR employee record.", Module: "hr", Scope: "write", RequiresApproval: true},
	{Name: "flow.list_tasks", Description: "List Flow tasks.", Module: "flow", Scope: "read"},
	{Name: "flow.create_task", Description: "Create a Flow task.", Module: "flow", Scope: "write", RequiresApproval: true},
	{Name: "finance.list_invoices", Description: "List finance invoices.", Module: "finance", Scope: "read"},
	{Name: "finance.record_expense", Description: "Record a business expense.", Module: "finance", Scope: "write", RequiresApproval: true},
	{Name: "finance.record_bill", Description: "Record a supplier bill.", Module: "finance", Scope: "write", RequiresApproval: true},
	{Name: "finance.create_quote", Description: "Create a customer quote.", Module: "finance", Scope: "write", RequiresApproval: true},
}

var promptExecutableWriteTools = map[string]struct{}{
	"crm.create_customer":    {},
	"erp.set_stock":          {},
	"erp.create_supplier":    {},
	"flow.create_task":       {},
	"finance.record_expense": {},
	"finance.record_bill":    {},
	"finance.create_quote":   {},
	"pos.refund_order":       {},
	"hr.create_employee":     {},
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

// PromptToolsForMode returns only the tools the model may explicitly propose in
// an answer stream. Read tools are server-grounded before the prompt, and write
// tools stay hidden until an approval executor exists.
func PromptToolsForMode(m CaretMode) []ToolDef {
	if m != ModeAgent {
		return nil
	}
	out := make([]ToolDef, 0, len(CoreTools))
	for _, t := range CoreTools {
		if t.Scope == "write" && promptToolCanExecute(t.Name) {
			out = append(out, t)
		}
	}
	return out
}

func promptToolCanExecute(name string) bool {
	_, ok := promptExecutableWriteTools[name]
	return ok
}
