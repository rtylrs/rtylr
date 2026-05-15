package handlers

// ToolDef describes a callable tool exposed to the LLM. The Caret echo
// handler renders these to JSON schema in the system prompt so the model
// knows what's available. Actual execution is dispatched per call site
// (HTTP, MCP, etc) — this module is metadata-only.
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
	{Name: "crm.list_customers", Description: "List CRM customers.", Module: "crm", Scope: "read"},
	{Name: "crm.create_customer", Description: "Create a CRM customer.", Module: "crm", Scope: "write", RequiresApproval: true},
	{Name: "pos.list_orders", Description: "List recent POS orders.", Module: "pos", Scope: "read"},
	{Name: "pos.refund_order", Description: "Refund a POS order.", Module: "pos", Scope: "write", RequiresApproval: true},
	{Name: "erp.list_products", Description: "List inventory products.", Module: "erp", Scope: "read"},
	{Name: "erp.set_stock", Description: "Set product stock quantity.", Module: "erp", Scope: "write", RequiresApproval: true},
	{Name: "hr.list_employees", Description: "List employees.", Module: "hr", Scope: "read"},
	{Name: "flow.list_tasks", Description: "List Flow tasks.", Module: "flow", Scope: "read"},
	{Name: "flow.create_task", Description: "Create a Flow task.", Module: "flow", Scope: "write", RequiresApproval: true},
	{Name: "finance.list_invoices", Description: "List finance invoices.", Module: "finance", Scope: "read"},
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
