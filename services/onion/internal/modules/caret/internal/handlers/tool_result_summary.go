package handlers

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
)

const maxToolSummaryRunes = 260

var toolCollectionLabels = map[string]string{
	"appointments":   "appointments",
	"articles":       "articles",
	"bank_accounts":  "bank accounts",
	"bills":          "bills",
	"brands":         "brands",
	"campaigns":      "campaigns",
	"categories":     "categories",
	"customers":      "customers",
	"employees":      "team members",
	"expenses":       "expenses",
	"invoices":       "invoices",
	"items":          "items",
	"locations":      "locations",
	"members":        "members",
	"orders":         "orders",
	"products":       "products",
	"rewards":        "rewards",
	"rows":           "results",
	"segments":       "segments",
	"services":       "services",
	"suppliers":      "suppliers",
	"tables":         "tables",
	"tasks":          "tasks",
	"tickets":        "tickets",
	"transactions":   "transactions",
	"warehouses":     "warehouses",
	"work_schedules": "work schedules",
}

func summarizeToolResult(name string, _, output map[string]any, status, errorMessage string) string {
	target := toolResultTarget(name)
	if status == "error" {
		if strings.TrimSpace(errorMessage) == "" {
			return clampToolSummary("Could not check " + target + ".")
		}
		return clampToolSummary("Could not check " + target + ": " + errorMessage + ".")
	}
	if len(output) == 0 {
		return clampToolSummary("Checked " + target + "; no result data came back.")
	}
	if summary, ok := stringValue(output["summary"]); ok && summary != "" {
		return clampToolSummary(summary)
	}

	if key, count, first := firstCollectionSummary(output); key != "" {
		label := toolCollectionLabels[key]
		if count == 0 {
			return clampToolSummary("No " + label + " found.")
		}
		summary := fmt.Sprintf("Found %d %s.", count, label)
		if first != "" {
			summary += " First: " + first + "."
		}
		return clampToolSummary(summary)
	}

	if total, ok := numericValue(output["total"]); ok {
		label := target
		if strings.HasPrefix(name, "crm.list_") {
			label = strings.TrimPrefix(name, "crm.list_")
			label = strings.ReplaceAll(label, "_", " ")
		}
		if total == 0 {
			return clampToolSummary("No " + label + " found.")
		}
		return clampToolSummary(fmt.Sprintf("Found %s %s.", formatToolNumber(total), label))
	}

	if rowCount, ok := numericValue(output["row_count"]); ok {
		if rowCount == 0 {
			return clampToolSummary("No matching results found.")
		}
		return clampToolSummary(fmt.Sprintf("Returned %s result rows.", formatToolNumber(rowCount)))
	}

	keys := sortedOutputKeys(output)
	if len(keys) == 0 {
		return clampToolSummary("Checked " + target + ".")
	}
	if len(keys) > 4 {
		keys = keys[:4]
	}
	return clampToolSummary("Checked " + target + "; returned " + strings.Join(keys, ", ") + ".")
}

func firstCollectionSummary(output map[string]any) (string, int, string) {
	keys := sortedOutputKeys(output)
	for _, key := range keys {
		label, ok := toolCollectionLabels[key]
		if !ok || label == "" {
			continue
		}
		items, ok := collectionValue(output[key])
		if !ok {
			continue
		}
		first := ""
		if len(items) > 0 {
			first = itemDisplayName(items[0])
		}
		return key, len(items), first
	}
	return "", 0, ""
}

func collectionValue(value any) ([]any, bool) {
	switch v := value.(type) {
	case []any:
		return v, true
	case []map[string]any:
		out := make([]any, len(v))
		for i := range v {
			out[i] = v[i]
		}
		return out, true
	default:
		return nil, false
	}
}

func itemDisplayName(item any) string {
	m, ok := item.(map[string]any)
	if !ok {
		if s, ok := stringValue(item); ok {
			return s
		}
		return ""
	}
	for _, key := range []string{"name", "title", "full_name", "display_name", "email", "sku", "code", "slug"} {
		if value, ok := stringValue(m[key]); ok && value != "" {
			return value
		}
	}
	first, _ := stringValue(m["first_name"])
	last, _ := stringValue(m["last_name"])
	if strings.TrimSpace(first+" "+last) != "" {
		return strings.TrimSpace(first + " " + last)
	}
	if id, ok := stringValue(m["id"]); ok && id != "" {
		return "ID " + id
	}
	return ""
}

func sortedOutputKeys(output map[string]any) []string {
	keys := make([]string, 0, len(output))
	for key := range output {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func toolResultTarget(name string) string {
	parts := strings.Split(name, ".")
	if len(parts) == 0 {
		return "business data"
	}
	module := parts[0]
	if len(parts) > 1 {
		action := strings.TrimPrefix(parts[1], "list_")
		action = strings.TrimPrefix(action, "get_")
		action = strings.TrimPrefix(action, "create_")
		action = strings.TrimPrefix(action, "record_")
		action = strings.TrimPrefix(action, "set_")
		action = strings.ReplaceAll(action, "_", " ")
		if action != "" {
			return action
		}
	}
	return strings.ReplaceAll(module, "_", " ") + " data"
}

func numericValue(value any) (float64, bool) {
	switch v := value.(type) {
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case float64:
		return v, true
	case json.Number:
		n, err := v.Float64()
		return n, err == nil
	default:
		return 0, false
	}
}

func stringValue(value any) (string, bool) {
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v), true
	case fmt.Stringer:
		return strings.TrimSpace(v.String()), true
	case int:
		return fmt.Sprintf("%d", v), true
	case int64:
		return fmt.Sprintf("%d", v), true
	case float64:
		return formatToolNumber(v), true
	default:
		return "", false
	}
}

func formatToolNumber(n float64) string {
	if math.Trunc(n) == n {
		return fmt.Sprintf("%.0f", n)
	}
	return fmt.Sprintf("%.2f", n)
}

func clampToolSummary(summary string) string {
	summary = strings.Join(strings.Fields(summary), " ")
	runes := []rune(summary)
	if len(runes) <= maxToolSummaryRunes {
		return summary
	}
	return string(runes[:maxToolSummaryRunes]) + "..."
}
