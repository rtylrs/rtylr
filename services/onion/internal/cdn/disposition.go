package cdn

import "strings"

// sanitizeFilename makes a user-supplied original filename safe to echo back in
// a Content-Disposition header. It drops control characters (which could inject
// extra headers), replaces path separators and quotes that would break the
// header grammar, trims surrounding spaces/dots, and caps the length. An empty
// or all-stripped name falls back to "download".
func sanitizeFilename(name string) string {
	const maxLen = 200
	var b strings.Builder
	for _, r := range name {
		if b.Len() >= maxLen {
			break
		}
		switch {
		case r < 0x20 || r == 0x7f:
			// control character (CR/LF/etc.) — drop it
		case r == '"' || r == '\\' || r == '/':
			b.WriteByte('_')
		default:
			b.WriteRune(r)
		}
	}
	out := strings.Trim(b.String(), " .")
	if out == "" {
		return "download"
	}
	return out
}
