package probe

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthorizedStopRequestRequiresMatchingToken(t *testing.T) {
	tests := []struct {
		name      string
		header    string
		expected  string
		wantAllow bool
	}{
		{name: "missing token", expected: "secret"},
		{name: "wrong token", header: "wrong", expected: "secret"},
		{name: "matching token", header: "secret", expected: "secret", wantAllow: true},
		{name: "empty configured token", header: "secret", wantAllow: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/stop", nil)
			req.Header.Set("X-Probe-Token", tt.header)
			if got := authorizedStopRequest(req, tt.expected); got != tt.wantAllow {
				t.Fatalf("authorizedStopRequest() = %v, want %v", got, tt.wantAllow)
			}
		})
	}
}

func TestHandleStopRejectsMissingToken(t *testing.T) {
	called := false
	req := httptest.NewRequest(http.MethodPost, "/stop", nil)
	recorder := httptest.NewRecorder()
	handleStop(recorder, req, "secret", func() { called = true })

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
	if called {
		t.Fatal("stop callback must not run without a valid token")
	}
}
