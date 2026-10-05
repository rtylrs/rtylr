package cdn

import "strings"

// quoteETag wraps a value in the double quotes required by the ETag grammar.
func quoteETag(s string) string {
	return `"` + s + `"`
}

// etagMatches reports whether a client's If-None-Match header covers etag, so
// the handler can reply 304 Not Modified. It honours the "*" wildcard and
// ignores the weak-validator "W/" prefix, since CDN assets are content-
// addressed and a weak match is as good as a strong one here.
func etagMatches(ifNoneMatch, etag string) bool {
	ifNoneMatch = strings.TrimSpace(ifNoneMatch)
	if ifNoneMatch == "" {
		return false
	}
	if ifNoneMatch == "*" {
		return true
	}
	for _, candidate := range strings.Split(ifNoneMatch, ",") {
		c := strings.TrimSpace(candidate)
		c = strings.TrimPrefix(c, "W/")
		if c == etag {
			return true
		}
	}
	return false
}
