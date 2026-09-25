package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCORSMiddlewareAllowsIdempotencyKey(t *testing.T) {
	request := httptest.NewRequest(http.MethodOptions, "/", nil)
	request.Header.Set("Origin", "https://app.voxire.com")
	recorder := httptest.NewRecorder()

	corsMiddleware()(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("OPTIONS request should not reach the wrapped handler")
	})).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if !strings.Contains(recorder.Header().Get("Access-Control-Allow-Headers"), "Idempotency-Key") {
		t.Fatalf("Access-Control-Allow-Headers = %q, want Idempotency-Key", recorder.Header().Get("Access-Control-Allow-Headers"))
	}
}
