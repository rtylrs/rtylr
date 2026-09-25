package application

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

const MaxPageContextJSONLen = 6000

// BuildRequestSystemContext creates the small, provider-safe scope envelope
// used by Caret's planner prompt. The client may still send rich page context
// to the authenticated Onion request so the server can bind a selected record
// to a tool call, but record values, filter values, business names, hints, and
// row payloads must never cross the provider boundary.
func BuildRequestSystemContext(systemContext string, pageContext map[string]any) string {
	metadata := plannerContextMetadata{}
	collectPlannerPageMetadata(&metadata, pageContext)
	collectPlannerTextMetadata(&metadata, systemContext)

	if metadata.empty() {
		return ""
	}

	var b strings.Builder
	b.WriteString("Planner scope metadata only; business values are intentionally withheld:\n")
	writePlannerMetadataLine(&b, "app", metadata.app)
	writePlannerMetadataLine(&b, "module", metadata.module)
	writePlannerMetadataLine(&b, "page", metadata.page)
	writePlannerMetadataLine(&b, "route", metadata.route)
	if len(metadata.selectedFields) > 0 {
		writePlannerMetadataLine(&b, "selected entity fields", strings.Join(metadata.selectedFields, ", "))
	}
	if metadata.selectedPresent {
		b.WriteString("- A selected entity is available for server-side binding; omit its identifier when referring to the current selection.\n")
	}
	if len(metadata.filterFields) > 0 {
		writePlannerMetadataLine(&b, "available filter fields", strings.Join(metadata.filterFields, ", "))
	}
	if metadata.visibleRowsPresent {
		b.WriteString("- Record values from the page are server-only; use a listed read tool instead of inferring them.\n")
	}
	return TruncateUTF8ToBytes(strings.TrimSpace(b.String()), MaxSystemContextLen)
}

type plannerContextMetadata struct {
	app                string
	module             string
	page               string
	route              string
	selectedFields     []string
	filterFields       []string
	selectedPresent    bool
	visibleRowsPresent bool
}

func (m plannerContextMetadata) empty() bool {
	return m.app == "" && m.module == "" && m.page == "" && m.route == "" &&
		len(m.selectedFields) == 0 && len(m.filterFields) == 0 && !m.selectedPresent && !m.visibleRowsPresent
}

var plannerSurfaceTokens = map[string]struct{}{
	"analytics": {}, "appointments": {}, "booking": {}, "business": {}, "calendar": {}, "dash": {},
	"campaign": {}, "campaigns": {}, "category": {}, "categories": {}, "crm": {},
	"customer": {}, "customers": {}, "dashboard": {}, "employees": {}, "erp": {},
	"finance": {}, "flow": {}, "hr": {}, "insights": {}, "inventory": {},
	"invoice": {}, "invoices": {}, "loyalty": {}, "menus": {}, "orders": {},
	"payroll": {}, "pos": {}, "products": {}, "projects": {}, "reports": {},
	"settings": {}, "store": {}, "support": {}, "tasks": {}, "team": {}, "rtylr": {},
}

func collectPlannerPageMetadata(metadata *plannerContextMetadata, pageContext map[string]any) {
	if metadata == nil || len(pageContext) == 0 {
		return
	}
	if metadata.app == "" {
		metadata.app = plannerSurfaceValue(pageContext["app"])
	}
	if metadata.module == "" {
		metadata.module = plannerSurfaceValue(pageContext["module"])
	}
	if metadata.page == "" {
		metadata.page = plannerSurfaceValue(pageContext["page"])
	}
	if metadata.route == "" {
		metadata.route = plannerRouteValue(pageContext["route"])
	}

	if selected, ok := pageContext["selected"]; ok {
		metadata.selectedPresent = true
		metadata.selectedFields = appendUniquePlannerFields(metadata.selectedFields, plannerFieldNames(selected)...)
	}
	if filters, ok := pageContext["filters"]; ok {
		metadata.filterFields = appendUniquePlannerFields(metadata.filterFields, plannerFieldNames(filters)...)
	}
	if _, ok := pageContext["visibleRows"]; ok {
		metadata.visibleRowsPresent = true
	}
}

func collectPlannerTextMetadata(metadata *plannerContextMetadata, systemContext string) {
	if metadata == nil {
		return
	}
	for _, rawLine := range strings.Split(systemContext, "\n") {
		line := strings.TrimSpace(rawLine)
		line = strings.TrimLeft(line, "- ")
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = strings.ToLower(strings.TrimSpace(key))
		switch key {
		case "app", "application":
			if metadata.app == "" {
				metadata.app = plannerSurfaceValue(value)
			}
		case "module", "visible module", "current module":
			if metadata.module == "" {
				metadata.module = plannerSurfaceValue(value)
			}
		case "page", "current page":
			if metadata.page == "" {
				metadata.page = plannerSurfaceValue(value)
			}
		case "route":
			if metadata.route == "" {
				metadata.route = plannerRouteValue(value)
			}
		case "selected entity fields", "selected scope fields":
			metadata.selectedPresent = true
			metadata.selectedFields = appendUniquePlannerFields(metadata.selectedFields, plannerFieldList(value)...)
		case "available filter fields", "filter fields":
			metadata.filterFields = appendUniquePlannerFields(metadata.filterFields, plannerFieldList(value)...)
		}
	}

	// Older clients sometimes put the page context in a labeled JSON block.
	// Inspect only the shape and field names; never copy any JSON value into the
	// provider prompt.
	trimmed := strings.TrimSpace(systemContext)
	if start := strings.Index(trimmed, "{"); start >= 0 {
		if end := strings.LastIndex(trimmed, "}"); end > start {
			var decoded map[string]any
			if err := json.Unmarshal([]byte(trimmed[start:end+1]), &decoded); err == nil {
				collectPlannerPageMetadata(metadata, decoded)
			}
		}
	}
}

func writePlannerMetadataLine(b *strings.Builder, key, value string) {
	if strings.TrimSpace(value) == "" {
		return
	}
	b.WriteString("- ")
	b.WriteString(key)
	b.WriteString(": ")
	b.WriteString(value)
	b.WriteString("\n")
}

func plannerSurfaceValue(value any) string {
	text := strings.ToLower(strings.TrimSpace(fmt.Sprint(value)))
	for _, rawToken := range strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != '_' && r != '-'
	}) {
		token := strings.Trim(rawToken, "_-")
		if _, ok := plannerSurfaceTokens[token]; ok {
			return token
		}
	}
	return ""
}

func plannerRouteValue(value any) string {
	raw := strings.TrimSpace(fmt.Sprint(value))
	if raw == "" {
		return ""
	}
	if query := strings.IndexAny(raw, "?#"); query >= 0 {
		raw = raw[:query]
	}
	segments := strings.Split(raw, "/")
	kept := make([]string, 0, 3)
	for _, segment := range segments {
		token := strings.ToLower(strings.TrimSpace(segment))
		if token == "" {
			continue
		}
		if _, ok := plannerSurfaceTokens[token]; ok {
			kept = append(kept, token)
		}
		if len(kept) == 3 {
			break
		}
	}
	if len(kept) == 0 {
		return ""
	}
	return "/" + strings.Join(kept, "/")
}

func plannerFieldNames(value any) []string {
	object, ok := value.(map[string]any)
	if !ok {
		return nil
	}
	fields := make([]string, 0, len(object))
	for key := range object {
		if normalized := plannerFieldName(key); normalized != "" {
			fields = append(fields, normalized)
		}
	}
	sort.Strings(fields)
	return fields
}

func plannerFieldList(value string) []string {
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t'
	})
	fields := make([]string, 0, len(parts))
	for _, part := range parts {
		if normalized := plannerFieldName(part); normalized != "" {
			fields = append(fields, normalized)
		}
	}
	return fields
}

func plannerFieldName(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" || len(value) > 80 {
		return ""
	}
	for _, r := range value {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != '_' && r != '-' && r != '.' {
			return ""
		}
	}
	return value
}

func appendUniquePlannerFields(existing []string, fields ...string) []string {
	seen := make(map[string]struct{}, len(existing)+len(fields))
	for _, field := range existing {
		seen[field] = struct{}{}
	}
	for _, field := range fields {
		if field == "" {
			continue
		}
		if _, ok := seen[field]; ok {
			continue
		}
		seen[field] = struct{}{}
		existing = append(existing, field)
	}
	sort.Strings(existing)
	if len(existing) > 16 {
		existing = existing[:16]
	}
	return existing
}

func RenderPageContext(pageContext map[string]any) string {
	if len(pageContext) == 0 {
		return ""
	}
	normalized, ok := NormalizePageContext(pageContext, 0).(map[string]any)
	if !ok || len(normalized) == 0 {
		return ""
	}
	payload, err := json.MarshalIndent(normalized, "", "  ")
	if err != nil {
		return ""
	}
	if len(payload) > MaxPageContextJSONLen {
		compact := compactPageContext(normalized, 0)
		payload, err = json.MarshalIndent(compact, "", "  ")
		if err != nil || len(payload) > MaxPageContextJSONLen {
			payload = []byte(`{"context_truncated":true}`)
		}
	}
	return SanitizeUserInput(string(payload), MaxPageContextJSONLen)
}

func compactPageContext(value any, depth int) any {
	if depth > 3 {
		return nil
	}
	switch value := value.(type) {
	case map[string]any:
		keys := make([]string, 0, len(value))
		for key := range value {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		if len(keys) > 16 {
			keys = keys[:16]
		}
		out := make(map[string]any, len(keys))
		for _, key := range keys {
			if item := compactPageContext(value[key], depth+1); item != nil {
				out[key] = item
			}
		}
		return out
	case []any:
		limit := len(value)
		if limit > 4 {
			limit = 4
		}
		out := make([]any, 0, limit)
		for i := 0; i < limit; i++ {
			if item := compactPageContext(value[i], depth+1); item != nil {
				out = append(out, item)
			}
		}
		return out
	case string:
		return TruncateUTF8ToBytes(value, 120)
	default:
		return value
	}
}

func NormalizePageContext(value any, depth int) any {
	if depth > 4 {
		return nil
	}
	switch v := value.(type) {
	case map[string]any:
		return normalizePageContextMap(v, depth)
	case []any:
		maxItems := len(v)
		if maxItems > 8 {
			maxItems = 8
		}
		out := make([]any, 0, maxItems)
		for i := 0; i < maxItems; i++ {
			if item := NormalizePageContext(v[i], depth+1); item != nil {
				out = append(out, item)
			}
		}
		return out
	case string:
		return SanitizeUserInput(v, 400)
	case float64, bool, nil:
		return v
	default:
		return SanitizeUserInput(strings.TrimSpace(toString(v)), 200)
	}
}

func normalizePageContextMap(m map[string]any, depth int) map[string]any {
	keys := make([]string, 0, len(m))
	for key := range m {
		if trimmed := strings.TrimSpace(key); trimmed != "" {
			keys = append(keys, trimmed)
		}
	}
	sort.Strings(keys)
	if len(keys) > 32 {
		keys = keys[:32]
	}

	out := make(map[string]any, len(keys))
	for _, key := range keys {
		normalizedKey := SanitizeUserInput(key, 80)
		if normalizedKey == "" {
			continue
		}
		if value := NormalizePageContext(m[key], depth+1); value != nil {
			out[normalizedKey] = value
		}
	}
	return out
}

func toString(value any) string {
	switch v := value.(type) {
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	default:
		return strings.TrimSpace(fmt.Sprint(value))
	}
}
