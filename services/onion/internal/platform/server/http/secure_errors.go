package http

import (
	"context"
	"encoding/json"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/getsentry/sentry-go"

	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/server/http/httpmetrics"
)

type SecureErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

func SecureErrorMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			started := time.Now()

			defer func() {
				if err := recover(); err != nil {
					// This middleware is inside the Sentry recovery boundary, so
					// recovering here would otherwise hide the panic from it. It
					// is also outside MetricsMiddleware, whose post-call
					// accounting is skipped when a panic unwinds its stack.
					// Report and account for the terminal request before writing
					// the sanitized response.
					reportRecoveredPanic(ctx, r, err)
					httpmetrics.RecordRecoveredPanic(started)

					logger.Ctx(ctx).Error("Panic recovered",
						logger.Any("error", err),
						logger.String("stack", string(debug.Stack())),
						logger.String("path", r.URL.Path),
						logger.String("method", r.Method),
						logger.String("ip", r.RemoteAddr),
					)

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)

					response := SecureErrorResponse{
						Error:   "Internal Server Error",
						Message: "An unexpected error occurred. Please try again later.",
						Status:  "error",
					}

					if err := json.NewEncoder(w).Encode(response); err != nil {
						logger.Ctx(ctx).Error("Failed to write error response", logger.Error(err))
					}
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// reportRecoveredPanic mirrors the production Sentry recovery contract for
// panics handled by SecureErrorMiddleware. Without this explicit handoff, the
// outer recovery middleware sees a normal return and never emits an event.
func reportRecoveredPanic(ctx context.Context, r *http.Request, recovered any) {
	hub := sentry.CurrentHub()
	if hub == nil {
		return
	}
	hub = hub.Clone()
	if hub == nil || hub.Scope() == nil {
		return
	}
	hub.Scope().SetRequest(r)
	hub.RecoverWithContext(ctx, recovered)
	hub.Flush(500 * time.Millisecond)
}

// recordRecoveredPanicMetrics completes the counters that MetricsMiddleware
// cannot reach after its inner handler panics. The request counter is already
// incremented before MetricsMiddleware calls the handler; only the terminal
// error and latency accounting needs to be completed here.
