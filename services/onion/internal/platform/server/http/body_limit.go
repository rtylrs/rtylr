package http

import (
	"context"
	stdErrors "errors"
	"net/http"

	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

// DefaultMaxRequestBodyBytes is the global ceiling for POST/PUT/PATCH bodies on
// the shared API server. JSON decode paths reuse the same limit; routes that
// need a larger or custom bound (multipart resumes, CDN uploads, webhooks) are
// exempt and enforce their own MaxBytesReader.
const DefaultMaxRequestBodyBytes = 4 << 20 // 4 MiB

func requestBodyLimitExempt(path string) bool {
	switch path {
	case "/careers/apply", "/voxhire/applications", "/voxhire/careers/apply":
		return true
	default:
		// CDN runmode (cmd/upload) — handlers set maxUpload (default 100 MiB).
		return path == "/upload" || path == "/purge" || path == "/sign"
	}
}

// RequestBodyLimitMiddleware wraps r.Body with http.MaxBytesReader for mutating
// methods so oversized payloads are rejected before handlers run.
func RequestBodyLimitMiddleware(maxBytes int64) MiddlewareFunc {
	if maxBytes <= 0 {
		maxBytes = DefaultMaxRequestBodyBytes
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodPost, http.MethodPut, http.MethodPatch:
			default:
				next.ServeHTTP(w, r)
				return
			}

			if requestBodyLimitExempt(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

func isRequestBodyTooLarge(err error) bool {
	var maxBytesErr *http.MaxBytesError
	return stdErrors.As(err, &maxBytesErr)
}

func writeRequestEntityTooLarge(ctx context.Context, w http.ResponseWriter) {
	logger.Ctx(ctx).Warn("Request body too large")

	WriteOneKitError(ctx, w, onekiterrors.New(http.StatusRequestEntityTooLarge, "Request body too large"))
}
