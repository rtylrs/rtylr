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

	req := httptest.NewRequest(http.MethodGet, "/x", http.NoBody)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	expectedExact := map[string]string{
		"X-Content-Type-Options":            "nosniff",
		"X-Frame-Options":                   "DENY",
		"X-XSS-Protection":                  "0",
		"Strict-Transport-Security":         "max-age=31536000; includeSubDomains; preload",
		"Referrer-Policy":                   "strict-origin-when-cross-origin",
		"X-Permitted-Cross-Domain-Policies": "none",
		"Cross-Origin-Opener-Policy":        "same-origin-allow-popups",
		"Cross-Origin-Resource-Policy":      "cross-origin",
	}
	for k, want := range expectedExact {
		if got := rec.Header().Get(k); got != want {
			t.Errorf("header %q = %q, want %q", k, got, want)
		}
	}

	// Permissions-Policy must lock down the sensitive features.
	pp := rec.Header().Get("Permissions-Policy")
	for _, feature := range []string{"camera=()", "microphone=()", "geolocation=()", "payment=()"} {
		if !contains(pp, feature) {
			t.Errorf("Permissions-Policy missing %q: %s", feature, pp)
		}
	}

	// CSP must include the frame-ancestors and object-src lockdowns.
	csp := rec.Header().Get("Content-Security-Policy")
	if csp == "" {
		t.Fatal("CSP missing")
	}
	for _, must := range []string{"default-src", "frame-ancestors 'none'", "object-src 'none'", "base-uri 'self'"} {
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
