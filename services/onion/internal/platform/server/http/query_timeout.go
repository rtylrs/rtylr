package http

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/voxire/onion/internal/platform/logger"
)

// isLongLivedRequest returns true for requests that intentionally hold
// the connection open past the 15s hard timeout (WebSocket upgrades,
// SSE streams, file uploads). Applying the request timeout to these
// generates a permanent stream of "exceeded hard timeout" warnings on
// every successful connection.
func isLongLivedRequest(r *http.Request) bool {
	if strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
		return true
	}
	if strings.Contains(strings.ToLower(r.Header.Get("Connection")), "upgrade") {
		return true
	}
	if strings.Contains(r.Header.Get("Accept"), "text/event-stream") {
		return true
	}
	return false
}

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
//
// WebSocket / SSE / Upgrade requests bypass the deadline entirely —
// they are designed to live longer than 15 s, and applying the timeout
// here either kills them prematurely or floods logs with false-positive
// "exceeded hard timeout" warnings on every successful connection.
func RequestTimeoutMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isLongLivedRequest(r) {
			next.ServeHTTP(w, r)
			return
		}

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

// Hijack passes the underlying connection up so WebSocket upgrades
// work even when this middleware is in the chain. See metrics.go for
// the same pattern.
func (t *timeoutResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := t.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("underlying ResponseWriter does not implement http.Hijacker")
	}
	return h.Hijack()
}
