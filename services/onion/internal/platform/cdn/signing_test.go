package cdn

import (
	"fmt"
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
		t.Fatalf("expired signed cache control = %q", got)
	}
}

// A live signed link may be kept by the viewer's browser until it expires, so
// a page that shows the same attachments again does not download them again.
// It must stay private so no shared cache serves it to anyone else.
func TestDeliveryCacheControlCachesLiveSignedLinkUntilExpiry(t *testing.T) {
	expires := time.Now().Add(90 * time.Minute).Unix()
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/files/id?expires=%d&sig=x", expires), nil)
	got := deliveryCacheControl(req)
	var maxAge int64
	if _, err := fmt.Sscanf(got, "private, max-age=%d", &maxAge); err != nil {
		t.Fatalf("live signed cache control = %q, want private with a max-age", got)
	}
	if maxAge <= 0 || maxAge > 90*60 {
		t.Fatalf("max-age = %d, want at most the %d seconds left on the link", maxAge, 90*60)
	}

	unsigned := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/files/id?expires=%d", expires), nil)
	if got := deliveryCacheControl(unsigned); got != "private, no-store" {
		t.Fatalf("expires without a signature = %q, want private, no-store", got)
	}
}
