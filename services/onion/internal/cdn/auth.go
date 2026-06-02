package cdn

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"strings"
)

// apiKeyAuth guards write endpoints. It accepts either:
//
//  1. the shared CDN_API_KEY (a service-to-service credential for internal
//     callers / build steps), or
//  2. any user's personal API key that carries the "upload" scope — created
//     and managed in Dash, stored hashed in hub_api_keys.
//
// When no shared key is configured the middleware is a no-op (local dev only),
// preserving the historical "auth disabled if unset" behaviour. The user-key
// path stays active regardless so DB-backed keys work in every environment.
func (c *CDN) apiKeyAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")

		// Shared key path (back-compat). Empty configured key => dev-open.
		if c.apiKey == "" {
			next.ServeHTTP(w, r)
			return
		}
		// Constant-time compare so a timing side-channel can't reveal the key.
		if token != "" && subtle.ConstantTimeCompare([]byte(token), []byte(c.apiKey)) == 1 {
			next.ServeHTTP(w, r)
			return
		}

		// Per-user key path: a personal `upload`-scoped key from hub_api_keys.
		if token != "" {
			sum := sha256.Sum256([]byte(token))
			if _, ok := c.uploadKeyOwnerByHash(r.Context(), hex.EncodeToString(sum[:])); ok {
				next.ServeHTTP(w, r)
				return
			}
		}

		http.Error(w, "unauthorized", http.StatusUnauthorized)
	})
}
