package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthRouteRateLimit_LoginBudget(t *testing.T) {
	handler := AuthRouteRateLimit(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	const path = "/common/auth/v1/login"
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodPost, path, nil)
		req.RemoteAddr = "203.0.113.10:1234"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: status = %d, want 200", i+1, rec.Code)
		}
	}

	req := httptest.NewRequest(http.MethodPost, path, nil)
	req.RemoteAddr = "203.0.113.10:1234"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("6th login from same IP: status = %d, want 429", rec.Code)
	}
}

func TestAuthRouteRateLimit_OtherPathsUnaffected(t *testing.T) {
	handler := AuthRouteRateLimit(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 50; i++ {
		req := httptest.NewRequest(http.MethodGet, "/common/auth/v1/email-exists", nil)
		req.RemoteAddr = "203.0.113.11:1234"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d to unlisted path: status = %d, want 200", i+1, rec.Code)
		}
	}
}

func TestAuthRouteRateLimit_RegisterBudget(t *testing.T) {
	handler := AuthRouteRateLimit(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	const path = "/common/auth/v1/register"
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodPost, path, nil)
		req.RemoteAddr = "203.0.113.12:1234"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: status = %d, want 200", i+1, rec.Code)
		}
	}

	req := httptest.NewRequest(http.MethodPost, path, nil)
	req.RemoteAddr = "203.0.113.12:1234"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("4th register from same IP: status = %d, want 429", rec.Code)
	}
}
