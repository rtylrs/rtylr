package config

import (
	"strings"
	"testing"
	"time"
)

func setLicenseConfigEnv(t *testing.T) {
	t.Helper()
	t.Setenv("RUN_MODE", "license")
	t.Setenv("JWT_SECRET", strings.Repeat("s", MinJWTSecretLength))
	t.Setenv("JWT_DURATION", "")
	t.Cleanup(func() { GlobalConfig = nil })
}

func TestLoadAndValidate_DefaultsJWTDuration(t *testing.T) {
	setLicenseConfigEnv(t)

	cfg, err := LoadAndValidate()
	if err != nil {
		t.Fatalf("LoadAndValidate: %v", err)
	}
	if cfg.JWTDuration != DefaultJWTDuration {
		t.Fatalf("JWTDuration = %v, want %v", cfg.JWTDuration, DefaultJWTDuration)
	}
}

func TestLoadAndValidate_RejectsMalformedJWTDuration(t *testing.T) {
	setLicenseConfigEnv(t)
	t.Setenv("JWT_DURATION", "not-a-duration")

	if _, err := LoadAndValidate(); err == nil {
		t.Fatal("expected malformed JWT_DURATION to fail validation")
	}
}

func TestLoadAndValidate_RejectsNonPositiveJWTDuration(t *testing.T) {
	setLicenseConfigEnv(t)
	t.Setenv("JWT_DURATION", "0s")

	if _, err := LoadAndValidate(); err == nil {
		t.Fatal("expected non-positive JWT_DURATION to fail validation")
	}
}

func TestLoadAndValidate_RejectsShortJWTSecret(t *testing.T) {
	setLicenseConfigEnv(t)
	t.Setenv("JWT_SECRET", "short")
	t.Setenv("JWT_DURATION", (24 * time.Hour).String())

	if _, err := LoadAndValidate(); err == nil {
		t.Fatal("expected short JWT_SECRET to fail validation")
	}
}
