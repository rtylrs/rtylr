package application

import (
	"strconv"
	"strings"
)

// inputText returns the first non-empty string among keys, trimmed. Models
// name the same field differently ("city", "location", "place"), so tools
// accept a few aliases instead of failing on a near miss.
func inputText(input map[string]any, keys ...string) string {
	for _, key := range keys {
		switch value := input[key].(type) {
		case string:
			if text := strings.TrimSpace(value); text != "" {
				return text
			}
		case float64:
			return strconv.FormatFloat(value, 'f', -1, 64)
		}
	}
	return ""
}

// inputNumber reads a number given as a JSON number or a numeric string.
func inputNumber(input map[string]any, keys ...string) (float64, bool) {
	for _, key := range keys {
		switch value := input[key].(type) {
		case float64:
			return value, true
		case int:
			return float64(value), true
		case string:
			cleaned := strings.ReplaceAll(strings.TrimSpace(value), ",", "")
			if number, err := strconv.ParseFloat(cleaned, 64); err == nil {
				return number, true
			}
		}
	}
	return 0, false
}

// inputNumbers reads a list of numbers, given as a JSON array or as a
// comma- or space-separated string.
func inputNumbers(input map[string]any, keys ...string) []float64 {
	for _, key := range keys {
		switch value := input[key].(type) {
		case []any:
			out := make([]float64, 0, len(value))
			for _, item := range value {
				if number, ok := inputNumber(map[string]any{"v": item}, "v"); ok {
					out = append(out, number)
				}
			}
			if len(out) > 0 {
				return out
			}
		case string:
			fields := strings.FieldsFunc(value, func(r rune) bool { return r == ',' || r == ';' || r == ' ' || r == '\n' || r == '\t' })
			out := make([]float64, 0, len(fields))
			for _, field := range fields {
				if number, err := strconv.ParseFloat(field, 64); err == nil {
					out = append(out, number)
				}
			}
			if len(out) > 0 {
				return out
			}
		}
	}
	return nil
}

// inputBool reads a boolean given as a JSON bool or "true"/"false".
func inputBool(input map[string]any, key string) bool {
	switch value := input[key].(type) {
	case bool:
		return value
	case string:
		parsed, _ := strconv.ParseBool(strings.TrimSpace(value))
		return parsed
	}
	return false
}
