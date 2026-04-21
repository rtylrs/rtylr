package http

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/voxire/onion/internal/platform/logger"
)

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
	ctx, cancel := context.WithCancel(context.Background())

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
		var validRequests []time.Time

		for _, t := range times {
			if t.After(windowStart) {
				validRequests = append(validRequests, t)
			}
		}

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
					var validRequests []time.Time

					for _, t := range times {
						if t.After(windowStart) {
							validRequests = append(validRequests, t)
						}
					}

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

func getClientKey(r *http.Request) string {
	ip := r.RemoteAddr
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		ip = forwardedFor
	}

	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		ip = realIP
	}

	return ip
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
	limiter := NewSecurityRateLimiter(limit, window)
	limiter.cleanup()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Path + "|" + getClientKey(r)
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
