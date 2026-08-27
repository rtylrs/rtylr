package http

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"

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
				(responseWriter.statusCode != http.StatusUnauthorized || r.Header.Get("Authorization") != "") {
				slack.Post(slack.Requests, fmt.Sprintf("`%d` %s %s · %s · %s",
					responseWriter.statusCode, r.Method, r.URL.Path, getRemoteIP(r), duration.Round(time.Millisecond)))
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

				if rateLimiter.CanSend("request-logs") {
					go func() {
						logRequest(ctx, r, responseWriter.statusCode, duration, sanitizedBody, sanitizedHeaders, sanitizedQuery, serviceName, count)
					}()
				}
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

type DuplicateDetector struct {
	requests map[string]*DuplicateInfo
	mu       *sync.RWMutex
	ttl      time.Duration
}

func newDuplicateDetector(ttl time.Duration) *DuplicateDetector {
	return &DuplicateDetector{
		requests: make(map[string]*DuplicateInfo),
		mu:       &sync.RWMutex{},
		ttl:      ttl,
	}
}

func (dd *DuplicateDetector) ShouldReport(requestKey string) (shouldReport bool, count int) {
	dd.mu.Lock()
	defer dd.mu.Unlock()

	now := time.Now()

	if info, exists := dd.requests[requestKey]; exists {
		info.count++
		info.lastSeen = now

		if info.count > 1 && now.Sub(info.lastReported) > 5*time.Minute {
			info.lastReported = now

			return true, info.count
		}

		return false, info.count
	}

	dd.requests[requestKey] = &DuplicateInfo{
		firstSeen:    now,
		lastSeen:     now,
		lastReported: now,
		count:        1,
	}

	return true, 1
}

func (dd *DuplicateDetector) startCleanup(ctx context.Context, interval time.Duration) {
	if ctx == nil || ctx.Done() == nil || interval <= 0 {
		return
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				dd.Cleanup()
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (dd *DuplicateDetector) Cleanup() {
	dd.mu.Lock()
	defer dd.mu.Unlock()

	now := time.Now()
	for key, info := range dd.requests {
		if now.Sub(info.lastSeen) > dd.ttl {
			delete(dd.requests, key)
		}
	}
}

func generateRequestKey(r *http.Request, statusCode int) string {
	var key strings.Builder

	_, _ = fmt.Fprintf(&key, "%s:%s:%d", r.Method, r.URL.Path, statusCode)

	importantParams := []string{"user_id", "session_id", "id"}
	for _, param := range importantParams {
		if value := r.URL.Query().Get(param); value != "" {
			_, _ = fmt.Fprintf(&key, ":%s=%s", param, value)
		}
	}

	if routeCtx := chi.RouteContext(r.Context()); routeCtx != nil {
		for i, paramKey := range routeCtx.URLParams.Keys {
			if i < len(routeCtx.URLParams.Values) {
				_, _ = fmt.Fprintf(&key, ":%s=%s", paramKey, routeCtx.URLParams.Values[i])
			}
		}
	}

	return key.String()
}

var sensitiveFieldParts = []string{
	"password",
	"token",
	"authorization",
	"secret",
	"api_key",
	"apikey",
	"private_key",
	"credential",
	"cookie",
}

func sanitizeMap(data map[string][]string) map[string][]string {
	sanitized := make(map[string][]string, len(data))
	for key, values := range data {
		if isSensitiveField(key) {
			sanitized[key] = []string{"[REDACTED]"}
			continue
		}
		trimmed := make([]string, len(values))
		for i, value := range values {
			trimmed[i] = truncateString(value, maxLoggedHeaderValueBytes)
		}
		sanitized[key] = trimmed
	}

	return sanitized
}

func sanitizeJSON(body string) string {
	if body == "" {
		return body
	}

	var data json.RawMessage
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		return body
	}

	data = sanitizeJSONValue(data)

	return truncateString(string(data), maxLoggedBodyPreviewBytes)
}

func sanitizeJSONValue(raw json.RawMessage) json.RawMessage {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err == nil && object != nil {
		for key, value := range object {
			if isSensitiveField(key) {
				object[key] = json.RawMessage(`"[REDACTED]"`)
				continue
			}
			object[key] = sanitizeJSONValue(value)
		}
		result, err := json.Marshal(object)
		if err == nil {
			return result
		}
	}

	var array []json.RawMessage
	if err := json.Unmarshal(raw, &array); err == nil && array != nil {
		for index, value := range array {
			array[index] = sanitizeJSONValue(value)
		}
		result, err := json.Marshal(array)
		if err == nil {
			return result
		}
	}

	return raw
}

func isSensitiveField(key string) bool {
	key = strings.ToLower(key)
	for _, f := range sensitiveFieldParts {
		if strings.Contains(key, f) {
			return true
		}
	}

	return key == "session_id" || strings.HasSuffix(key, "_session_id")
}

func logRequest(ctx context.Context, r *http.Request, statusCode int, duration time.Duration, bodyContent string, headers, query map[string][]string, serviceName string, count int) {
	routeParams := routeParamsFromRequest(r)

	queryParams := make([]string, 0, len(query))
	for key, values := range query {
		queryParams = append(queryParams, fmt.Sprintf("%s: %s", key, strings.Join(values, ",")))
	}

	headerParams := make([]string, 0, len(headers))
	for key, values := range headers {
		headerParams = append(headerParams, fmt.Sprintf("%s: %s", key, strings.Join(values, ",")))
	}

	var bodyPreview string
	if bodyContent != "" {
		bodyPreview = bodyContent
	} else {
		bodyPreview = "[No body]"
	}

	var statusIcon string
	if statusCode >= 400 && statusCode < 500 {
		statusIcon = "❌"
	} else if statusCode >= 500 {
		statusIcon = "💥"
	} else {
		statusIcon = "✅"
	}

	var countInfo string
	if count > 1 {
		countInfo = fmt.Sprintf(" (occurred %d times)", count)
	}

	message := fmt.Sprintf("%s HTTP request%s", statusIcon, countInfo)
	fields := []logger.Field{
		logger.String("service", serviceName),
		logger.String("method", r.Method),
		logger.String("url", r.URL.String()),
		logger.Int("status", statusCode),
		logger.String("duration", duration.String()),
		logger.String("user_agent", truncateString(r.UserAgent(), 100)),
		logger.String("remote_ip", getRemoteIP(r)),
		logger.String("route_params", formatList(routeParams, "• ")),
		logger.String("query_params", formatList(queryParams, "• ")),
		logger.String("headers", formatList(headerParams, "• ")),
		logger.String("body_preview", bodyPreview),
		logger.String("timestamp", time.Now().Format("2006-01-02 15:04:05 UTC")),
	}

	if statusCode >= 500 {
		logger.Ctx(ctx).Error(message, fields...)

		return
	}

	logger.Ctx(ctx).Warn(message, fields...)
}

func routeParamsFromRequest(r *http.Request) []string {
	if r == nil {
		return nil
	}

	routeCtx := chi.RouteContext(r.Context())
	if routeCtx == nil {
		return nil
	}

	routeParams := make([]string, 0, len(routeCtx.URLParams.Keys))
	for i, key := range routeCtx.URLParams.Keys {
		if i < len(routeCtx.URLParams.Values) {
			routeParams = append(routeParams, fmt.Sprintf("%s: %s", key, routeCtx.URLParams.Values[i]))
		}
	}

	return routeParams
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	return s[:maxLen-3] + "..."
}

func getRemoteIP(r *http.Request) string {
	return getClientKey(r)
}

func formatList(items []string, prefix string) string {
	if len(items) == 0 {
		return "None"
	}

	var result strings.Builder
	for _, item := range items {
		result.WriteString(prefix)
		result.WriteString(item)
		result.WriteString("\n")
	}

	return strings.TrimSpace(result.String())
}
