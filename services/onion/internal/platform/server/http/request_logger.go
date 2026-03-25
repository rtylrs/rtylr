package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/utils/alerts"
)

func RequestLoggerMiddleware(ctx context.Context) func(next http.Handler) http.Handler {
	slackBotToken := os.Getenv("SLACK_BOT_TOKEN")
	requestLogsChannelID := os.Getenv("SLACK_REQUEST_LOGS_CHANNEL")

	serviceName := "onion"

	if slackBotToken == "" || requestLogsChannelID == "" {
		logger.Global().Warn("Slack request logging not configured, skipping request logging")

		return func(next http.Handler) http.Handler {
			return next
		}
	}

	rateLimiter := &RateLimiter{
		chatLimit:   30 * time.Second,
		globalLimit: 10 * time.Second,
		chatTimers:  make(map[string]time.Time),
		globalTimer: time.Now(),
		mu:          &sync.RWMutex{},
	}

	duplicateDetector := &DuplicateDetector{
		requests: make(map[string]*DuplicateInfo),
		mu:       &sync.RWMutex{},
		ttl:      15 * time.Minute,
	}

	alertsSvc := alerts.NewAlertsService(ctx, slackBotToken)
	alertsSvc.AddChannel("request-logs", requestLogsChannelID)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			var bodyContent string

			if r.Body != nil && r.Header.Get("Content-Type") == "application/json" {
				bodyBytes, err := io.ReadAll(r.Body)
				if err == nil {
					bodyContent = string(bodyBytes)
					r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				}
			}

			sanitizedBody := sanitizeJSON(bodyContent)
			sanitizedQuery := sanitizeMap(r.URL.Query())
			sanitizedHeaders := sanitizeMap(r.Header)

			responseWriter := &ResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(responseWriter, r)

			duration := time.Since(start)

			go func() {
				if responseWriter.statusCode < 400 {
					return
				}

				requestKey := generateRequestKey(r, responseWriter.statusCode)
				shouldReport, count := duplicateDetector.ShouldReport(requestKey)

				if !shouldReport {
					return
				}

				if rateLimiter.CanSend("request-logs") {
					go func() {
						logRequestToAlerts(ctx, alertsSvc, r, responseWriter.statusCode, duration, sanitizedBody, sanitizedHeaders, sanitizedQuery, serviceName, count)
					}()
				}
			}()
		})
	}
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
	key := fmt.Sprintf("%s:%s:%d", r.Method, r.URL.Path, statusCode)

	importantParams := []string{"user_id", "session_id", "id"}
	for _, param := range importantParams {
		if value := r.URL.Query().Get(param); value != "" {
			key += fmt.Sprintf(":%s=%s", param, value)
		}
	}

	if routeCtx := chi.RouteContext(r.Context()); routeCtx != nil {
		for i, paramKey := range routeCtx.URLParams.Keys {
			if i < len(routeCtx.URLParams.Values) {
				key += fmt.Sprintf(":%s=%s", paramKey, routeCtx.URLParams.Values[i])
			}
		}
	}

	return key
}

var sensitiveFields = []string{"password", "token", "authorization"}

func sanitizeMap(data map[string][]string) map[string][]string {
	sanitized := make(map[string][]string, len(data))
	for key, values := range data {
		if isSensitiveField(key) {
			sanitized[key] = []string{"[REDACTED]"}
		} else {
			sanitized[key] = values
		}
	}

	return sanitized
}

func sanitizeJSON(body string) string {
	if body == "" {
		return body
	}

	var data interface{}
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		return body
	}

	sanitizeInterface(data)

	result, err := json.Marshal(data)
	if err != nil {
		return body
	}

	return string(result)
}

func sanitizeInterface(v interface{}) {
	switch val := v.(type) {
	case map[string]interface{}:
		for k, v2 := range val {
			if isSensitiveField(k) {
				val[k] = "[REDACTED]"
			} else {
				sanitizeInterface(v2)
			}
		}
	case []interface{}:
		for _, item := range val {
			sanitizeInterface(item)
		}
	}
}

func isSensitiveField(key string) bool {
	key = strings.ToLower(key)
	for _, f := range sensitiveFields {
		if key == f {
			return true
		}
	}

	return false
}

func logRequestToAlerts(ctx context.Context, alertsSvc *alerts.Service, r *http.Request, statusCode int, duration time.Duration, bodyContent string, headers, query map[string][]string, serviceName string, count int) {
	routeCtx := chi.RouteContext(ctx)

	var routeParams []string

	if routeCtx != nil {
		for i, key := range routeCtx.URLParams.Keys {
			if i < len(routeCtx.URLParams.Values) {
				routeParams = append(routeParams, fmt.Sprintf("%s: %s", key, routeCtx.URLParams.Values[i]))
			}
		}
	}

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
		bodyPreview = truncateString(bodyContent, 1000)
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

	message := fmt.Sprintf(`--------------------------------

%s HTTP Request%s

🏢 Service: %s
📱 Method: %s
🌐 URL: %s
📊 Status: %d
⏱️ Duration: %v
👤 User Agent: %s
🌍 Remote IP: %s

📋 Route Parameters:
%s

🔍 Query Parameters:
%s

📬 Headers:
%s

📝 Body Preview:
%s

📅 Timestamp: %s`,
		statusIcon,
		countInfo,
		serviceName,
		r.Method,
		r.URL.String(),
		statusCode,
		duration,
		truncateString(r.UserAgent(), 100),
		getRemoteIP(r),
		formatList(routeParams, "• "),
		formatList(queryParams, "• "),
		formatList(headerParams, "• "),
		bodyPreview,
		time.Now().Format("2006-01-02 15:04:05 UTC"),
	)

	if err := alertsSvc.SendMessageImmediate(ctx, "request-logs", message); err != nil {
		logger.Global().Error("Failed to send alerts request log", logger.Error(err))
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	return s[:maxLen-3] + "..."
}

func getRemoteIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}

	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	return r.RemoteAddr
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
