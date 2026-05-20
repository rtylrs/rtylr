package handlers

import (
	"encoding/json"
	"regexp"
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

var toolBlockRe = regexp.MustCompile(`(?s)<tool>\s*(\{.*?\})\s*</tool>`)

// ExtractToolCalls finds and parses every <tool>…</tool> block in s. The
// remainder string contains the original input with each block removed so
// the visible text doesn't include the raw JSON markers.
func ExtractToolCalls(s string) (calls []ToolCallMarker, remainder string) {
	matches := toolBlockRe.FindAllStringSubmatchIndex(s, -1)
	if len(matches) == 0 {
		return nil, s
	}
	calls = make([]ToolCallMarker, 0, len(matches))
	var out []byte
	last := 0
	for _, m := range matches {
		out = append(out, s[last:m[0]]...)
		last = m[1]
		raw := s[m[2]:m[3]]
		var call ToolCallMarker
		if err := json.Unmarshal([]byte(raw), &call); err == nil {
			calls = append(calls, call)
		}
	}
	out = append(out, s[last:]...)
	remainder = string(out)
	return calls, remainder
}
