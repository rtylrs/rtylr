package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPaddleWebhookHTTPHandlerRejectsOversizedBody(t *testing.T) {
	body := strings.NewReader(strings.Repeat("x", maxPaddleWebhookBodyBytes+1))
	req := httptest.NewRequest(http.MethodPost, "/subscriptions/paddle-webhook", body)
	rec := httptest.NewRecorder()

	PaddleWebhookHTTPHandler(nil).ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusRequestEntityTooLarge)
	}
}
