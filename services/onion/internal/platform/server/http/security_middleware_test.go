package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsMaliciousPath(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		// Sub-segments of legitimate API paths must NOT match.
		{"/hr/setup/status", false},
		{"/erp/admin-panel", false},
		{"/console", false},
		{"/console/stats", false},
		{"/console/resources", false},
		{"/upload", false},
		{"/upload/plan", false},
		{"/files/abc-123.jpg", false},
		{"/images/abc-123.png", false},
		{"/uploads/file.png", false},

		// First-segment matches against the pattern list ARE blocked.
		{"/admin/settings", true},
		{"/etc/passwd", true},
		{"/proc/self", true},
		{"/wp-admin", true}, // /wp-admin IS in the pattern list

		// Patterns declared with a trailing slash must still match; before the
		// normalisation these entries were dead because a path segment can
		// never contain '/'.
		{"/.git/config", true},
		{"/.git", true},
		{"/debug/pprof", true},
		{"/log/app.log", true},
		{"/lock", true},
		{"/api/v1/probe", true},

		// But the allow-listed routes they would otherwise shadow stay reachable.
		{"/.well-known/rtylr.json", false},

		// A trailing-slash pattern must not become a substring match.
		{"/logger", false},

		// Empty path / root — never malicious.
		{"/", false},
		{"", false},

		// Case insensitivity.
		{"/Admin/Settings", true},
		{"/ETC/passwd", true},
	}

	for _, tc := range cases {
		if got := isMaliciousPath(tc.path); got != tc.want {
			t.Errorf("isMaliciousPath(%q) = %v, want %v", tc.path, got, tc.want)
		}
	}
}

func TestHasSuspiciousHeaders(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	if hasSuspiciousHeaders(r) {
		t.Error("clean request should have no suspicious headers")
	}

	r.Header.Set("X-Forwarded-Method", "TRACE")
	if !hasSuspiciousHeaders(r) {
		t.Error("X-Forwarded-Method should be flagged")
	}
}

func TestIsLegitimateTraffic_AcceptsBrowsersOnPublicPaths(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
	r.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)")
	if !isLegitimateTraffic(r) {
		t.Error("Mozilla on /health should look legitimate")
	}
}

func TestIsLegitimateTraffic_RejectsBrowserOnPrivateAPI(t *testing.T) {
	// Regression: previously /erp/product/v1/products matched because every path contains
	// "/", which short-circuited the legitimate-paths check to true. Only the
	// explicit allowlist should match now.
	r := httptest.NewRequest(http.MethodGet, "/erp/product/v1/products", http.NoBody)
	r.Header.Set("User-Agent", "Mozilla/5.0")
	if isLegitimateTraffic(r) {
		t.Error("/erp/product/v1/products is a private API, browser UA should not make it legitimate")
	}
}

func TestIsLegitimateTraffic_RejectsScannerUA(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
	r.Header.Set("User-Agent", "sqlmap/1.0")
	if isLegitimateTraffic(r) {
		t.Error("sqlmap should not look legitimate")
	}
}

func TestIsLegitimateTraffic_NeedsBothAgentAndPath(t *testing.T) {
	// A scanner UA on a public path is not legitimate.
	r := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
	r.Header.Set("User-Agent", "nikto/2.1")
	if isLegitimateTraffic(r) {
		t.Error("scanner UA should not be legitimate even on /health")
	}

	// A real browser with no UA at all is not legitimate.
	r = httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
	r.Header.Del("User-Agent")
	if isLegitimateTraffic(r) {
		t.Error("missing UA should not be legitimate")
	}
}

func TestHasSuspiciousParams(t *testing.T) {
	cases := []struct {
		query string
		want  bool
	}{
		{"q=hello", false},
		{"name=alice", false},
		{"cmd=whoami", true},
		{"sql=DROP+TABLE+users", true},
		{"script=alert(1)", true},
		{"CMD=ls", true}, // case insensitive
	}
	for _, tc := range cases {
		r := httptest.NewRequest(http.MethodGet, "/?"+tc.query, http.NoBody)
		if got := hasSuspiciousParams(r); got != tc.want {
			t.Errorf("hasSuspiciousParams(%q) = %v, want %v", tc.query, got, tc.want)
		}
	}
}

func TestSecurityMiddlewareAllowsCDNTransformParams(t *testing.T) {
	cases := []string{
		"/images/4c85c397-1234-9d1ef64f5e8b?sharpen=50",
		"/images/4c85c397-1234-9d1ef64f5e8b?threshold=128",
		"/images/4c85c397-1234-9d1ef64f5e8b?aspect-ratio=4:3",
		"/images/4c85c397-1234-9d1ef64f5e8b?sharpen=50&aspect-ratio=4:3&width=800",
	}
	for _, target := range cases {
		nextCalled := false
		handler := SecurityMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusOK)
		}))

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, http.NoBody))

		if rec.Code != http.StatusOK || !nextCalled {
			t.Errorf("GET %s: status = %d, nextCalled = %t; CDN transform request must reach the handler", target, rec.Code, nextCalled)
		}
	}
}

func TestSecurityMiddlewareStillBlocksSuspiciousParamsOutsideAllowList(t *testing.T) {
	nextCalled := false
	handler := SecurityMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/hr/equipment/v1/assets?cmd=whoami", http.NoBody))

	if rec.Code != http.StatusBadRequest || nextCalled {
		t.Errorf("status = %d, nextCalled = %t; suspicious params must still be rejected outside allow-listed routes", rec.Code, nextCalled)
	}
}

func TestSanitizeHeaders_StripsAuth(t *testing.T) {
	in := map[string][]string{
		"Authorization": {"Bearer secret-token"},
		"Cookie":        {"session=abc"},
		"X-Internal":    {"value"},
	}
	out := sanitizeHeaders(in)
	for _, banned := range []string{"Authorization", "Cookie"} {
		if vals, ok := out[banned]; ok {
			for _, v := range vals {
				if v != "[REDACTED]" && v != "" {
					t.Errorf("expected %s to be redacted, got %q", banned, v)
				}
			}
		}
	}
	if _, ok := out["X-Internal"]; !ok {
		t.Error("non-sensitive headers should pass through")
	}
}

func TestSanitizeQueryParams_StripsSensitiveValues(t *testing.T) {
	in := map[string][]string{
		"token":    {"sk-abc123"},
		"password": {"hunter2"},
		"q":        {"search-term"},
	}
	out := sanitizeQueryParams(in)
	if vals := out["token"]; len(vals) > 0 && vals[0] == "sk-abc123" {
		t.Error("token value should be redacted in sanitized output")
	}
	if vals := out["password"]; len(vals) > 0 && vals[0] == "hunter2" {
		t.Error("password value should be redacted in sanitized output")
	}
	if vals := out["q"]; len(vals) == 0 || vals[0] != "search-term" {
		t.Error("non-sensitive query params should pass through unchanged")
	}
}
