package http

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

var authRateLimitTestSequence atomic.Uint64

// The route limiters are package globals. Give every test invocation its own
// client key so -count=2 does not inherit the first invocation's budget.
func authRateLimitTestRemoteAddr() string {
	n := authRateLimitTestSequence.Add(1)
	return fmt.Sprintf("198.18.%d.%d:%d", (n>>8)&0xff, n&0xff, 10000+(n%50000))
}

func TestAuthRouteRateLimit_LoginBudget(t *testing.T) {
	remoteAddr := authRateLimitTestRemoteAddr()
	handler := AuthRouteRateLimit(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	const path = "/common/auth/v1/login"
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodPost, path, nil)
		req.RemoteAddr = remoteAddr
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: status = %d, want 200", i+1, rec.Code)
		}
	}

	req := httptest.NewRequest(http.MethodPost, path, nil)
	req.RemoteAddr = remoteAddr
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("6th login from same IP: status = %d, want 429", rec.Code)
	}
}

func TestAuthRouteRateLimit_OtherPathsUnaffected(t *testing.T) {
	remoteAddr := authRateLimitTestRemoteAddr()
	handler := AuthRouteRateLimit(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 50; i++ {
		req := httptest.NewRequest(http.MethodGet, "/common/auth/v1/email-exists", nil)
		req.RemoteAddr = remoteAddr
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d to unlisted path: status = %d, want 200", i+1, rec.Code)
		}
	}
}

func TestAuthRouteRateLimit_RegisterBudget(t *testing.T) {
	remoteAddr := authRateLimitTestRemoteAddr()
	handler := AuthRouteRateLimit(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	const path = "/common/auth/v1/register"
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodPost, path, nil)
		req.RemoteAddr = remoteAddr
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: status = %d, want 200", i+1, rec.Code)
		}
	}

	req := httptest.NewRequest(http.MethodPost, path, nil)
	req.RemoteAddr = remoteAddr
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("4th register from same IP: status = %d, want 429", rec.Code)
	}
}

func TestAuthRouteRateLimit_DirectPeerCannotRotateCFHeader(t *testing.T) {
	t.Setenv("TRUSTED_PROXY_CIDRS", "203.0.113.0/24")
	remoteAddr := authRateLimitTestRemoteAddr()
	handler := AuthRouteRateLimit(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	const path = "/common/auth/v1/login"
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodPost, path, nil)
		req.RemoteAddr = remoteAddr
		req.Header.Set("CF-Connecting-IP", fmt.Sprintf("198.51.100.%d", i+1))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: status = %d, want 200", i+1, rec.Code)
		}
	}

	req := httptest.NewRequest(http.MethodPost, path, nil)
	req.RemoteAddr = remoteAddr
	req.Header.Set("CF-Connecting-IP", "198.51.100.200")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("sixth direct request = %d, want 429", rec.Code)
	}
}
