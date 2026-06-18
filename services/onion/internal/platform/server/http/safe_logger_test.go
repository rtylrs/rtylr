package http

import (
	"net/http/httptest"
	"testing"
)

func TestSanitizeRequestForLoggingRedactsSensitiveQueryValues(t *testing.T) {
	req := httptest.NewRequest("GET", "/flow/ws?token=jwt&workspace_id=1&code=oauth&signature=sig", nil)

	got := sanitizeRequestForLogging(req)

	if got.URL.Query().Get("token") != "[REDACTED]" {
		t.Fatalf("expected token to be redacted, got %q", got.URL.Query().Get("token"))
	}
	if got.URL.Query().Get("code") != "[REDACTED]" {
		t.Fatalf("expected code to be redacted, got %q", got.URL.Query().Get("code"))
	}
	if got.URL.Query().Get("signature") != "[REDACTED]" {
		t.Fatalf("expected signature to be redacted, got %q", got.URL.Query().Get("signature"))
	}
	if got.URL.Query().Get("workspace_id") != "1" {
		t.Fatalf("expected workspace_id to remain visible, got %q", got.URL.Query().Get("workspace_id"))
	}
	if req.URL.Query().Get("token") != "jwt" {
		t.Fatalf("expected original request to remain untouched, got %q", req.URL.Query().Get("token"))
	}
}

func TestAccessLogsDisabledByDefault(t *testing.T) {
	t.Setenv("HTTP_ACCESS_LOGS", "")
	t.Setenv("ACCESS_LOGS", "")

	if accessLogsEnabled() {
		t.Fatal("expected access logs to be disabled")
	}
}

func TestAccessLogsCanBeEnabled(t *testing.T) {
	t.Setenv("HTTP_ACCESS_LOGS", "on")
	t.Setenv("ACCESS_LOGS", "")

	if !accessLogsEnabled() {
		t.Fatal("expected access logs to be enabled")
	}
}
