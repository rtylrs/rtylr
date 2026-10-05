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
	scope, ok := c.validateSignedDelivery(rec, req)
	if !ok {
		t.Fatalf("signed url should validate: status=%d body=%q", rec.Code, rec.Body.String())
	}
	// A valid signature is the credential for the file it names, which is what
	// makes private (non-public) assets deliverable.
	if !scope.signed {
		t.Fatal("a validated signature should grant the signed read scope")
	}
}

// An unsigned request keeps the caller's existing scope rather than inheriting
// the anonymous one, so signing stays optional and an authenticated tenant can
// still read its own non-public assets.
func TestUnsignedDeliveryKeepsCallerScope(t *testing.T) {
	c := &CDN{baseURL: "https://cdn.voxire.com", apiKey: "secret"}
	req := httptest.NewRequest(http.MethodGet, "/files/id", nil)
	req = req.WithContext(withDeliveryScope(req.Context(), tenantRead(4242)))
	rec := httptest.NewRecorder()

	scope, ok := c.validateSignedDelivery(rec, req)
	if !ok {
		t.Fatalf("unsigned delivery should pass: status=%d body=%q", rec.Code, rec.Body.String())
	}
	if scope.signed || scope.tenant != 4242 {
		t.Fatalf("scope = %+v, want the caller's tenant scope", scope)
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
	if _, ok := c.validateSignedDelivery(rec, req); ok {
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
