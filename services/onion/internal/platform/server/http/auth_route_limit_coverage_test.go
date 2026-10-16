package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// hit drives AuthRouteRateLimit from one client key and reports how many
// requests the downstream handler saw before the limiter answered 429. It
// returns (allowed, throttledAt) where throttledAt is the 1-based request
// number that first got 429, or 0 if the whole burst was allowed.
func hitBurst(t *testing.T, path string, burst int) (allowed, throttledAt int) {
	t.Helper()

	remoteAddr := authRateLimitTestRemoteAddr()
	allowed = 0
	throttledAt = 0

	handler := AuthRouteRateLimit(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 1; i <= burst; i++ {
		req := httptest.NewRequest(http.MethodPost, path, nil)
		req.RemoteAddr = remoteAddr
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		switch rec.Code {
		case http.StatusOK:
			allowed++
		case http.StatusTooManyRequests:
			if throttledAt == 0 {
				throttledAt = i
			}
		default:
			t.Fatalf("path %s request %d: status = %d, want 200 or 429", path, i, rec.Code)
		}
	}

	return allowed, throttledAt
}

// Every route the table claims to protect must actually be in the table with
// the budget documented for it. Before the coverage change all five of these
// paths were absent, so a full burst from one client was served every time.
func TestAuthRouteRateLimit_ShortSecretAndOracleRoutesAreLimited(t *testing.T) {
	for _, tc := range []struct {
		path      string
		wantAllow int
	}{
		// 6-digit code, anonymous, flips provisioning + membership state.
		{path: "/common/auth/v1/verify-email", wantAllow: 3},
		// Unauthenticated user-enumeration oracle.
		{path: "/common/auth/v1/email-exists", wantAllow: 20},
		// 4-digit POS PIN.
		{path: "/pos/user/v1/users/verify-pin", wantAllow: 5},
		// Anonymous OAuth exchange and revocation, and the authorize redirect.
		{path: "/common/oauth/v1/token", wantAllow: 30},
		{path: "/common/oauth/v1/revoke", wantAllow: 30},
		{path: "/oauth/authorize", wantAllow: 30},
		// Current-password check behind a bearer token.
		{path: "/hub/user/v1/users/password", wantAllow: 5},
		// 6-digit TOTP, ±1 window, mints a 5-minute verified state.
		{path: "/hub/security/v1/mfa/verify", wantAllow: 10},
		{path: "/hub/security/v1/mfa/recovery", wantAllow: 10},
		// Anonymous presigned-URL mint.
		{path: "/recruit/public/resume-upload-url", wantAllow: 10},
	} {
		allowed, throttledAt := hitBurst(t, tc.path, tc.wantAllow+5)

		if allowed != tc.wantAllow {
			t.Errorf("%s: allowed %d requests in a burst of %d, want %d",
				tc.path, allowed, tc.wantAllow+5, tc.wantAllow)
		}
		if throttledAt != tc.wantAllow+1 {
			t.Errorf("%s: first 429 at request %d, want %d", tc.path, throttledAt, tc.wantAllow+1)
		}
	}
}

// A legitimate flow must not be broken by the new budgets. Signing up and
// verifying an address is a handful of calls from one address; a burst at the
// exact budget, and one past it, is the boundary that matters.
func TestAuthRouteRateLimit_LegitimateBurstIsNotBroken(t *testing.T) {
	cases := []struct {
		path      string
		wantAllow int
	}{
		{path: "/common/auth/v1/verify-email", wantAllow: 3},
		{path: "/pos/user/v1/users/verify-pin", wantAllow: 5},
		{path: "/hub/security/v1/mfa/verify", wantAllow: 10},
		{path: "/hub/security/v1/mfa/recovery", wantAllow: 10},
		{path: "/recruit/public/resume-upload-url", wantAllow: 10},
		{path: "/common/auth/v1/email-exists", wantAllow: 20},
	}

	for _, tc := range cases {
		allowed, throttledAt := hitBurst(t, tc.path, tc.wantAllow)

		if allowed != tc.wantAllow || throttledAt != 0 {
			t.Errorf("%s: a burst of exactly %d got %d allowed and first 429 at %d, want all %d allowed",
				tc.path, tc.wantAllow, allowed, throttledAt, tc.wantAllow)
		}
	}
}

// The bucket is per path and per client. Tightening one route must not spend
// another route's budget, and one abusive client must not throttle a different
// client — otherwise a single attacker could lock out a whole office behind
// one NAT egress.
func TestAuthRouteRateLimit_BucketsAreIsolatedPerPathAndClient(t *testing.T) {
	burnEmailExists := authRateLimitTestRemoteAddr()
	handler := AuthRouteRateLimit(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	post := func(path, remoteAddr string) int {
		req := httptest.NewRequest(http.MethodPost, path, nil)
		req.RemoteAddr = remoteAddr
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec.Code
	}

	// Exhaust one client's email-exists budget.
	for i := 0; i < 20; i++ {
		if code := post("/common/auth/v1/email-exists", burnEmailExists); code != http.StatusOK {
			t.Fatalf("setup: request %d to email-exists: status = %d, want 200", i+1, code)
		}
	}
	if code := post("/common/auth/v1/email-exists", burnEmailExists); code != http.StatusTooManyRequests {
		t.Fatalf("setup: 21st email-exists request: status = %d, want 429", code)
	}

	// A different client keeps its own budget on the same path.
	if code := post("/common/auth/v1/email-exists", authRateLimitTestRemoteAddr()); code != http.StatusOK {
		t.Errorf("a second client on the throttled path: status = %d, want 200", code)
	}

	// The abusive client keeps its budget on the other short-secret routes.
	for _, path := range []string{"/common/auth/v1/verify-email", "/hub/security/v1/mfa/verify", "/hub/security/v1/mfa/recovery"} {
		if code := post(path, burnEmailExists); code != http.StatusOK {
			t.Errorf("same client, different path %s: status = %d, want 200", path, code)
		}
	}
}

// CF-Connecting-IP is only honoured from a trusted proxy, so a directly
// reachable listener cannot mint a fresh bucket per request on the new
// short-secret routes either. This is the property that makes an IP-keyed
// budget meaningful at all against a single attacker.
func TestAuthRouteRateLimit_ShortSecretRouteIgnoresRotatedProxyHeader(t *testing.T) {
	t.Setenv("TRUSTED_PROXY_CIDRS", "203.0.113.0/24")
	remoteAddr := authRateLimitTestRemoteAddr()

	handler := AuthRouteRateLimit(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	const path = "/common/auth/v1/verify-email"
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodPost, path, nil)
		req.RemoteAddr = remoteAddr
		req.Header.Set("X-Forwarded-For", "10.0.0.9")
		req.Header.Set("X-Real-IP", "10.0.0.10")
		req.Header.Set("CF-Connecting-IP", "198.51.100.7")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: status = %d, want 200", i+1, rec.Code)
		}
	}

	req := httptest.NewRequest(http.MethodPost, path, nil)
	req.RemoteAddr = remoteAddr
	req.Header.Set("CF-Connecting-IP", "198.51.100.250")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("4th request with a fresh spoofed CF header: status = %d, want 429", rec.Code)
	}
}
