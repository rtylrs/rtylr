package handlers

import "strings"

const (
	// maxUserPromptLen bounds a single user turn (~4k tokens). Anything longer is
	// truncated before it reaches the model / context builder.
	maxUserPromptLen = 16000
	// maxSystemContextLen bounds the per-app context the client injects.
	maxSystemContextLen = 8000
)

// sanitizeUserInput defends the prompt boundary. The streaming layer parses
// <thinking>…</thinking> and <tool>…</tool> markers out of MODEL output and acts
// on them, so if a user puts those exact tokens in their prompt they could forge
// reasoning blocks or tool calls. We neutralize the markers (insert a space so the
// exact-match regexes no longer fire) and bound the length.
func sanitizeUserInput(s string, maxLen int) string {
	replacer := strings.NewReplacer(
		"<thinking>", "<thinking >",
		"</thinking>", "</thinking >",
		"<tool>", "<tool >",
		"</tool>", "</tool >",
	)
	s = replacer.Replace(s)
	s = strings.TrimSpace(s)
	if len(s) > maxLen {
		s = s[:maxLen]
	}
	return s
}
