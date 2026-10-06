package http

import (
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/logger"
)

// Rate limiters created before Start is called (e.g. package-level vars)
// use context.Background() and are effectively immortal — that's fine for
// singletons that last the whole process.
var serverCtx = context.Background()

var (
	rateLimiterMu      sync.RWMutex
	activeRateLimiters = make(map[*SecurityRateLimiter]struct{})
)

var rateLimitExceededJSON = []byte(`{"error":"Rate limit exceeded","message":"Too many requests, please try again later","retry_after":60}`)

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

	rl := &SecurityRateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
		ctx:      ctx,
		cancel:   cancel,
	}

	rateLimiterMu.Lock()
	activeRateLimiters[rl] = struct{}{}
	rateLimiterMu.Unlock()

	return rl
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

	rateLimiterMu.Lock()
	delete(activeRateLimiters, rl)
	rateLimiterMu.Unlock()
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
				if _, writeErr := io.Copy(w, bytes.NewReader(rateLimitExceededJSON)); writeErr != nil {
					logger.Ctx(r.Context()).Error("Failed to write rate limit response", logger.Error(writeErr))
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
// CF-Connecting-IP is honored only when the socket peer is in the configured
// Cloudflare/proxy CIDR set. Directly reachable API listeners,
// so a client can otherwise rotate this header and receive a fresh bucket on
// every request. X-Forwarded-For / X-Real-IP are deliberately not honored.
func getClientKey(r *http.Request) string {
	peer := peerIP(r.RemoteAddr)
	if config.IsTrustedProxyIP(r.RemoteAddr) {
		if cfIP := strings.TrimSpace(r.Header.Get("CF-Connecting-IP")); net.ParseIP(cfIP) != nil {
			return cfIP
		}
	}
	return peer
}

func peerIP(remoteAddr string) string {
	if host, _, err := net.SplitHostPort(remoteAddr); err == nil {
		return host
	}
	return strings.Trim(remoteAddr, "[]")
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

// AIUserRateLimit returns middleware that rate-limits per authenticated user for
// the AI SSE endpoint. Falls back to IP when the request is not authenticated.
//
// Rollouts:
//   - /ai/generate-text-stream → 5/min/user  (AI cost control)
func AIUserRateLimit(limit int, window time.Duration) func(http.Handler) http.Handler {
	return perRouteRateLimitWithKeyFn(limit, window, func(r *http.Request) string {
		if authUser, ok := authPkg.GetAuthUser(r.Context()); ok && authUser.UserID > 0 {
			return r.URL.Path + "|user:" + strconv.FormatInt(authUser.UserID, 10)
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
				if _, writeErr := w.Write([]byte(`{"error":"Rate limit exceeded","message":"Too many attempts, please try again shortly","retry_after":60}`)); writeErr != nil {
					logger.Ctx(r.Context()).Error("Failed to write per-route rate limit response", logger.Error(writeErr))
				}
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// authRouteLimiters pins the anonymous and credential-bearing routes to tight
// per-IP budgets, closing the gap between the 1000 req/s global guardrail and
// what a credential endpoint may absorb. Budgets follow the rollout table on
// PerRouteRateLimit: login and password-reset confirm allow 5/min/IP, register
// variants 3/min/IP. Password-reset request and invite acceptance sit at
// 5/min because both mint account-access tokens.
//
// Short-secret endpoints get budgets sized to the secret's search space, not
// to a single copy-pasted 5. At 5/min an attacker spends 12 attempts/hour/IP
// on a 6-digit code; the whole 1e6 space needs ~833 hours from one IP, which a
// botnet of a few thousand hosts finishes inside the 10-minute code TTL. The
// 1e6-space entries are therefore cut to 3/min and the 1e4-space (POS PIN)
// entries to 5/min, so the sweep is bounded above the TTL at any realistic
// botnet size. None of this is a hard bound — see the limitation note below.
//
// Limitations, stated plainly rather than implied away:
//   - The key is the client IP only. A distributed attacker rotates source
//     addresses and gets a fresh bucket per address, so the effective ceiling
//     is limit x (number of distinct IPs), not limit.
//   - Buckets are in-process (map + mutex on SecurityRateLimiter). With N
//     replicas behind a load balancer each replica allows the full budget, so
//     the real ceiling is limit x N for a client that spreads requests across
//     them. Nothing here is shared state; there is no Redis bucket.
//   - getClientKey only honours CF-Connecting-IP from a trusted proxy CIDR, so
//     a directly-reachable listener cannot be given fresh buckets by
//     rotating a header — but it also means every client behind one corporate
//     NAT or CGNAT egress shares a single bucket.
var authRouteLimiters = map[string]func(http.Handler) http.Handler{
	// --- account credentials ---
	"/common/auth/v1/login":                  PerRouteRateLimit(5, time.Minute),
	"/common/auth/v1/register":               PerRouteRateLimit(3, time.Minute),
	"/common/auth/v1/register/google":        PerRouteRateLimit(3, time.Minute),
	"/common/auth/v1/refresh":                PerRouteRateLimit(30, time.Minute),
	"/common/auth/v1/revoke":                 PerRouteRateLimit(30, time.Minute),
	"/common/auth/v1/password-reset/request": PerRouteRateLimit(5, time.Minute),
	"/common/auth/v1/password-reset/confirm": PerRouteRateLimit(5, time.Minute),
	"/common/auth/v1/invite/accept":          PerRouteRateLimit(5, time.Minute),
	// The HR invite routes, which were removed from the anonymous route list and
	// so now run behind the bearer gate. Both take a 72h invite token disclosed
	// to the inviter in the creation response, and accept is the step that binds
	// an account to a business. The token is a uuid, so this is not a search-
	// space bound and the single-use plus recipient checks are the real defence;
	// 5/min is a cheap brake on a stolen-token replay loop that would otherwise
	// run at the global guardrail, and it matches the account invite above.
	"/hr/employees/invite/accept": PerRouteRateLimit(5, time.Minute),
	"/hr/employees/invite/verify": PerRouteRateLimit(5, time.Minute),

	// --- short secrets, sized to their search space ---
	// Anonymous and unauthenticated: a code here flips user_provisioning_states
	// and activates authorization_business_members, so a sweep is an
	// account-takeover primitive, not just a nuisance. 3/min on a 1e6 code.
	"/common/auth/v1/verify-email": PerRouteRateLimit(3, time.Minute),
	// 4-digit POS PIN (1e4 space) but authenticated and location-scoped, and a
	// sibling lockout covers the per-account dimension. 5/min keeps the whole
	// space at ~3.3 hours/IP, so a sweep needs a large fleet; the per-account
	// lockout, not this bucket, is the real control there.
	"/pos/user/v1/users/verify-pin": PerRouteRateLimit(5, time.Minute),
	// Manager override PIN, same 1e4 floor as above: PINs issued under the
	// pre-policy 4-8 digit rule still verify, so 1e4 is the real space to sweep
	// even though new ones must be 6+. The two POS PIN surfaces must not end up
	// with different budgets. The durable per-business counter is the control
	// that survives a bucket reset; this is the per-IP brake in front of it.
	"/pos/manager/v1/verify": PerRouteRateLimit(5, time.Minute),
	// 6-digit TOTP with a ±1 step window, and success mints a 5-minute
	// verified state (team repository_authorization.go IsMFAVerified). Two
	// codes are live per 30s step, so 5/min would cost an attacker nothing
	// while locking out a user with a clumsy authenticator. 10/min still caps
	// one IP at 14400 attempts/day against a 1e6 space.
	"/hub/security/v1/mfa/verify":   PerRouteRateLimit(10, time.Minute),
	"/hub/security/v1/mfa/recovery": PerRouteRateLimit(10, time.Minute),

	// --- anonymous oracles, not brute force ---
	// Unauthenticated user-enumeration: {"exists":true} for a known address
	// is a password-spray oracle, so the budget is the enumeration budget, not
	// a credential one. 20/min still confirms one targeted address quickly and
	// is too slow (28,800/day/IP) for a bulk dump.
	"/common/auth/v1/email-exists": PerRouteRateLimit(20, time.Minute),
	// Anonymous presigned-URL mint: each call is a free R2 upload slot plus a
	// storage-abuse lever, and unlike SubmitApplication (5/min) it had no
	// per-route limit at all. The legitimate flow is a handful of presigns
	// during one application, so 10/min leaves generous headroom.
	"/recruit/public/resume-upload-url": PerRouteRateLimit(10, time.Minute),

	// --- public read surfaces (unchanged) ---
	"/pos/orders/v1/orders/public":                   PerRouteRateLimit(10, time.Minute),
	"/pos/orders/v1/orders/{order_id}/public-status": PerRouteRateLimit(60, time.Minute),
	"/pos/menu-item/v1/menu-items/public":            PerRouteRateLimit(60, time.Minute),
}

// AuthRouteRateLimit applies the per-route auth budgets by exact request
// path and passes everything else through untouched. It runs after the global
// limiter in the middleware stack so authenticated traffic never pays for it.
func AuthRouteRateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if limiter, ok := authRouteLimiters[r.URL.Path]; ok {
			limiter(next).ServeHTTP(w, r)
			return
		}
		for route, limiter := range authRouteLimiters {
			if routePatternMatchesPath(route, r.URL.Path) {
				limiter(next).ServeHTTP(w, r)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
