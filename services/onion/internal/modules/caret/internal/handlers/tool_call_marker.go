package handlers

import (
	"encoding/json"
	"strings"
)

// ToolCallMarker is the wire shape we expect inside model output when it
// wants to invoke a tool. The model emits these inline, the handler scans
// them out of the stream and dispatches the call asynchronously.
//
// Format on the wire (one per fenced JSON block, surrounded by sentinels):
//
//	<tool>{"name":"crm.list_customers","input":{"limit":10}}</tool>
type ToolCallMarker struct {
	Name  string         `json:"name"`
	Input map[string]any `json:"input"`
}

// ExtractToolCalls finds and parses every <tool>…</tool> block in s. The
// remainder string contains the original input with each block removed so
// the visible text doesn't include the raw JSON markers.
func ExtractToolCalls(s string) (calls []ToolCallMarker, remainder string) {
	const openTag, closeTag = "<tool>", "</tool>"
	var out strings.Builder
	for {
		start := strings.Index(s, openTag)
		if start < 0 {
			out.WriteString(s)
			break
		}
		out.WriteString(s[:start])
		rest := s[start+len(openTag):]
		end := strings.Index(rest, closeTag)
		if end < 0 {
			out.WriteString(s[start:])
			break
		}
		var call ToolCallMarker
		raw := strings.TrimSpace(rest[:end])
		if err := json.Unmarshal([]byte(raw), &call); err == nil && strings.TrimSpace(call.Name) != "" {
			calls = append(calls, call)
		} else {
			out.WriteString(s[start : start+len(openTag)+end+len(closeTag)])
		}
		s = rest[end+len(closeTag):]
	}
	return calls, out.String()
}
