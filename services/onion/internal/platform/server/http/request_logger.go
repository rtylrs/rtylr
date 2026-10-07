package http

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/slack"
)

const (
	// maxLoggedRequestBodyBytes is how much of a JSON body we read for logging
	// previews. The handler still receives the full body via replay.
	maxLoggedRequestBodyBytes        = 8 << 10
	maxLoggedBodyPreviewBytes        = 1024
	maxLoggedHeaderValueBytes        = 200
	duplicateDetectorCleanupInterval = time.Minute
)

func RequestLoggerMiddleware(ctx context.Context) func(next http.Handler) http.Handler {
	serviceName := logger.ServiceName()

	rateLimiter := &RateLimiter{
		chatLimit:   30 * time.Second,
		globalLimit: 10 * time.Second,
		chatTimers:  make(map[string]time.Time),
		globalTimer: time.Now(),
		mu:          &sync.RWMutex{},
	}

	duplicateDetector := newDuplicateDetector(15 * time.Minute)
	duplicateDetector.startCleanup(ctx, duplicateDetectorCleanupInterval)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			var bodyContent string

			if r.Body != nil && isJSONContentType(r.Header.Get("Content-Type")) {
				bodyContent = captureRequestBodyPreview(r)
			}

			sanitizedBody := sanitizeJSON(bodyContent)
			sanitizedQuery := sanitizeMap(r.URL.Query())
			sanitizedHeaders := sanitizeMap(r.Header)

			// WebSocket upgrades require the raw ResponseWriter for Hijack;
			// skip wrapping and let the connection pass through unlogged.
			if strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
				next.ServeHTTP(w, r)
				return
			}

			responseWriter := &ResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(responseWriter, r)

			duration := time.Since(start)

			// Failed requests (4xx/5xx) → the Slack request-logs channel (batched).
			// 2xx/3xx are left to Better Stack to keep the channel signal-dense. Malicious
			// requests are blocked upstream by SecurityMiddleware and never reach here.
			// Anonymous 401s (no token) are expected bot/scanner noise — skip them.
			if slack.Enabled() && responseWriter.statusCode >= 400 &&
				(responseWriter.statusCode != http.StatusUnauthorized || r.Header.Get("Authorization") != "") &&
				rateLimiter.CanSend("request-logs") {
				slack.Post(slack.Requests, fmt.Sprintf("`%d` %s %s · %s · %s",
					responseWriter.statusCode, r.Method, requestRouteTemplate(r), getRemoteIP(r), duration.Round(time.Millisecond)))
			}

			go func() {
				if responseWriter.statusCode < 400 {
					return
				}

				// SecurityMiddleware already batched these.
				if isMaliciousPath(r.URL.Path) {
					return
				}

				// Anonymous 401s (no Authorization header) are expected bot/scanner
				// traffic — not actionable errors. Skip them to keep logs clean.
				if responseWriter.statusCode == http.StatusUnauthorized && r.Header.Get("Authorization") == "" {
					return
				}

				requestKey := generateRequestKey(r, responseWriter.statusCode)
				shouldReport, count := duplicateDetector.ShouldReport(requestKey)

				if !shouldReport {
					return
				}

				go func() {
					logRequest(ctx, r, responseWriter.statusCode, duration, sanitizedBody, sanitizedHeaders, sanitizedQuery, serviceName, count)
				}()
			}()
		})
	}
}

type replayReadCloser struct {
	io.Reader
	closer io.Closer
}

func (r *replayReadCloser) Close() error {
	return r.closer.Close()
}

func isJSONContentType(contentType string) bool {
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return strings.Contains(strings.ToLower(contentType), "application/json")
	}
	mediaType = strings.ToLower(mediaType)
	return mediaType == "application/json" || strings.HasSuffix(mediaType, "+json")
}

func captureRequestBodyPreview(r *http.Request) string {
	originalBody := r.Body
	bodyBytes, err := io.ReadAll(io.LimitReader(originalBody, maxLoggedRequestBodyBytes+1))
	if len(bodyBytes) > 0 {
		r.Body = &replayReadCloser{
			Reader: io.MultiReader(bytes.NewReader(bodyBytes), originalBody),
			closer: originalBody,
		}
	}
	if err != nil {
		return ""
	}
	if len(bodyBytes) > maxLoggedRequestBodyBytes {
		return string(bodyBytes[:maxLoggedRequestBodyBytes])
	}
	return string(bodyBytes)
}

type ResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *ResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *ResponseWriter) Write(b []byte) (int, error) {
	return rw.ResponseWriter.Write(b)
}

func (rw *ResponseWriter) Unwrap() http.ResponseWriter { return rw.ResponseWriter }

func (rw *ResponseWriter) Flush() {
	if f, ok := rw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Hijack delegates to the underlying ResponseWriter so WebSocket upgrades work.
func (rw *ResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("underlying ResponseWriter does not implement http.Hijacker")
	}
	return h.Hijack()
}

type RateLimiter struct {
	chatLimit   time.Duration
	globalLimit time.Duration
	chatTimers  map[string]time.Time
	globalTimer time.Time
	mu          *sync.RWMutex
}

func (rl *RateLimiter) CanSend(chatID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	if now.Sub(rl.globalTimer) < rl.globalLimit {
		return false
	}

	if lastSent, exists := rl.chatTimers[chatID]; exists {
		if now.Sub(lastSent) < rl.chatLimit {
			return false
		}
	}

	rl.globalTimer = now
	rl.chatTimers[chatID] = now

	return true
}

type DuplicateInfo struct {
	firstSeen    time.Time
	lastSeen     time.Time
	lastReported time.Time
	count        int
}
