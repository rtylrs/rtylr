package api

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/testsupport"
	"github.com/voxire/onion/internal/testutil"
)

func webhookSecret(t *testing.T) []byte {
	t.Helper()
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		t.Fatal(err)
	}
	return secret
}

func signGitHubBody(secret []byte, body string) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(body))
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func gitHubDelivery(body, signature, delivery, event string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/flow/github/webhook", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if signature != "" {
		req.Header.Set("X-Hub-Signature-256", signature)
	}
	if delivery != "" {
		req.Header.Set("X-GitHub-Delivery", delivery)
	}
	if event != "" {
		req.Header.Set("X-GitHub-Event", event)
	}
	return req
}

// A nil service panics if reached, so these statuses prove the request was
// refused before any processing.
func TestGitHubWebhookRefusesUnsignedAndMissignedDeliveries(t *testing.T) {
	secret := webhookSecret(t)
	body := `{"action":"deleted","installation":{"id":1}}`
	cases := map[string]struct {
		handler http.HandlerFunc
		req     *http.Request
		want    int
	}{
		"no signature":     {GitHubWebhookHTTPHandler(nil, secret), gitHubDelivery(body, "", "d-1", "installation"), http.StatusUnauthorized},
		"other secret":     {GitHubWebhookHTTPHandler(nil, secret), gitHubDelivery(body, signGitHubBody([]byte("another-secret"), body), "d-1", "installation"), http.StatusUnauthorized},
		"body changed":     {GitHubWebhookHTTPHandler(nil, secret), gitHubDelivery(body+" ", signGitHubBody(secret, body), "d-1", "installation"), http.StatusUnauthorized},
		"sha1 signature":   {GitHubWebhookHTTPHandler(nil, secret), gitHubDelivery(body, "sha1="+strings.Repeat("0", 40), "d-1", "installation"), http.StatusUnauthorized},
		"not hex":          {GitHubWebhookHTTPHandler(nil, secret), gitHubDelivery(body, "sha256=zz", "d-1", "installation"), http.StatusUnauthorized},
		"secret unset":     {GitHubWebhookHTTPHandler(nil, nil), gitHubDelivery(body, signGitHubBody(nil, body), "d-1", "installation"), http.StatusServiceUnavailable},
		"no delivery id":   {GitHubWebhookHTTPHandler(nil, secret), gitHubDelivery(body, signGitHubBody(secret, body), "", "installation"), http.StatusBadRequest},
		"no event":         {GitHubWebhookHTTPHandler(nil, secret), gitHubDelivery(body, signGitHubBody(secret, body), "d-1", ""), http.StatusBadRequest},
		"long delivery id": {GitHubWebhookHTTPHandler(nil, secret), gitHubDelivery(body, signGitHubBody(secret, body), strings.Repeat("d", 65), "installation"), http.StatusBadRequest},
	}
	for name, tc := range cases {
		res := httptest.NewRecorder()
		tc.handler(res, tc.req)
		if res.Code != tc.want {
			t.Errorf("%s: status = %d, want %d", name, res.Code, tc.want)
		}
	}

	big := `{"pad":"` + strings.Repeat("x", maxGitHubWebhookBodyBytes) + `"}`
	res := httptest.NewRecorder()
	GitHubWebhookHTTPHandler(nil, secret)(res, gitHubDelivery(big, signGitHubBody(secret, big), "d-2", "push"))
	if res.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("oversized body: status = %d, want %d", res.Code, http.StatusRequestEntityTooLarge)
	}
}

func TestGitHubWebhookAcceptsSignedDeliveriesOnce(t *testing.T) {
	db := testutil.OpenDB(t)
	cleanup := func() {
		if _, err := db.Exec("DELETE FROM flow_github_webhook_deliveries WHERE delivery_id LIKE 'test-api-%'"); err != nil {
			t.Fatalf("cleanup: %v", err)
		}
	}
	cleanup()
	t.Cleanup(cleanup)
	secret := webhookSecret(t)
	handler := GitHubWebhookHTTPHandler(application.NewService(context.Background(), testsupport.NewRepository(db)), secret)
	send := func(delivery, event, body string) int {
		res := httptest.NewRecorder()
		handler(res, gitHubDelivery(body, signGitHubBody(secret, body), delivery, event))
		return res.Code
	}

	revoke := `{"action":"revoked","sender":{"id":770099,"login":"nobody"}}`
	if got := send("test-api-revoke", "github_app_authorization", revoke); got != http.StatusOK {
		t.Fatalf("signed delivery: status = %d, want 200", got)
	}
	if got := send("test-api-revoke", "github_app_authorization", revoke); got != http.StatusOK {
		t.Fatalf("repeated delivery: status = %d, want 200", got)
	}
	var recorded int
	if err := db.QueryRow("SELECT COUNT(*) FROM flow_github_webhook_deliveries WHERE delivery_id = 'test-api-revoke'").Scan(&recorded); err != nil || recorded != 1 {
		t.Fatalf("recorded deliveries = %d, %v; want 1", recorded, err)
	}

	unknown := `{"action":"deleted","installation":{"id":424242}}`
	if got := send("test-api-unknown", "installation", unknown); got != http.StatusAccepted {
		t.Fatalf("unknown installation: status = %d, want 202", got)
	}
	if got := send("test-api-malformed", "installation", `{"action":`); got != http.StatusBadRequest {
		t.Fatalf("malformed payload: status = %d, want 400", got)
	}
}

// GitHub's published test vector for X-Hub-Signature-256.
func TestGitHubWebhookSignatureMatchesGitHubsExample(t *testing.T) {
	secret := []byte("It's a Secret to Everybody")
	body := []byte("Hello, World!")
	const header = "sha256=757107ea0eb2509fc211221cce984b8a37570b6d7586c22c46f4379c8b043e17"
	if !verifyGitHubWebhookSignature(secret, body, header) {
		t.Fatal("GitHub's example signature did not verify")
	}
	if verifyGitHubWebhookSignature(secret, []byte("Hello, World?"), header) {
		t.Fatal("a changed body verified")
	}
}
