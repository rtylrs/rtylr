package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPaddlePOSWebhookHTTPHandlerRejectsOversizedBody(t *testing.T) {
	body := strings.NewReader(strings.Repeat("x", maxPaddleWebhookBodyBytes+1))
	req := httptest.NewRequest(http.MethodPost, "/pos/payments/webhook", body)
	rec := httptest.NewRecorder()

	PaddlePOSWebhookHTTPHandler(nil).ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusRequestEntityTooLarge)
	}
}
