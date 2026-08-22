package cdn

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPIKeyAuth(t *testing.T) {
	// next writes 418 so a test can tell "the request passed the middleware"
	// apart from a 401 produced by the middleware itself.
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})

	t.Run("dev-open when no shared key is configured", func(t *testing.T) {
		t.Setenv("APP_ENV", "development")
		c := &CDN{apiKey: ""}
		rec := httptest.NewRecorder()
		c.apiKeyAuth(next).ServeHTTP(rec, httptest.NewRequest("POST", "/upload", nil))
		if rec.Code != http.StatusTeapot {
			t.Errorf("expected pass-through, got %d", rec.Code)
		}
	})

	t.Run("accepts the shared key", func(t *testing.T) {
		c := &CDN{apiKey: "s3cret"}
		req := httptest.NewRequest("POST", "/upload", nil)
		req.Header.Set("Authorization", "Bearer s3cret")
		rec := httptest.NewRecorder()
		c.apiKeyAuth(next).ServeHTTP(rec, req)
		if rec.Code != http.StatusTeapot {
			t.Errorf("expected accept, got %d", rec.Code)
		}
	})

	t.Run("rejects a missing token", func(t *testing.T) {
		c := &CDN{apiKey: "s3cret"}
		rec := httptest.NewRecorder()
		c.apiKeyAuth(next).ServeHTTP(rec, httptest.NewRequest("POST", "/upload", nil))
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("fails closed without a key outside development", func(t *testing.T) {
		t.Setenv("APP_ENV", "production")
		c := &CDN{apiKey: ""}
		rec := httptest.NewRecorder()
		c.apiKeyAuth(next).ServeHTTP(rec, httptest.NewRequest("POST", "/upload", nil))
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rec.Code)
		}
	})
}

func TestValidateSecurityConfig(t *testing.T) {
	if err := ValidateSecurityConfig("production", ""); err == nil {
		t.Fatal("expected production without CDN_API_KEY to fail")
	}
	if err := ValidateSecurityConfig("development", ""); err != nil {
		t.Fatalf("development without CDN_API_KEY should be allowed: %v", err)
	}
}
