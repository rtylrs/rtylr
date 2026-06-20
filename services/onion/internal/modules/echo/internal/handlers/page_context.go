package handlers

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

const maxPageContextJSONLen = 6000

func buildRequestSystemContext(systemContext string, pageContext map[string]any) string {
	parts := make([]string, 0, 2)
	if trimmed := sanitizeUserInput(systemContext, maxSystemContextLen); trimmed != "" {
		parts = append(parts, trimmed)
	}
	if rendered := renderPageContext(pageContext); rendered != "" {
		parts = append(parts, "Structured page context:\n"+rendered)
	}
	return sanitizeUserInput(strings.Join(parts, "\n\n"), maxSystemContextLen)
}

func renderPageContext(pageContext map[string]any) string {
	if len(pageContext) == 0 {
		return ""
	}
	normalized, ok := normalizePageContext(pageContext, 0).(map[string]any)
	if !ok || len(normalized) == 0 {
		return ""
	}
	payload, err := json.MarshalIndent(normalized, "", "  ")
	if err != nil {
		return ""
	}
	if len(payload) > maxPageContextJSONLen {
		payload = payload[:maxPageContextJSONLen]
	}
	return sanitizeUserInput(string(payload), maxPageContextJSONLen)
}

func normalizePageContext(value any, depth int) any {
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
			if item := normalizePageContext(v[i], depth+1); item != nil {
				out = append(out, item)
			}
		}
		return out
	case string:
		return sanitizeUserInput(v, 400)
	case float64, bool, nil:
		return v
	default:
		return sanitizeUserInput(strings.TrimSpace(toString(v)), 200)
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
		normalizedKey := sanitizeUserInput(key, 80)
		if normalizedKey == "" {
			continue
		}
		if value := normalizePageContext(m[key], depth+1); value != nil {
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
