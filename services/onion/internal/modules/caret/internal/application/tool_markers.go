package application

import (
	"encoding/json"
	"strings"
)

// toolMarker is one <tool>{"name","input"}</tool> call in model output.
type toolMarker struct {
	Name  string         `json:"name"`
	Input map[string]any `json:"input"`
}

// parseToolMarkers finds every well-formed tool marker in text and returns
// the calls and the text with them removed.
func parseToolMarkers(text string) ([]toolMarker, string) {
	const open, close = "<tool>", "</tool>"
	var calls []toolMarker
	var rest strings.Builder
	for {
		start := strings.Index(text, open)
		if start < 0 {
			rest.WriteString(text)
			break
		}
		end := strings.Index(text[start:], close)
		if end < 0 {
			rest.WriteString(text)
			break
		}
		rest.WriteString(text[:start])
		var call toolMarker
		if err := json.Unmarshal([]byte(strings.TrimSpace(text[start+len(open):start+end])), &call); err == nil && strings.TrimSpace(call.Name) != "" {
			call.Name = strings.TrimSpace(call.Name)
			if call.Input == nil {
				call.Input = map[string]any{}
			}
			calls = append(calls, call)
		}
		text = text[start+end+len(close):]
	}
	return calls, rest.String()
}
