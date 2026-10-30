package application

import (
	"encoding/json"
	"strings"

	"github.com/voxire/onion/internal/platform/snips"
)

// toolMarker is one <tool>{"name","input"}</tool> call in model output.
type toolMarker struct {
	Name  string         `json:"name"`
	Input map[string]any `json:"input"`
}

// parseToolMarkers finds every well-formed tool marker in text, and any call
// in the model's own markup (see snips.ParseNativeToolCalls), and returns the
// calls and the text with them removed.
func parseToolMarkers(text string) ([]toolMarker, string) {
	const openTag, closeTag = "<tool>", "</tool>"
	var calls []toolMarker
	var rest strings.Builder
	for {
		start := strings.Index(text, openTag)
		// Native markup is read where it stands; inside a <tool> marker it is
		// part of an input and stays as written.
		if native := snips.NativeToolCallStart(text); native >= 0 && (start < 0 || native < start) {
			rest.WriteString(text[:native])
			found, end, _ := snips.ParseNativeToolCalls(text[native:], true)
			for _, call := range found {
				calls = append(calls, toolMarker{Name: strings.TrimSpace(call.Name), Input: call.Input})
			}
			text = text[native+end:]
			continue
		}
		if start < 0 {
			rest.WriteString(text)
			break
		}
		end := strings.Index(text[start:], closeTag)
		if end < 0 {
			rest.WriteString(text)
			break
		}
		rest.WriteString(text[:start])
		var call toolMarker
		if err := json.Unmarshal([]byte(strings.TrimSpace(text[start+len(openTag):start+end])), &call); err == nil && strings.TrimSpace(call.Name) != "" {
			call.Name = strings.TrimSpace(call.Name)
			if call.Input == nil {
				call.Input = map[string]any{}
			}
			calls = append(calls, call)
		}
		text = text[start+end+len(closeTag):]
	}
	return calls, rest.String()
}
