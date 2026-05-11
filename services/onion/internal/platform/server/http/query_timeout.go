package http

import (
	"context"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/platform/logger"
)

const (
	// RequestTimeout is the hard deadline for any single HTTP request.
	// When this deadline fires, all in-flight QueryContext / ExecContext
	// calls on the request context are cancelled, so no goroutine hangs
	// past this point waiting for MySQL.
	RequestTimeout = 15 * time.Second

	// SlowQueryThreshold is the soft warn threshold. Requests that
	// complete (or time out) after this duration are logged with
	// slow_query=true so they can be tracked in log aggregation.
	SlowQueryThreshold = 5 * time.Second
)

// RequestTimeoutMiddleware enforces a hard 15 s deadline on every request
// and logs a structured warning for any request that exceeds 5 s.
func RequestTimeoutMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		ctx, cancel := context.WithTimeout(r.Context(), RequestTimeout)
		defer cancel()

		r = r.WithContext(ctx)

		rw := &timeoutResponseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)

		elapsed := time.Since(start)

		if elapsed >= SlowQueryThreshold {
			fields := []logger.Field{
				logger.Bool("slow_query", true),
				logger.String("path", r.URL.Path),
				logger.String("method", r.Method),
				logger.String("duration", elapsed.String()),
				logger.Int("status", rw.status),
			}
			if elapsed >= RequestTimeout {
				logger.Ctx(ctx).Error("request exceeded hard timeout", fields...)
			} else {
				logger.Ctx(ctx).Warn("slow request", fields...)
			}
		}
	})
}

// timeoutResponseWriter captures the status code so it can be logged.
type timeoutResponseWriter struct {
	http.ResponseWriter
	status int
}

func (t *timeoutResponseWriter) WriteHeader(code int) {
	t.status = code
	t.ResponseWriter.WriteHeader(code)
}
