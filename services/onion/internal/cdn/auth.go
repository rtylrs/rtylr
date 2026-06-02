package cdn

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

// apiKeyAuth guards write endpoints with a static bearer token. It is a
// service-to-service credential (the frontends never upload directly; their
// backends or build steps do), which is why this is a shared API key rather
// than a per-user JWT. When no key is configured the middleware is a no-op —
// useful for local dev, never for production.
func (c *CDN) apiKeyAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if c.apiKey == "" {
			next.ServeHTTP(w, r)
			return
		}
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		// Constant-time compare so a timing side-channel can't reveal the key.
		if subtle.ConstantTimeCompare([]byte(token), []byte(c.apiKey)) != 1 {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
