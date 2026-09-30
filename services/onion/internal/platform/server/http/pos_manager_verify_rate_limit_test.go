package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// /pos/manager/v1/verify is a PIN endpoint with a durable per-business
// attempt counter behind it. The counter bounds one business; this bucket is
// the per-IP brake in front of it, so it has to exist and fire on its own.
func TestPosManagerVerifyIsRateLimited(t *testing.T) {
	const path = "/pos/manager/v1/verify"

	allowed, throttledAt := hitBurst(t, path, 10)
	if allowed != 5 || throttledAt != 6 {
		t.Fatalf("%s: allowed %d, first 429 at %d; want 5 allowed and 429 at 6", path, allowed, throttledAt)
	}
}

// A manager verifying their own PIN to approve a void, a refund or a drawer
// open does it a few times in a shift, not a few times a minute. A burst at
// exactly the budget must pass.
func TestPosManagerVerifyAllowsTheBudgetBeforeThrottling(t *testing.T) {
	const path = "/pos/manager/v1/verify"

	allowed, throttledAt := hitBurst(t, path, 5)
	if allowed != 5 || throttledAt != 0 {
		t.Fatalf("%s: a burst of exactly 5 got %d allowed and first 429 at %d; want all 5 allowed",
			path, allowed, throttledAt)
	}
}

// The three routes added for the HR invite acceptance and the POS manager PIN
// are separate buckets. One exhausted route must not spend another's budget, and
// one client must not lock out a different client on the same route - otherwise
// a single abusive client behind one office egress denies the whole office.
func TestNewCredentialRouteLimitsAreIsolatedPerPathAndClient(t *testing.T) {
	paths := []string{
		"/hr/employees/invite/accept",
		"/hr/employees/invite/verify",
		"/pos/manager/v1/verify",
	}

	for _, burned := range paths {
		t.Run(burned, func(t *testing.T) {
			abusive := authRateLimitTestRemoteAddr()
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

			for i := 0; i < 5; i++ {
				if code := post(burned, abusive); code != http.StatusOK {
					t.Fatalf("setup: request %d to %s: status = %d, want 200", i+1, burned, code)
				}
			}
			if code := post(burned, abusive); code != http.StatusTooManyRequests {
				t.Fatalf("setup: 6th request to %s: status = %d, want 429", burned, code)
			}

			if code := post(burned, authRateLimitTestRemoteAddr()); code != http.StatusOK {
				t.Errorf("a second client on the throttled route %s: status = %d, want 200", burned, code)
			}

			for _, other := range paths {
				if other == burned {
					continue
				}
				if code := post(other, abusive); code != http.StatusOK {
					t.Errorf("abusive client, different route %s: status = %d, want 200", other, code)
				}
			}
		})
	}
}
