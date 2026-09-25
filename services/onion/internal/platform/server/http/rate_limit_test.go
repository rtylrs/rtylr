package http

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func TestSecurityRateLimiter_AllowsUnderLimit(t *testing.T) {
	rl := NewSecurityRateLimiter(3, time.Second)
	defer rl.Stop()

	for i := 0; i < 3; i++ {
		if !rl.isAllowed("1.2.3.4") {
			t.Fatalf("request %d should be allowed (under limit)", i+1)
		}
	}
}

func TestSecurityRateLimiter_BlocksOverLimit(t *testing.T) {
	rl := NewSecurityRateLimiter(2, time.Second)
	defer rl.Stop()

	rl.isAllowed("ip-a")
	rl.isAllowed("ip-a")
	if rl.isAllowed("ip-a") {
		t.Fatal("expected third request to be blocked")
	}

	// Different key should not be affected.
	if !rl.isAllowed("ip-b") {
		t.Fatal("different key should be unaffected by another key's limit")
	}
}

func TestSecurityRateLimiter_AllowsAfterWindow(t *testing.T) {
	rl := NewSecurityRateLimiter(1, 50*time.Millisecond)
	defer rl.Stop()

	if !rl.isAllowed("k") {
		t.Fatal("first request should be allowed")
	}
	if rl.isAllowed("k") {
		t.Fatal("second request inside window should be blocked")
	}

	time.Sleep(75 * time.Millisecond)
	if !rl.isAllowed("k") {
		t.Fatal("after window, request should be allowed again")
	}
}

func TestRateLimitMiddleware_Returns429(t *testing.T) {
	mw := RateLimitMiddleware(1, 10*time.Second)
	called := 0
	wrapped := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called++
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/x", http.NoBody)
		req.RemoteAddr = "1.1.1.1:" + strconv.Itoa(1234+i)
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		switch i {
		case 0:
			if rec.Code != http.StatusOK {
				t.Errorf("first request: want 200, got %d", rec.Code)
			}
		default:
			if rec.Code != http.StatusTooManyRequests {
				t.Errorf("request %d: want 429, got %d", i+1, rec.Code)
			}
			if rec.Header().Get("Retry-After") != "60" {
				t.Errorf("expected Retry-After header")
			}
		}
	}
	if called != 1 {
		t.Errorf("downstream handler should run only once, ran %d times", called)
	}
}

func TestGetClientKey_TrustsOnlyCloudflare(t *testing.T) {
	t.Setenv("TRUSTED_PROXY_CIDRS", "10.0.0.0/8")
	r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	r.RemoteAddr = "10.0.0.1:1234"
	if got := getClientKey(r); got != "10.0.0.1" {
		t.Errorf("default to RemoteAddr, got %q", got)
	}

	// Attacker-settable headers must NOT change the bucket (else they can rotate
	// them to mint unlimited buckets and bypass the limiters).
	r.Header.Set("X-Forwarded-For", "spoofed.example, proxy.example")
	r.Header.Set("X-Real-IP", "9.9.9.9")
	if got := getClientKey(r); got != "10.0.0.1" {
		t.Errorf("X-Forwarded-For/X-Real-IP must be ignored, got %q", got)
	}

	// CF-Connecting-IP (set by Cloudflare, unspoofable through it) is trusted.
	r.Header.Set("CF-Connecting-IP", "8.8.8.8")
	if got := getClientKey(r); got != "8.8.8.8" {
		t.Errorf("CF-Connecting-IP should win, got %q", got)
	}

	// A directly reachable self-host listener must ignore a rotated header even
	// when the value looks like a valid public IP.
	t.Setenv("TRUSTED_PROXY_CIDRS", "203.0.113.0/24")
	r.RemoteAddr = "198.51.100.10:4321"
	r.Header.Set("CF-Connecting-IP", "9.9.9.9")
	if got := getClientKey(r); got != "198.51.100.10" {
		t.Errorf("untrusted peer header changed bucket: %q", got)
	}

	r.Header.Del("CF-Connecting-IP")
	r.RemoteAddr = "[2001:db8::1]:443"
	if got := getClientKey(r); got != "2001:db8::1" {
		t.Errorf("IPv6 host = %q, want address without port", got)
	}
}

func TestPerRouteRateLimit_KeysOnPathAndIP(t *testing.T) {
	mw := PerRouteRateLimit(1, 10*time.Second)

	srv := func(path, ip string) int {
		w := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		req := httptest.NewRequest(http.MethodGet, path, http.NoBody)
		req.RemoteAddr = ip
		rec := httptest.NewRecorder()
		w.ServeHTTP(rec, req)
		return rec.Code
	}

	if got := srv("/auth/login", "1.1.1.1:1"); got != http.StatusOK {
		t.Errorf("first /auth/login: want 200, got %d", got)
	}
	if got := srv("/auth/login", "1.1.1.1:1"); got != http.StatusTooManyRequests {
		t.Errorf("second /auth/login same IP: want 429, got %d", got)
	}
	// Different path with same IP — different bucket — should be allowed.
	if got := srv("/auth/register", "1.1.1.1:1"); got != http.StatusOK {
		t.Errorf("different path same IP: want 200, got %d", got)
	}
	// Different IP same path — different bucket.
	if got := srv("/auth/login", "2.2.2.2:1"); got != http.StatusOK {
		t.Errorf("same path different IP: want 200, got %d", got)
	}
}
