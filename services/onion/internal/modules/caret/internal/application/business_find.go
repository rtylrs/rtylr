package application

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"unicode/utf8"
)

// BusinessFindToolName searches the business's records by name across apps.
const BusinessFindToolName = "business.find"

// findSources are the list tools business.find asks, with the label for
// what they hold. Each runs through its own module's executor and
// permissions; business.find only filters what they return.
var findSources = []struct{ tool, details, label string }{
	{"crm.list_customers", "crm.get_customer", "customer"},
	{"hr.list_employees", "hr.get_employee", "team member"},
	{"erp.list_products", "erp.get_product", "product"},
	{"erp.list_suppliers", "erp.get_supplier", "supplier"},
	{"finance.list_invoices", "finance.get_invoice", "invoice"},
	{"support.list_tickets", "support.get_ticket", "support ticket"},
	{"flow.list_tasks", "flow.get_task", "task"},
	{"booking.list_appointments", "booking.get_appointment", "appointment"},
}

const maxFindMatchesPerSource = 5

// BusinessFindExecutor runs business.find.
type BusinessFindExecutor struct {
	Executors func() []ToolExecutor
}

func (e *BusinessFindExecutor) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	if name != BusinessFindToolName {
		return nil, false, nil
	}
	query := strings.TrimSpace(inputText(input, "query", "name", "search", "q"))
	if utf8.RuneCountInString(query) < 2 || utf8.RuneCountInString(query) > 100 {
		return nil, true, errors.New("give a name or word to look for, 2 to 100 characters")
	}
	if e == nil || e.Executors == nil {
		return nil, true, errors.New("business search is unavailable")
	}
	type found struct {
		label, details string
		matches        []map[string]any
		failed         bool
	}
	results := make([]found, len(findSources))
	var wg sync.WaitGroup
	for i, source := range findSources {
		results[i] = found{label: source.label, details: source.details}
		if !promptToolCanExecute(source.tool) {
			continue
		}
		wg.Add(1)
		go func(i int, tool string) {
			defer wg.Done()
			output, executed, _, err := ExecuteToolWithTimeout(ctx, e.Executors(), tool, map[string]any{"search": query, "limit": 50.0})
			if err != nil || !executed {
				results[i].failed = err != nil
				return
			}
			results[i].matches = matchingRecords(output, query)
		}(i, source.tool)
	}
	wg.Wait()
	matches := []map[string]any{}
	searched := []string{}
	unavailable := []string{}
	for _, result := range results {
		searched = append(searched, result.label+"s")
		if result.failed {
			unavailable = append(unavailable, result.label+"s")
		}
		for _, match := range result.matches {
			match["kind"] = result.label
			if _, ok := ToolByName(result.details); ok {
				match["details_tool"] = result.details
			}
			matches = append(matches, match)
		}
	}
	out := map[string]any{"query": query, "matches": matches, "searched": searched}
	if len(unavailable) > 0 {
		out["not_checked"] = unavailable
	}
	if len(matches) == 0 {
		out["note"] = "Nothing matches that name in the records Caret can see."
	}
	return out, true, nil
}

// matchingRecords returns the records of a list output whose name or other
// text fields contain every word of the query, best matches first.
func matchingRecords(output map[string]any, query string) []map[string]any {
	words := strings.Fields(strings.ToLower(query))
	type scored struct {
		record map[string]any
		score  int
	}
	candidates := []scored{}
	for _, value := range output {
		items, ok := value.([]any)
		if !ok {
			if typed, ok := value.([]map[string]any); ok {
				for _, item := range typed {
					items = append(items, item)
				}
			}
		}
		for _, item := range items {
			record, ok := item.(map[string]any)
			if !ok {
				continue
			}
			name := recordName(record)
			haystack := strings.ToLower(name + " " + recordText(record))
			all := true
			for _, word := range words {
				if !strings.Contains(haystack, word) {
					all = false
					break
				}
			}
			if !all {
				continue
			}
			score := 1
			if strings.Contains(strings.ToLower(name), strings.ToLower(query)) {
				score = 2
			}
			if strings.EqualFold(name, query) {
				score = 3
			}
			summary := map[string]any{"name": name}
			for _, key := range []string{"id", "email", "phone", "status", "role", "sku", "number", "invoice_number", "due_date", "total", "amount"} {
				if v, ok := record[key]; ok && v != nil && fmt.Sprint(v) != "" {
					summary[key] = v
				}
			}
			candidates = append(candidates, scored{summary, score})
		}
	}
	sort.SliceStable(candidates, func(i, j int) bool { return candidates[i].score > candidates[j].score })
	out := []map[string]any{}
	for i, candidate := range candidates {
		if i == maxFindMatchesPerSource {
			break
		}
		out = append(out, candidate.record)
	}
	return out
}

func recordName(record map[string]any) string {
	for _, key := range []string{"name", "full_name", "display_name", "title", "subject", "invoice_number", "number"} {
		if value, ok := record[key].(string); ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	first, _ := record["first_name"].(string)
	last, _ := record["last_name"].(string)
	return strings.TrimSpace(first + " " + last)
}

func recordText(record map[string]any) string {
	parts := []string{}
	for key, value := range record {
		if text, ok := value.(string); ok && len(text) < 200 && key != "id" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, " ")
}
