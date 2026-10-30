package handlers

import (
	"fmt"
	"sort"
	"strings"

	echoapp "github.com/voxire/onion/internal/modules/caret/internal/application"
)

const maxToolPreviewDetails = 5

func buildToolActionPreview(tool ToolDef, input map[string]any) string {
	if tool.Scope != "write" && !tool.RequiresApproval {
		return ""
	}

	if preview := buddyRoutinePreview(tool.Name, input); preview != "" {
		return preview
	}
	action := friendlyToolAction(tool)
	details := previewInputDetails(input)
	if details == "" {
		return fmt.Sprintf("I am about to %s. Confirm before I make any changes.", action)
	}
	return fmt.Sprintf("I am about to %s: %s. Confirm before I make any changes.", action, details)
}

func friendlyToolAction(tool ToolDef) string {
	switch tool.Name {
	case "memory.remember":
		return "remember this for future Caret conversations"
	case "crm.create_customer":
		return "create a customer"
	case "pos.refund_order":
		return "refund an order"
	case "erp.set_stock":
		return "set product stock"
	case "erp.create_product":
		return "add an inventory product"
	case "erp.update_product":
		return "update an inventory product"
	case "erp.delete_product":
		return "delete an inventory product"
	case "erp.create_category":
		return "create an inventory category"
	case "erp.create_brand":
		return "create an inventory brand"
	case "erp.create_supplier":
		return "create a supplier"
	case "erp.create_warehouse":
		return "create a warehouse"
	case "erp.create_purchase_order":
		return "create a purchase order"
	case "erp.approve_purchase_order":
		return "approve a purchase order"
	case "erp.receive_purchase_order":
		return "receive a purchase order into inventory"
	case "erp.create_stock_transfer":
		return "create a stock transfer"
	case "erp.receive_stock_transfer":
		return "receive a stock transfer"
	case "hr.create_employee":
		return "create an employee record"
	case "flow.create_task":
		return "create a task"
	case "finance.record_expense":
		return "record an expense"
	case "finance.record_bill":
		return "record a supplier bill"
	case "finance.create_quote":
		return "create a customer quote"
	case "finance.create_invoice":
		return "create a customer invoice"
	case "finance.record_payment":
		return "record a customer payment"
	case "finance.send_invoice":
		return "send an invoice"
	case "booking.create_appointment":
		return "create a booking appointment"
	case "insights.schedule_report":
		return "schedule a recurring business insight"
	default:
		description := strings.TrimSuffix(strings.TrimSpace(tool.Description), ".")
		if description != "" {
			return strings.ToLower(description[:1]) + description[1:]
		}
		return strings.ReplaceAll(tool.Name, ".", " ")
	}
}

func previewInputDetails(input map[string]any) string {
	if len(input) == 0 {
		return ""
	}

	keys := make([]string, 0, len(input))
	for key, value := range input {
		if strings.TrimSpace(key) == "" || value == nil {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	if len(keys) > maxToolPreviewDetails {
		keys = keys[:maxToolPreviewDetails]
	}

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		value := previewValue(input[key])
		if value == "" {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s %s", humanizePreviewKey(key), value))
	}
	return strings.Join(parts, ", ")
}

func humanizePreviewKey(key string) string {
	key = strings.TrimSpace(strings.ToLower(key))
	key = strings.NewReplacer("_", " ", "-", " ").Replace(key)
	switch key {
	case "id":
		return "ID"
	case "customer id":
		return "customer ID"
	case "order id":
		return "order ID"
	case "product id":
		return "product ID"
	case "supplier id":
		return "supplier ID"
	default:
		return key
	}
}

func previewValue(value any) string {
	switch v := value.(type) {
	case string:
		return quotePreviewValue(echoapp.SanitizeUserInput(v, 80))
	case fmt.Stringer:
		return quotePreviewValue(echoapp.SanitizeUserInput(v.String(), 80))
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
		return echoapp.SanitizeUserInput(fmt.Sprint(v), 80)
	case []any:
		return fmt.Sprintf("%d items", len(v))
	case map[string]any:
		if len(v) == 0 {
			return ""
		}
		return fmt.Sprintf("%d fields", len(v))
	default:
		return quotePreviewValue(echoapp.SanitizeUserInput(fmt.Sprint(v), 80))
	}
}

func quotePreviewValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return fmt.Sprintf("%q", value)
}

// buddyRoutinePreview reads a routine change back in plain words, so the
// approval says what will run and when rather than listing raw fields.
func buddyRoutinePreview(name string, input map[string]any) string {
	title, _ := input["name"].(string)
	title = strings.TrimSpace(title)
	every := everyMinutesText(input["every_minutes"])
	switch name {
	case "buddy.create_routine":
		text := "I'll set up a routine"
		if title != "" {
			text += fmt.Sprintf(" called %q", title)
		}
		if every != "" {
			text += ", " + every
		}
		if first, ok := input["first_run_at"].(string); ok && strings.TrimSpace(first) != "" {
			text += ", starting " + strings.TrimSpace(first)
		}
		return text + ". It runs on its own and posts in your Buddy chat."
	case "buddy.update_routine":
		if enabled, ok := input["enabled"].(bool); ok && len(input) <= 2 {
			if enabled {
				return "I'll resume this routine."
			}
			return "I'll pause this routine. It won't run until you turn it back on."
		}
		if every != "" {
			return fmt.Sprintf("I'll change this routine to run %s.", every)
		}
		return "I'll update this routine."
	case "buddy.delete_routine":
		return "I'll delete this routine. It won't run again."
	}
	return ""
}

func everyMinutesText(value any) string {
	minutes, ok := value.(float64)
	if !ok || minutes <= 0 {
		return ""
	}
	n := int(minutes)
	unit := func(count int, one, many string) string {
		if count == 1 {
			return "every " + one
		}
		return fmt.Sprintf("every %d %s", count, many)
	}
	switch {
	case n%10080 == 0:
		return unit(n/10080, "week", "weeks")
	case n%1440 == 0:
		return unit(n/1440, "day", "days")
	case n%60 == 0:
		return unit(n/60, "hour", "hours")
	default:
		return unit(n, "minute", "minutes")
	}
}
