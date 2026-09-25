package license

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	platformconfig "github.com/voxire/onion/internal/platform/config"
)

func TestHostedModeWithoutKeyIsValid(t *testing.T) {
	report := VerifyConfigured(&platformconfig.AppConfig{
		DeploymentMode: ModeHosted,
		LicenseIssuer:  "voxire",
	})

	if !report.Valid {
		t.Fatalf("expected hosted mode to be valid, got %#v", report)
	}
	if report.Status != StatusHosted {
		t.Fatalf("expected hosted status, got %q", report.Status)
	}
}

func TestSelfHostedMissingLicenseIsInvalid(t *testing.T) {
	report := VerifyConfigured(&platformconfig.AppConfig{
		DeploymentMode: ModeSelfHosted,
		LicenseIssuer:  "voxire",
	})

	if report.Valid {
		t.Fatalf("expected missing self-hosted license to be invalid")
	}
	if report.Status != StatusMissing {
		t.Fatalf("expected missing status, got %q", report.Status)
	}
}

func TestSelfHostedModeDoesNotFollowRuntimeOverride(t *testing.T) {
	previousBuildMode := BuildDeploymentMode
	BuildDeploymentMode = ""
	t.Cleanup(func() { BuildDeploymentMode = previousBuildMode })
	t.Setenv("SELF_HOSTED", "true")
	if IsSelfHosted(&platformconfig.AppConfig{DeploymentMode: ModeHosted}) {
		t.Fatal("runtime SELF_HOSTED override must not change configured hosted mode")
	}
}

func TestSignedLicenseVerifies(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	token := signLicense(t, priv, Entitlement{
		LicenseID: "lic_test",
		Product:   "rtylr",
		Plan:      "team",
		Issuer:    "voxire",
		Features:  []string{"pos", "erp"},
		SeatLimit: 25,
		ExpiresAt: time.Now().UTC().Add(24 * time.Hour),
	})

	setVendorPublicKeyForTest(t, pub)
	report := VerifyKey(&platformconfig.AppConfig{
		DeploymentMode:     ModeSelfHosted,
		LicenseIssuer:      "voxire",
		LicenseGracePeriod: 14 * 24 * time.Hour,
	}, token)

	if !report.Valid {
		t.Fatalf("expected valid signed license, got %#v", report)
	}
	if !HasFeature(report, "pos") {
		t.Fatalf("expected pos feature to be enabled")
	}
}

func TestGenerateSignedLicenseFromSeedVerifies(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	seed := priv.Seed()

	token, err := GenerateSignedKey(base64.RawURLEncoding.EncodeToString(seed), Entitlement{
		LicenseID:  "lic_seed",
		OwnerEmail: "owner@example.com",
		Product:    "rtylr",
		Plan:       "self_hosted",
		Issuer:     "voxire",
		Features:   []string{"pos"},
		SelfHosted: true,
	})
	if err != nil {
		t.Fatalf("generate signed key: %v", err)
	}

	setVendorPublicKeyForTest(t, pub)
	report := VerifyKey(&platformconfig.AppConfig{
		DeploymentMode: ModeSelfHosted,
		LicenseIssuer:  "voxire",
	}, token)

	if !report.Valid {
		t.Fatalf("expected valid generated license, got %#v", report)
	}
	if report.OwnerEmail != "owner@example.com" {
		t.Fatalf("expected owner email in report, got %q", report.OwnerEmail)
	}
	if got := HashKey(token); len(got) != 64 {
		t.Fatalf("expected sha256 hash, got %q", got)
	}
}

func TestExpiredLicenseAfterGraceIsInvalid(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	token := signLicense(t, priv, Entitlement{
		LicenseID:  "lic_expired",
		Product:    "rtylr",
		Issuer:     "voxire",
		ExpiresAt:  time.Now().UTC().Add(-48 * time.Hour),
		GraceUntil: time.Now().UTC().Add(-24 * time.Hour),
	})

	setVendorPublicKeyForTest(t, pub)
	report := VerifyKey(&platformconfig.AppConfig{
		DeploymentMode: ModeSelfHosted,
		LicenseIssuer:  "voxire",
	}, token)

	if report.Valid {
		t.Fatalf("expected expired license to be invalid")
	}
	if report.Status != StatusExpired {
		t.Fatalf("expected expired status, got %q", report.Status)
	}
}

func signLicense(t *testing.T, priv ed25519.PrivateKey, entitlement Entitlement) string {
	t.Helper()

	payload, err := json.Marshal(entitlement)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	message := tokenPrefix + "." + encodedPayload
	signature := ed25519.Sign(priv, []byte(message))

	return message + "." + base64.RawURLEncoding.EncodeToString(signature)
}

func setVendorPublicKeyForTest(t *testing.T, pub ed25519.PublicKey) {
	t.Helper()
	previous := VendorPublicKey
	VendorPublicKey = base64.RawURLEncoding.EncodeToString(pub)
	t.Cleanup(func() { VendorPublicKey = previous })
}
