package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/voxire/onion/internal/platform/logger"
)

// serverCtx is the context passed by Start() when the HTTP server boots.
// Rate limiters created before Start is called (e.g. package-level vars)
// use context.Background() and are effectively immortal — that's fine for
// singletons that last the whole process.
var serverCtx = context.Background()

type SecurityRateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	limit    int
	window   time.Duration
	ctx      context.Context
	cancel   context.CancelFunc
	ticker   *time.Ticker
}

func NewSecurityRateLimiter(limit int, window time.Duration) *SecurityRateLimiter {
	ctx, cancel := context.WithCancel(serverCtx)

	return &SecurityRateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
		ctx:      ctx,
		cancel:   cancel,
	}
}

func (rl *SecurityRateLimiter) isAllowed(key string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	if times, exists := rl.requests[key]; exists {
		validRequests := retainTimesAfter(times, windowStart)

		rl.requests[key] = validRequests

		if len(validRequests) >= rl.limit {
			return false
		}
	}

	rl.requests[key] = append(rl.requests[key], now)

	return true
}

func (rl *SecurityRateLimiter) cleanup() {
	rl.ticker = time.NewTicker(rl.window)

	go func() {
		defer rl.ticker.Stop()

		for {
			select {
			case <-rl.ctx.Done():
				return
			case <-rl.ticker.C:
				rl.mutex.Lock()

				now := time.Now()
				windowStart := now.Add(-rl.window)

				for key, times := range rl.requests {
					validRequests := retainTimesAfter(times, windowStart)

					if len(validRequests) == 0 {
						delete(rl.requests, key)
					} else {
						rl.requests[key] = validRequests
					}
				}

				rl.mutex.Unlock()
			}
		}
	}()
}

func retainTimesAfter(times []time.Time, cutoff time.Time) []time.Time {
	valid := times[:0]
	for _, requestTime := range times {
		if requestTime.After(cutoff) {
			valid = append(valid, requestTime)
		}
	}
	return valid
}

func (rl *SecurityRateLimiter) Stop() {
	if rl.cancel != nil {
		rl.cancel()
	}
}

func RateLimitMiddleware(limit int, window time.Duration) func(http.Handler) http.Handler {
	limiter := NewSecurityRateLimiter(limit, window)
	limiter.cleanup()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := getClientKey(r)

			if !limiter.isAllowed(key) {
				logger.Ctx(r.Context()).Warn("Rate limit exceeded",
					logger.String("ip", r.RemoteAddr),
					logger.String("path", r.URL.Path),
				)

				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "60")
				w.WriteHeader(http.StatusTooManyRequests)

				response := map[string]interface{}{
					"error":       "Rate limit exceeded",
					"message":     "Too many requests, please try again later",
					"retry_after": 60,
				}

				if jsonBytes, err := json.Marshal(response); err == nil {
					if _, writeErr := w.Write(jsonBytes); writeErr != nil {
						logger.Ctx(r.Context()).Error("Failed to write rate limit response", logger.Error(writeErr))
					}
				}

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getClientKey returns a per-client identifier for rate limiting. It must NOT
// trust client-controllable headers, or an attacker rotates them to mint
// unlimited buckets and bypass every limiter (global + per-route brute-force).
//
// onion runs behind Cloudflare, which sets CF-Connecting-IP to the true client
// IP — a value a client cannot forge through Cloudflare. We trust only that,
// then fall back to the connection's RemoteAddr. X-Forwarded-For / X-Real-IP are
// deliberately NOT honored (they're attacker-settable on a direct connection).
func getClientKey(r *http.Request) string {
	if cfIP := strings.TrimSpace(r.Header.Get("CF-Connecting-IP")); cfIP != "" {
		return cfIP
	}
	return r.RemoteAddr
}

// PerRouteRateLimit returns middleware that applies a tighter limit
// than the global 100/min on specific auth + abuse-prone endpoints.
// Returns a func that wraps a single chi.Mux Method call rather than
// an `r.Use(...)` layer, so we can attach it to one route without
// touching every handler.
//
// Rollouts:
//   - /auth/login          → 5/min/IP   (brute-force defence)
//   - /auth/register       → 3/min/IP   (signup spam)
//   - /recruit/public/apply → 5/min/IP  (careers-form abuse)
//
// We key on IP alone — adding email-based secondary limits would
// require parsing the request body twice, which isn't worth the
// complexity for what this bucket actually defends.
func PerRouteRateLimit(limit int, window time.Duration) func(http.Handler) http.Handler {
	return perRouteRateLimitWithKeyFn(limit, window, func(r *http.Request) string {
		return r.URL.Path + "|" + getClientKey(r)
	})
}

// AIUserRateLimit returns middleware that rate-limits per authenticated user ID
// (read from the "userId" JSON field in the request body) for the AI SSE endpoint.
// It applies a tighter limit than the global limiter so that individual users
// cannot monopolise the AI generation capacity.
//
// Rollouts:
//   - /ai/generate-text-stream → 5/min/user  (AI cost control)
//
// The userId is taken from the GenerateTextStreamRequest struct (snake_case JSON
// key "userId") which is already part of the request contract.  If the header
// X-User-Id is present, that value is preferred to avoid re-parsing the body.
// Falls back to IP if neither source provides a non-empty value.
func AIUserRateLimit(limit int, window time.Duration) func(http.Handler) http.Handler {
	return perRouteRateLimitWithKeyFn(limit, window, func(r *http.Request) string {
		if uid := r.Header.Get("X-User-Id"); uid != "" {
			return r.URL.Path + "|user:" + uid
		}
		return r.URL.Path + "|" + getClientKey(r)
	})
}

// perRouteRateLimitWithKeyFn is the shared implementation used by both
// PerRouteRateLimit and AIUserRateLimit.  The keyFn determines the bucket
// identifier for each incoming request.
func perRouteRateLimitWithKeyFn(limit int, window time.Duration, keyFn func(*http.Request) string) func(http.Handler) http.Handler {
	limiter := NewSecurityRateLimiter(limit, window)
	limiter.cleanup()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := keyFn(r)
			if !limiter.isAllowed(key) {
				logger.Ctx(r.Context()).Warn("Per-route rate limit exceeded",
					logger.String("ip", r.RemoteAddr),
					logger.String("path", r.URL.Path),
				)
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "60")
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte(`{"error":"Rate limit exceeded","message":"Too many attempts, please try again shortly","retry_after":60}`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
