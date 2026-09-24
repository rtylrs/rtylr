package license

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync/atomic"
	"testing"

	platformconfig "github.com/voxire/onion/internal/platform/config"
)

// selfHostConfig is the minimum a self-hosted deployment needs for the
// validator to actually reach out: a mode, a key, and a clock state file it is
// allowed to write.
func selfHostConfig(t *testing.T, authorityURL string) *platformconfig.AppConfig {
	t.Helper()
	return &platformconfig.AppConfig{
		DeploymentMode:        "self_hosted",
		LicenseKey:            "test-license-key",
		LicenseAuthorityURL:   authorityURL,
		LicenseClockStateFile: filepath.Join(t.TempDir(), "clock.json"),
	}
}

// #1079 restored reading LICENSE_AUTHORITY_URL into the config, but
// NewCachedValidator ignored it and always built its client against the hosted
// default. Since that constructor is the only place a client is made, the
// setting reached nothing and an air-gapped install could not point at its own
// authority (#1178).
//
// Asserted through the constructor and a real request rather than through
// config loading, which is what the previous test covered and what let this
// stay hidden.
func TestCachedValidatorCallsTheConfiguredAuthority(t *testing.T) {
	var calls atomic.Int32
	var gotPath string

	authority := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"valid":true,"status":"valid"}`))
	}))
	defer authority.Close()

	report := NewCachedValidator(selfHostConfig(t, authority.URL)).Refresh(context.Background())

	if got := calls.Load(); got != 1 {
		t.Fatalf("the configured authority was called %d times, want 1: the validator is still using the hosted default", got)
	}
	if gotPath != "/license/verify" {
		t.Errorf("authority path = %q, want /license/verify", gotPath)
	}
	if !report.Valid {
		t.Errorf("report = %+v, want the configured authority's answer to be used", report)
	}
}

// A deployment that configures nothing must keep pointing at the hosted
// authority. Config already defaults the field, and NewClient falls back for an
// empty string, so both routes to "unset" are covered here.
func TestCachedValidatorFallsBackToTheHostedAuthority(t *testing.T) {
	for _, configured := range []string{"", "   "} {
		validator := NewCachedValidator(selfHostConfig(t, configured))
		if got := validator.client.authorityURL; got != platformconfig.DefaultLicenseAuthorityURL {
			t.Errorf("authority for %q = %q, want the hosted default %q",
				configured, got, platformconfig.DefaultLicenseAuthorityURL)
		}
	}
}

// A nil config must not panic the constructor; it is how several helpers in
// this package are already called.
func TestCachedValidatorHandlesANilConfig(t *testing.T) {
	validator := NewCachedValidator(nil)
	if validator == nil {
		t.Fatal("NewCachedValidator(nil) returned nothing")
	}
	if got := validator.client.authorityURL; got != platformconfig.DefaultLicenseAuthorityURL {
		t.Errorf("authority = %q, want the hosted default", got)
	}
}

// A hosted deployment must never contact an authority, whatever the field
// says. This is the reason honouring the configured value is safe: the URL is
// only ever used by an operator running their own install.
func TestHostedDeploymentNeverCallsAnAuthority(t *testing.T) {
	var calls atomic.Int32
	authority := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer authority.Close()

	cfg := selfHostConfig(t, authority.URL)
	cfg.DeploymentMode = "hosted"

	validator := NewCachedValidator(cfg)
	validator.Refresh(context.Background())
	validator.Current(context.Background())

	if got := calls.Load(); got != 0 {
		t.Errorf("a hosted deployment contacted the authority %d times, want 0", got)
	}
}

func TestCachedValidatorPreservesInvalidAuthorityReport(t *testing.T) {
	var calls atomic.Int32
	authority := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"valid":false,"status":"expired","reason":"license expired"}`))
	}))
	defer authority.Close()

	validator := NewCachedValidator(selfHostConfig(t, authority.URL))
	first := validator.Refresh(context.Background())
	if first.Status != StatusExpired || first.Reason != "license expired" {
		t.Fatalf("first report = %+v, want the authority's expired report", first)
	}

	cached := validator.Current(context.Background())
	if cached.Status != StatusExpired || cached.Reason != "license expired" {
		t.Fatalf("cached report = %+v, want the authority's expired report", cached)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("authority was called %d times, want one refresh", got)
	}
}
