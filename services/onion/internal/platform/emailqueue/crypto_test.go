package emailqueue

import (
	"strings"
	"testing"
)

func TestSealOpenRoundTrip(t *testing.T) {
	t.Setenv("EMAIL_ENCRYPTION_KEY", "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")

	body := `<a href="https://auth.example.com/password-reset/confirm?token=abc123">Reset</a>`
	sealed := sealHTML(body)
	if sealed == body {
		t.Fatal("body was stored unencrypted despite configured key")
	}
	if !strings.HasPrefix(sealed, encryptedPrefix) {
		t.Fatalf("sealed body missing %q prefix", encryptedPrefix)
	}
	if strings.Contains(sealed, "abc123") {
		t.Fatal("ciphertext leaks the token")
	}
	if got := openHTML(sealed); got != body {
		t.Fatalf("openHTML round-trip mismatch: %q", got)
	}
}

func TestOpenHTML_LegacyPlaintextPassthrough(t *testing.T) {
	t.Setenv("EMAIL_ENCRYPTION_KEY", "")
	body := "<p>legacy row written before encryption</p>"
	if got := openHTML(body); got != body {
		t.Fatalf("legacy plaintext must pass through unchanged, got %q", got)
	}
}

func TestOpenHTML_UndecryptableReturnsEmpty(t *testing.T) {
	t.Setenv("EMAIL_ENCRYPTION_KEY", "fedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210")
	tampered := encryptedPrefix + "bm90LWFuLWFjdHVhbC1jcnlwdG9ncmFt"
	if got := openHTML(tampered); got != "" {
		t.Fatalf("undecryptable row must return empty, got %q", got)
	}
}

func TestSealHTML_NoKeyStoresPlaintext(t *testing.T) {
	t.Setenv("EMAIL_ENCRYPTION_KEY", "")
	body := "<p>no key configured</p>"
	if got := sealHTML(body); got != body {
		t.Fatalf("without a key the body must be stored as-is, got %q", got)
	}
}

func TestHexDecode(t *testing.T) {
	got, err := hexDecode("00ff10")
	if err != nil || len(got) != 3 || got[1] != 0xff {
		t.Fatalf("hexDecode = %v, %v", got, err)
	}
	if _, err := hexDecode("zz"); err == nil {
		t.Fatal("expected error for non-hex input")
	}
}
