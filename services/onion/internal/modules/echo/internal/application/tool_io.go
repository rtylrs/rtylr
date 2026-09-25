package application

import (
	"encoding/json"
	"regexp"
	"strings"
	"unicode"
)

var camelBoundary = regexp.MustCompile(`([a-z0-9])([A-Z])`)

// normalizeToolInput gives every executor one stable shape even when a model
// emits common camelCase aliases or wraps write fields in a fields/data
// object. It is intentionally conservative: it never invents IDs or values.
func NormalizeToolInput(name string, input map[string]any) map[string]any {
	if len(input) == 0 {
		return map[string]any{}
	}
	out := make(map[string]any, len(input))
	for key, value := range input {
		canonical := canonicalToolFieldName(key)
		if _, exists := out[canonical]; !exists {
			out[canonical] = normalizeToolValue(value)
		}
	}
	if tool, ok := ToolByName(name); ok && tool.Scope == "write" {
		for _, wrapper := range []string{"fields", "data", "payload", "record"} {
			nested, ok := out[wrapper].(map[string]any)
			if !ok {
				continue
			}
			for key, value := range nested {
				if _, exists := out[key]; !exists {
					out[key] = value
				}
			}
		}
	}
	return out
}

func canonicalToolFieldName(key string) string {
	key = strings.TrimSpace(key)
	key = camelBoundary.ReplaceAllString(key, `${1}_${2}`)
	key = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '_' || r == '-' {
			return unicode.ToLower(r)
		}
		return '_'
	}, key)
	return strings.Trim(strings.ReplaceAll(key, "-", "_"), "_")
}

func normalizeToolValue(value any) any {
	switch value := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(value))
		for key, nested := range value {
			out[canonicalToolFieldName(key)] = normalizeToolValue(nested)
		}
		return out
	case []any:
		out := make([]any, len(value))
		for i, nested := range value {
			out[i] = normalizeToolValue(nested)
		}
		return out
	default:
		return value
	}
}

func SanitizeToolOutput(output map[string]any) map[string]any {
	if output == nil {
		return nil
	}
	value := sanitizeToolValue(output, 0)
	clean, ok := value.(map[string]any)
	if !ok {
		return map[string]any{"result": value}
	}
	encoded, err := json.Marshal(clean)
	if err == nil && len(encoded) <= MaxToolOutputBytes {
		return clean
	}
	return map[string]any{
		"truncated":        true,
		"notice":           "The tool returned more data than this response can safely display. Use a narrower filter or date range.",
		"available_fields": sortedToolOutputKeys(clean),
	}
}

func sanitizeToolValue(value any, depth int) any {
	if depth > 6 {
		return "[nested output omitted]"
	}
	switch value := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(value))
		for key, nested := range value {
			if sensitiveToolField(key) {
				out[key] = "[redacted]"
				continue
			}
			out[key] = sanitizeToolValue(nested, depth+1)
		}
		return out
	case []any:
		out := make([]any, len(value))
		for i, nested := range value {
			out[i] = sanitizeToolValue(nested, depth+1)
		}
		return out
	default:
		return value
	}
}

func sensitiveToolField(key string) bool {
	key = strings.ToLower(strings.ReplaceAll(strings.TrimSpace(key), "-", "_"))
	for _, term := range []string{"password", "passwd", "secret", "access_token", "refresh_token", "api_key", "apikey", "authorization", "cookie", "private_key", "ssn"} {
		if strings.Contains(key, term) {
			return true
		}
	}
	return false
}

func sortedToolOutputKeys(output map[string]any) []string {
	keys := make([]string, 0, len(output))
	for key := range output {
		keys = append(keys, key)
	}
	// JSON object key order is not important; sorting is done without pulling
	// in another package so the fallback remains deterministic in logs/tests.
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[j] < keys[i] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	return keys
}
