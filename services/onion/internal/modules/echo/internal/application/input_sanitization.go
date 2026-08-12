package application

import (
	"strings"
	"unicode/utf8"
)

const (
	// maxUserPromptLen bounds a single user turn (~4k tokens). Anything longer is
	// truncated before it reaches the model / context builder.
	MaxUserPromptLen = 16000
	// maxSystemContextLen bounds the per-app context the client injects.
	MaxSystemContextLen = 8000
)

// sanitizeUserInput defends the prompt boundary. The streaming layer parses
// <thinking>…</thinking> and <tool>…</tool> markers out of MODEL output and acts
// on them, so if a user puts those exact tokens in their prompt they could forge
// reasoning blocks or tool calls. We neutralize the markers (insert a space so the
// exact-match regexes no longer fire) and bound the length.
func SanitizeUserInput(s string, maxLen int) string {
	s = strings.ToValidUTF8(s, "�")
	replacer := strings.NewReplacer(
		"<thinking>", "<thinking >",
		"</thinking>", "</thinking >",
		"<tool>", "<tool >",
		"</tool>", "</tool >",
	)
	s = replacer.Replace(s)
	s = strings.TrimSpace(s)
	return TruncateUTF8ToBytes(s, maxLen)
}

// truncateUTF8ToBytes applies a byte budget without splitting a UTF-8 rune.
// Prompt and continuation budgets are byte-based because provider request
// limits are measured over encoded payloads, not Go rune counts.
func TruncateUTF8ToBytes(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	if len(s) <= maxLen {
		return s
	}

	s = s[:maxLen]
	for len(s) > 0 && !utf8.ValidString(s) {
		s = s[:len(s)-1]
	}
	return s
}
