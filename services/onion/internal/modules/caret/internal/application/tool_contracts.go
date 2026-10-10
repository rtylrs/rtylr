package application

import "strings"

// Tool contracts are server-owned descriptions of the shape, freshness, and
// consequence of a tool call. They keep the model from guessing an input
// format and give the approval UI/API enough context to explain a proposal.
// The registry remains the source of capability membership; this file is the
// source of execution semantics.
type toolContract struct {
	Input          string
	Output         string
	Freshness      string
	SideEffect     string
	ApprovalReason string
}

func enrichToolDefinition(tool ToolDef) ToolDef {
	contract := contractForTool(tool)
	if strings.TrimSpace(tool.InputHint) == "" {
		tool.InputHint = contract.Input
	}
	if strings.TrimSpace(tool.OutputHint) == "" {
		tool.OutputHint = contract.Output
	}
	if strings.TrimSpace(tool.Freshness) == "" {
		tool.Freshness = contract.Freshness
	}
	if strings.TrimSpace(tool.SideEffect) == "" {
		tool.SideEffect = contract.SideEffect
	}
	if strings.TrimSpace(tool.ApprovalReason) == "" {
		tool.ApprovalReason = contract.ApprovalReason
	}
	return tool
}

func contractForTool(tool ToolDef) toolContract {
	if tool.Scope == "write" {
		return toolContract{
			Input:          writeInputHint(tool),
			Output:         "A confirmation object with the created or changed record and a stable operation identifier.",
			Freshness:      "live; execute only after explicit approval",
			SideEffect:     "writes business data or sends a business-side effect",
			ApprovalReason: "This action can change business data or communicate externally.",
		}
	}

	switch tool.Name {
	case "business.snapshot":
		return toolContract{
			Input:      `{"question":"what the owner wants to understand"}`,
			Output:     "A tenant-scoped evidence bundle with metric labels, rows, query specs, and freshness metadata.",
			Freshness:  "live at request time",
			SideEffect: "read-only; no business data is changed",
		}
	case "business.get_kpis":
		return toolContract{
			Input:      `{"period":"today|week|month"}`,
			Output:     "Headline tenant-scoped KPI values with the period used for each value.",
			Freshness:  "live at request time",
			SideEffect: "read-only; no business data is changed",
		}
	case "analytics.query":
		return toolContract{
			Input:      `{"question":"the user's business analytics question"}`,
			Output:     "Rows and a count from the authenticated tenant's allowlisted analytics view.",
			Freshness:  "live at request time",
			SideEffect: "read-only; no business data is changed",
		}
	case "documents.map_form_fields":
		return toolContract{
			Input:      `{"document":"text containing Label: value lines","fields":[{"id":"optional","label":"Phone number","role":"Edit|Checkbox|Button","value":"optional current value"}],"form":"optional form name","threshold":0.5,"allow_submit":false}` + " (limits: 8 KB document, 32 fields, 64 values, 1200 field-value pairs; threshold defaults to 0.5 when omitted or 0)",
			Output:     "A per-field plan with the chosen action (fill, check, click, or skip), the value to use, a probability, and a needs_review flag; fields are ordered for application (fills, then checks, then any approved submit click). It maps values that already exist in the supplied document and never invents data; plan values may contain the document's personal data and are returned only to the requesting user.",
			Freshness:  "computed from the supplied document and field list at request time",
			SideEffect: "read-only; no business data is changed",
		}
	case "erp.list_reorder_suggestions":
		return toolContract{
			Input:      `{}`,
			Output:     "Products at or below their configured reorder point, with current stock, reorder point, suggested quantity, and confidence.",
			Freshness:  "live at request time",
			SideEffect: "read-only; no business data is changed",
		}
	case "web.search":
		return toolContract{
			Output:     "Up to eight public results, each with title, URL, snippet and, for news, a publish date. Raw material to answer from, not an answer.",
			Freshness:  "live web",
			SideEffect: "read-only; sends the query to a public search engine",
		}
	case ToolSearchName:
		return toolContract{
			Output:     "Up to eight matching tools with their exact call contract; call one next.",
			Freshness:  "the tool catalog",
			SideEffect: "read-only",
		}
	case "utilities.calculate":
		return toolContract{Output: "The exact result of the expression.", Freshness: "computed", SideEffect: "read-only"}
	case "utilities.time":
		return toolContract{Output: "The timezone, local time and date.", Freshness: "live clock", SideEffect: "read-only"}
	case "account.profile":
		return toolContract{Output: "The signed-in person's name, role and current business.", Freshness: "live", SideEffect: "read-only"}
	case "memory.remember":
		return toolContract{
			Input:          `{"key":"stable preference name","value":"fact or preference to remember"}`,
			Output:         "A confirmation of the saved memory and its normalized key.",
			Freshness:      "persistent after approval",
			SideEffect:     "persists user-approved memory used in future conversations",
			ApprovalReason: "This changes what Caret will remember about the business or user.",
		}
	}

	return toolContract{
		Input:      genericInputHint(tool),
		Output:     "A tenant-scoped result matching the tool description; empty results are reported as empty, not inferred.",
		Freshness:  "live at request time",
		SideEffect: "read-only; no business data is changed",
	}
}

func genericInputHint(tool ToolDef) string {
	name := strings.ToLower(tool.Name)
	description := strings.ToLower(tool.Description)
	switch {
	case strings.Contains(name, ".get_") && strings.Contains(description, " by id"):
		return `{"id":"record id from a list or search tool"}`
	case strings.Contains(description, "date range") || strings.Contains(description, "over a period"):
		return `{"start_date":"YYYY-MM-DD","end_date":"YYYY-MM-DD"}`
	case strings.Contains(name, ".get_"):
		// Dashboards, settings and summaries need no identifier.
		return `{}`
	case strings.Contains(name, ".list_"):
		return `{"limit":number,"offset":number,"search":"optional filter","status":"optional filter"}`
	case strings.Contains(name, "report"), strings.Contains(name, "series"), strings.Contains(name, "comparison"):
		return `{"start_date":"YYYY-MM-DD","end_date":"YYYY-MM-DD","timezone":"IANA timezone"}`
	default:
		return `{"filters":{...},"limit":number}`
	}
}

func writeInputHint(tool ToolDef) string {
	name := strings.ToLower(tool.Name)
	switch {
	case name == "crm.create_customer":
		return `{"first_name":"optional","last_name":"optional","name":"optional full name","email":"optional","phone":"optional","notes":"optional"}`
	case name == "flow.create_task":
		return `{"title":"required task title","description":"optional","due_date":"YYYY-MM-DD","priority":"optional","assignee_id":"optional"}`
	case name == "finance.create_invoice":
		return `{"customer_id":"customer id","items":[{"description":"text","quantity":number,"unit_price":number}],"due_date":"YYYY-MM-DD"}`
	case name == "finance.record_expense":
		return `{"amount":number,"category_id":"category id","description":"reason","expense_date":"YYYY-MM-DD"}`
	case name == "erp.create_product":
		return `{"name":"required","sale_price":number,"cost_price":number,"stock_quantity":number,"sku":"optional","barcode":"optional","description":"optional","image_url":"optional","category_id":"optional id","supplier_id":"optional id","brand_id":"optional id","low_stock_threshold":number,"reorder_quantity":number}`
	case name == "erp.update_product":
		return `{"product_id":"required id","name":"optional","sale_price":number,"cost_price":number,"stock_quantity":number,"sku":"optional","barcode":"optional","low_stock_threshold":number,"reorder_quantity":number,"is_active":true}`
	case name == "erp.delete_product":
		return `{"product_id":"required id"}`
	case name == "erp.set_stock":
		return `{"product_id":"product id","quantity":number,"reason":"inventory adjustment reason"}`
	case strings.Contains(name, "remember"):
		return `{"key":"stable preference name","value":"fact or preference to remember"}`
	case strings.Contains(name, "schedule"):
		return `{"template_id":"report template","cadence":"daily|weekly|monthly","channel":"delivery channel","recipient":"destination"}`
	case strings.Contains(name, "refund"):
		return `{"order_id":"order id","amount":number,"reason":"business reason"}`
	case strings.Contains(name, "create"), strings.Contains(name, "record"), strings.Contains(name, "set_"):
		return `{"fields":{...required fields from the target module}}`
	default:
		return `{"id":"target record id","reason":"why this action is needed","fields":{...}}`
	}
}

func toolContractPrompt(tool ToolDef) string {
	tool = enrichToolDefinition(tool)
	header := tool.Name + " [" + tool.Scope
	if tool.RequiresApproval {
		header += ", approval required"
	}
	header += "]"
	parts := []string{header}
	if tool.InputHint != "" {
		parts = append(parts, "input: "+tool.InputHint)
	}
	if tool.OutputHint != "" {
		parts = append(parts, "returns: "+tool.OutputHint)
	}
	if tool.Freshness != "" {
		parts = append(parts, "freshness: "+tool.Freshness)
	}
	if tool.SideEffect != "" {
		parts = append(parts, "effect: "+tool.SideEffect)
	}
	if tool.ApprovalReason != "" {
		parts = append(parts, "why approval: "+tool.ApprovalReason)
	}
	return strings.Join(parts, "; ")
}
