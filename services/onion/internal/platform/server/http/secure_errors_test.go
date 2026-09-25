package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSecureErrorMiddleware_HappyPath(t *testing.T) {
	handler := SecureErrorMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))

	req := httptest.NewRequest(http.MethodGet, "/x", http.NoBody)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTeapot {
		t.Errorf("expected 418 to pass through, got %d", rec.Code)
	}
}

func TestSecureErrorMiddleware_RecoversFromPanic(t *testing.T) {
	handler := SecureErrorMiddleware()(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		panic("unexpected boom")
	}))

	req := httptest.NewRequest(http.MethodGet, "/x", http.NoBody)
	rec := httptest.NewRecorder()

	// Must not panic out of the middleware.
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("middleware leaked panic: %v", r)
		}
	}()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}

	var body SecureErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if body.Status != "error" {
		t.Errorf("expected status=error, got %q", body.Status)
	}
	if strings.Contains(body.Message, "boom") || strings.Contains(body.Error, "boom") {
		t.Errorf("response leaked panic value: %+v", body)
	}
}
