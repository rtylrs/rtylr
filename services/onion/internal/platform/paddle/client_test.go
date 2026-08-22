package paddle

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/platform/config"
)

func TestPaddleHTTPClientHasTimeout(t *testing.T) {
	if paddleHTTPClient.Timeout != paddleRequestTimeout {
		t.Fatalf("expected Paddle HTTP timeout %s, got %s", paddleRequestTimeout, paddleHTTPClient.Timeout)
	}
}

func TestReadPaddleResponseBody(t *testing.T) {
	body, err := readPaddleResponseBody(strings.NewReader(`{"ok":true}`))
	if err != nil {
		t.Fatalf("readPaddleResponseBody: %v", err)
	}
	if string(body) != `{"ok":true}` {
		t.Fatalf("unexpected body: %s", string(body))
	}
}

func TestReadPaddleResponseBodyRejectsOversizedBody(t *testing.T) {
	_, err := readPaddleResponseBody(strings.NewReader(strings.Repeat("x", maxPaddleResponseBodyBytes+1)))
	if err == nil {
		t.Fatal("expected oversized body to fail")
	}
}

func signedWebhookHeader(secret string, body []byte, ts time.Time) string {
	timestamp := strconv.FormatInt(ts.Unix(), 10)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp + ":"))
	mac.Write(body)
	return "ts=" + timestamp + ";h1=" + hex.EncodeToString(mac.Sum(nil))
}

func TestVerifyWebhookSignatureRequiresSecretInProduction(t *testing.T) {
	old := config.GlobalConfig
	config.GlobalConfig = nil
	t.Cleanup(func() { config.GlobalConfig = old })
	t.Setenv("APP_ENV", "production")
	t.Setenv("PADDLE_WEBHOOK_SECRET", "")

	if err := VerifyWebhookSignature([]byte(`{}`), ""); err == nil {
		t.Fatal("expected missing secret to fail in production")
	}
}

func TestVerifyWebhookSignatureAllowsExplicitInsecureOptOutOutsideProduction(t *testing.T) {
	old := config.GlobalConfig
	config.GlobalConfig = nil
	t.Cleanup(func() { config.GlobalConfig = old })
	t.Setenv("APP_ENV", "development")
	t.Setenv("PADDLE_WEBHOOK_SECRET", "")
	t.Setenv("PADDLE_ALLOW_INSECURE_WEBHOOKS", "1")

	if err := VerifyWebhookSignature([]byte(`{}`), ""); err != nil {
		t.Fatalf("expected explicit local opt-out to be allowed, got %v", err)
	}
}

func TestVerifyWebhookSignatureRefusesInsecureOptOutInProduction(t *testing.T) {
	old := config.GlobalConfig
	config.GlobalConfig = nil
	t.Cleanup(func() { config.GlobalConfig = old })
	t.Setenv("APP_ENV", "production")
	t.Setenv("PADDLE_WEBHOOK_SECRET", "")
	t.Setenv("PADDLE_ALLOW_INSECURE_WEBHOOKS", "1")

	if err := VerifyWebhookSignature([]byte(`{}`), ""); err == nil {
		t.Fatal("expected production to refuse insecure webhook opt-out")
	}
}

func TestVerifyWebhookSignatureValidatesHMAC(t *testing.T) {
	old := config.GlobalConfig
	config.GlobalConfig = nil
	t.Cleanup(func() { config.GlobalConfig = old })
	t.Setenv("APP_ENV", "production")
	t.Setenv("PADDLE_WEBHOOK_SECRET", "whsec_test")

	body := []byte(`{"event_type":"subscription.updated"}`)
	header := signedWebhookHeader("whsec_test", body, time.Now())

	if err := VerifyWebhookSignature(body, header); err != nil {
		t.Fatalf("expected valid signature, got %v", err)
	}
}

func TestVerifyWebhookSignatureRejectsStaleTimestamp(t *testing.T) {
	old := config.GlobalConfig
	config.GlobalConfig = nil
	t.Cleanup(func() { config.GlobalConfig = old })
	t.Setenv("APP_ENV", "production")
	t.Setenv("PADDLE_WEBHOOK_SECRET", "whsec_test")

	body := []byte(`{"event_type":"subscription.updated"}`)
	header := signedWebhookHeader("whsec_test", body, time.Now().Add(-10*time.Minute))

	if err := VerifyWebhookSignature(body, header); err == nil {
		t.Fatal("expected stale signature timestamp to fail")
	}
}
