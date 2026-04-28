package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecurityHeadersMiddleware_SetsExpectedHeaders(t *testing.T) {
	handler := SecurityHeadersMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	expectedExact := map[string]string{
		"X-Content-Type-Options":    "nosniff",
		"X-Frame-Options":           "DENY",
		"X-XSS-Protection":          "1; mode=block",
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains; preload",
		"Referrer-Policy":           "strict-origin-when-cross-origin",
		"Permissions-Policy":        "geolocation=(), microphone=(), camera=()",
	}
	for k, want := range expectedExact {
		if got := rec.Header().Get(k); got != want {
			t.Errorf("header %q = %q, want %q", k, got, want)
		}
	}

	// CSP must include the frame-ancestors lockdown.
	csp := rec.Header().Get("Content-Security-Policy")
	if csp == "" {
		t.Fatal("CSP missing")
	}
	for _, must := range []string{"default-src", "frame-ancestors 'none'"} {
		if !contains(csp, must) {
			t.Errorf("CSP missing %q: %s", must, csp)
		}
	}

	// Server / X-Powered-By must not be set by the middleware itself.
	if got := rec.Header().Get("Server"); got != "" {
		t.Errorf("middleware should not introduce Server header, got %q", got)
	}
	if got := rec.Header().Get("X-Powered-By"); got != "" {
		t.Errorf("middleware should not introduce X-Powered-By header, got %q", got)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
