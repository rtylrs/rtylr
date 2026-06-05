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

// isLongLivedRequest matches WS / SSE / Upgrade so the 30s timeout doesn't fire.
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
	// Match SSE endpoints by path too — browser fetch()-based SSE sends
	// Accept: */* (not text/event-stream), so the header check alone misses it
	// and the 30s timeout would kill the stream mid-flight (→ tunnel 502).
	if strings.HasSuffix(r.URL.Path, "/generate-text-stream") {
		return true
	}
	return false
}

const (
	// RequestTimeout is the hard deadline for any single HTTP request.
	// When this deadline fires, all in-flight QueryContext / ExecContext
	// calls on the request context are cancelled, so no goroutine hangs
	// past this point waiting for MySQL.
	RequestTimeout = 30 * time.Second

	// SlowQueryThreshold is the soft warn threshold. Requests that
	// complete (or time out) after this duration are logged with
	// slow_query=true so they can be tracked in log aggregation.
	SlowQueryThreshold = 5 * time.Second
)

// RequestTimeoutMiddleware enforces a 30s hard deadline and warns on >5s.
// Skips long-lived (WS/SSE/Upgrade) requests.
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

func (t *timeoutResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := t.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("underlying ResponseWriter does not implement http.Hijacker")
	}
	return h.Hijack()
}
