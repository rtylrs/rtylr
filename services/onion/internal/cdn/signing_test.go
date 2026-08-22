package cdn

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSignedDeliveryRoundTrip(t *testing.T) {
	c := &CDN{baseURL: "https://cdn.voxire.com", apiKey: "secret"}
	u, err := c.signedURL("https://cdn.voxire.com/images/4c85c397-9fc3-4bd8-adf5-9d1ef64f5e8b?width=300", time.Now().Add(time.Hour).Unix(), c.signingSecret())
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, u.RequestURI(), nil)
	rec := httptest.NewRecorder()
	if !c.validateSignedDelivery(rec, req) {
		t.Fatalf("signed url should validate: status=%d body=%q", rec.Code, rec.Body.String())
	}
}

func TestSignedDeliveryRejectsTamperedQuery(t *testing.T) {
	c := &CDN{baseURL: "https://cdn.voxire.com", apiKey: "secret"}
	u, err := c.signedURL("https://cdn.voxire.com/images/4c85c397-9fc3-4bd8-adf5-9d1ef64f5e8b?width=300", time.Now().Add(time.Hour).Unix(), c.signingSecret())
	if err != nil {
		t.Fatal(err)
	}
	query := u.Query()
	query.Set("width", "600")
	u.RawQuery = query.Encode()

	req := httptest.NewRequest(http.MethodGet, u.RequestURI(), nil)
	rec := httptest.NewRecorder()
	if c.validateSignedDelivery(rec, req) {
		t.Fatal("tampered signed url should not validate")
	}
	if rec.Code != http.StatusForbidden || !strings.Contains(rec.Body.String(), "invalid signed url") {
		t.Fatalf("status=%d body=%q", rec.Code, rec.Body.String())
	}
}

func TestDeliveryCacheControlPrivateForSignedRequests(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/files/id?expires=1&sig=x", nil)
	if got := deliveryCacheControl(req); got != "private, no-store" {
		t.Fatalf("signed cache control = %q", got)
	}
}
