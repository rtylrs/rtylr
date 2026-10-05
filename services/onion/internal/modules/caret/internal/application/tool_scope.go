package application

import "strings"

// BindToolInputToPageScope resolves references such as "this product" on the
// server. The selected identifier stays in the authenticated request context
// and is added only after the planner has emitted a tool marker, so the
// provider can choose a tool without receiving the record identifier.
func BindToolInputToPageScope(name string, input, pageContext map[string]any) map[string]any {
	fields, ok := pageScopeFields[name]
	if !ok || len(pageContext) == 0 {
		return input
	}

	selected, ok := pageContext["selected"].(map[string]any)
	if !ok || len(selected) == 0 {
		return input
	}

	bound := make(map[string]any, len(input)+1)
	for key, value := range input {
		bound[key] = value
	}
	for _, field := range fields {
		if hasToolInputValue(bound, fields) {
			break
		}
		for _, source := range []string{field, "id"} {
			value, exists := selected[source]
			if !exists || value == nil {
				continue
			}
			if text, isText := value.(string); isText && strings.TrimSpace(text) == "" {
				continue
			}
			bound[field] = value
			break
		}
	}
	return bound
}

func hasToolInputValue(input map[string]any, fields []string) bool {
	for _, field := range fields {
		value, ok := input[field]
		if !ok || value == nil {
			continue
		}
		if text, isText := value.(string); isText && strings.TrimSpace(text) == "" {
			continue
		}
		return true
	}
	return false
}

// pageScopeFields deliberately covers only tools whose target can be resolved
// from a selected record. Lists and aggregate reads must still receive their
// explicit filters from the user/request planner.
var pageScopeFields = map[string][]string{
	"crm.get_customer":         {"customer_id", "id"},
	"pos.get_order":            {"order_id", "id"},
	"pos.refund_order":         {"order_id", "id"},
	"store.get_order":          {"order_id", "id"},
	"store.get_product":        {"product_id", "id"},
	"support.get_ticket":       {"ticket_id", "id"},
	"support.get_customer":     {"customer_id", "id"},
	"support.get_article":      {"article_id", "id"},
	"erp.get_product":          {"product_id", "id", "barcode"},
	"erp.update_product":       {"product_id"},
	"erp.delete_product":       {"product_id"},
	"erp.set_stock":            {"product_id"},
	"erp.get_supplier":         {"supplier_id", "id"},
	"erp.get_warehouse":        {"warehouse_id", "id"},
	"erp.get_purchase_order":   {"purchase_order_id", "id"},
	"erp.get_stock_transfer":   {"stock_transfer_id", "id"},
	"hr.get_employee":          {"employee_id", "id"},
	"hr.get_payroll_run":       {"payroll_run_id", "run_id", "id"},
	"hr.get_timesheet":         {"timesheet_id", "id"},
	"booking.get_appointment":  {"appointment_id", "id"},
	"flow.get_task":            {"task_id", "id"},
	"finance.get_invoice":      {"invoice_id", "id"},
	"finance.get_expense":      {"expense_id", "id"},
	"finance.get_bank_account": {"bank_account_id", "id"},
}
