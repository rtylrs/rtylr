package http

import (
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/platform/origins"
)

func allowedHeaders() []string {
	return []string{
		"Authorization",
		"Content-Type",
		"Idempotency-Key",
		"Accept",
		"Origin",
		"User-Agent",
		"Referer",
		"X-Requested-With",
		"Access-Control-Allow-Origin",
		"Access-Control-Allow-Credentials",
		"Access-Control-Allow-Headers",
		"Access-Control-Allow-Methods",
		"Access-Control-Expose-Headers",
		"x-request-id",
		"profile_token",
		"client_platform",
		"X-Product-Platform",
		"Product-Platform",
		"client_version",
		"language",
		"client_ip",
	}
}

func corsMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			allowed := origins.IsAllowed(origin)

			// The allowed origin is reflected per-request, so the response
			// varies by Origin. Without this, a shared cache could serve one
			// origin's Access-Control-Allow-Origin to another. Add (not Set) so
			// it composes with any Vary a handler sets (e.g. the CDN's Accept).
			w.Header().Add("Vary", "Origin")

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(allowedHeaders(), ", "))
			w.Header().Set("Access-Control-Expose-Headers", "Content-Length")
			w.Header().Set("Access-Control-Max-Age", "600")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
