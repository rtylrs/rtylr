package http

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// RouteRateLimitMetaKey is the @meta key OneKit routes use to declare a
// per-client budget, e.g. @meta("rate_limit", "10/m").
const RouteRateLimitMetaKey = "rate_limit"

// ParseRouteRateLimit reads a "<count>/<s|m|h>" budget such as "10/m".
func ParseRouteRateLimit(spec string) (int, time.Duration, error) {
	count, unit, ok := strings.Cut(strings.TrimSpace(spec), "/")
	if !ok {
		return 0, 0, fmt.Errorf("rate limit %q: want <count>/<s|m|h>", spec)
	}
	limit, err := strconv.Atoi(count)
	if err != nil || limit <= 0 {
		return 0, 0, fmt.Errorf("rate limit %q: count must be a positive integer", spec)
	}
	window, ok := map[string]time.Duration{"s": time.Second, "m": time.Minute, "h": time.Hour}[unit]
	if !ok {
		return 0, 0, fmt.Errorf("rate limit %q: unit must be s, m or h", spec)
	}
	return limit, window, nil
}

// RouteRateLimitMiddleware enforces each OneKit route's @meta("rate_limit")
// budget per route and client, with the same 429 response as PerRouteRateLimit.
// routeLimit reports the route and its declared budget; generated packages each
// have their own RequestMetadata type, so the caller reads it.
func RouteRateLimitMiddleware(routeLimit func(*http.Request) (route, spec string)) func(http.Handler) http.Handler {
	var mu sync.Mutex
	limiters := map[string]func(http.Handler) http.Handler{}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			route, spec := routeLimit(r)
			if spec == "" {
				next.ServeHTTP(w, r)
				return
			}
			mu.Lock()
			limited, ok := limiters[route]
			if !ok {
				limit, window, err := ParseRouteRateLimit(spec)
				if err != nil {
					mu.Unlock()
					// A malformed budget is a schema bug; fail closed rather than
					// silently serving an unlimited public route.
					http.Error(w, "route rate limit misconfigured", http.StatusInternalServerError)
					return
				}
				limited = PerRouteRateLimit(limit, window)
				limiters[route] = limited
			}
			mu.Unlock()
			limited(next).ServeHTTP(w, r)
		})
	}
}
