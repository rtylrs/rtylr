package http

import (
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

// responseCompressMiddleware gzip-deflates JSON/HTML/text responses when the
// client sends Accept-Encoding: gzip. Long-lived streams (SSE, WebSocket) and
// /metrics are skipped — compressing those mid-flight breaks clients.
func responseCompressMiddleware() MiddlewareFunc {
	compress := middleware.Compress(5)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isLongLivedRequest(r) || r.URL.Path == "/metrics" {
				next.ServeHTTP(w, r)
				return
			}

			compress(next).ServeHTTP(w, r)
		})
	}
}
