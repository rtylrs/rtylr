package http

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
)

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
	"referer",
	"referrer",
	"secret",
	"api_key",
	"apikey",
	"private_key",
	"credential",
	"cookie",
	"email",
	"phone",
	"address",
	"birthdate",
	"date_of_birth",
	"name",
	"pin",
	"cvv",
	"cvc",
	"ssn",
	"tax_id",
	"national_id",
	"otp",
	"card_number",
	"expiry",
	"passwd",
	"pwd",
	"x-api-key",
	"sessionid",
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

func isSensitiveField(key string) bool {
	key = strings.ToLower(key)
	for _, f := range sensitiveFieldParts {
		if strings.Contains(key, f) {
			return true
		}
	}

	return key == "session_id" || strings.HasSuffix(key, "_session_id")
}
