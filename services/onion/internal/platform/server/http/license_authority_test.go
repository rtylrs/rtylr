package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	platformconfig "github.com/voxire/onion/internal/platform/config"
)

func TestLicenseAuthorityRoutes(t *testing.T) {
	previousConfig := platformconfig.GlobalConfig
	t.Cleanup(func() { platformconfig.GlobalConfig = previousConfig })
	r := chi.NewRouter()
	cfg := &platformconfig.AppConfig{
		RunMode:        "license",
		DeploymentMode: "hosted",
		InternalToken:  "authority-secret",
	}
	platformconfig.GlobalConfig = cfg

	RegisterLicenseAuthorityRoutes(r, cfg, nil)

	// Test GET /license/status
	req := httptest.NewRequest(http.MethodGet, "/license/status", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 from /license/status, got %d: %s", rec.Code, rec.Body.String())
	}

	// Test POST /license/verify with empty body
	req = httptest.NewRequest(http.MethodPost, "/license/verify", strings.NewReader(`{}`))
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 from /license/verify, got %d: %s", rec.Code, rec.Body.String())
	}

	// Test POST /license/activate without instance_id. Activation requires the
	// internal token, so this asserts the body validation past that gate.
	req = httptest.NewRequest(http.MethodPost, "/license/activate", strings.NewReader(`{"license_key":"abc"}`))
	req.Header.Set("Internal-Token", "authority-secret")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 from /license/activate without instance_id, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestLicenseActivationDoesNotUseConfiguredKeyAsAnonymousFallback(t *testing.T) {
	previousConfig := platformconfig.GlobalConfig
	platformconfig.GlobalConfig = nil
	t.Cleanup(func() { platformconfig.GlobalConfig = previousConfig })
	t.Setenv("INTERNAL_TOKEN", "authority-secret")
	r := chi.NewRouter()
	cfg := &platformconfig.AppConfig{
		RunMode:        "license",
		DeploymentMode: "self_hosted",
		LicenseKey:     "server-side-license-key",
	}
	RegisterLicenseAuthorityRoutes(r, cfg, nil)

	// Without the internal token the request is denied before any license
	// evaluation, so the server-side key can never stand in for a caller.
	req := httptest.NewRequest(http.MethodPost, "/license/activate", strings.NewReader(`{"instance_id":"attacker-1"}`))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous activation = %d, want 401: %s", rec.Code, rec.Body.String())
	}

	// An authenticated internal caller that omits license_key is still
	// rejected rather than activated against the server's own key.
	req = httptest.NewRequest(http.MethodPost, "/license/activate", strings.NewReader(`{"instance_id":"attacker-1"}`))
	req.Header.Set("Internal-Token", "authority-secret")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code == http.StatusOK {
		t.Fatalf("activation without an explicit key succeeded using the server-side key: %s", rec.Body.String())
	}
}

func TestLicenseAuthorityRoutesAreNotMountedOnPublicAPI(t *testing.T) {
	r := chi.NewRouter()
	RegisterLicenseAuthorityRoutes(r, &platformconfig.AppConfig{
		RunMode:        "api",
		DeploymentMode: "self_hosted",
		LicenseKey:     "server-side-license-key",
	}, nil)

	req := httptest.NewRequest(http.MethodPost, "/license/activate", strings.NewReader(`{"instance_id":"attacker-1","license_key":"provided"}`))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("public API authority route status = %d, want 404", rec.Code)
	}
}

func TestLicenseActivationRequiresInternalTokenWhenConfigured(t *testing.T) {
	r := chi.NewRouter()
	previousConfig := platformconfig.GlobalConfig
	t.Cleanup(func() { platformconfig.GlobalConfig = previousConfig })
	cfg := &platformconfig.AppConfig{
		RunMode:        "license",
		DeploymentMode: "self_hosted",
		InternalToken:  "authority-secret",
	}
	platformconfig.GlobalConfig = cfg
	RegisterLicenseAuthorityRoutes(r, cfg, nil)

	req := httptest.NewRequest(http.MethodPost, "/license/activate", strings.NewReader(`{"instance_id":"legitimate","license_key":"provided"}`))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("activation without internal token = %d, want 401", rec.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/license/activate", strings.NewReader(`{"instance_id":"legitimate","license_key":"provided"}`))
	req.Header.Set("Internal-Token", "authority-secret")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code == http.StatusUnauthorized {
		t.Fatal("valid internal token was rejected")
	}
}

// Regression: activation used to be gated on "if a token is configured, then
// check it", so an authority started without INTERNAL_TOKEN accepted anonymous
// activations. Activation must require a verified internal token whether or not
// one is configured.
func TestLicenseActivationFailsClosedWithoutConfiguredInternalToken(t *testing.T) {
	previousConfig := platformconfig.GlobalConfig
	platformconfig.GlobalConfig = nil
	t.Cleanup(func() { platformconfig.GlobalConfig = previousConfig })
	t.Setenv("INTERNAL_TOKEN", "")

	tests := []struct {
		name        string
		cfgToken    string
		header      string
		wantAllowed bool
	}{
		{name: "no token configured, no header"},
		{name: "no token configured, guessed header", header: "internal"},
		{name: "no token configured, empty header", header: ""},
		{name: "token configured, no header", cfgToken: "authority-secret"},
		{name: "token configured, wrong header", cfgToken: "authority-secret", header: "guess"},
		{name: "token configured, empty header", cfgToken: "authority-secret", header: ""},
		{name: "token configured, correct header", cfgToken: "authority-secret", header: "authority-secret", wantAllowed: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			cfg := &platformconfig.AppConfig{
				RunMode:        "license",
				DeploymentMode: "self_hosted",
				InternalToken:  tt.cfgToken,
				LicenseKey:     "server-side-license-key",
			}
			platformconfig.GlobalConfig = cfg
			RegisterLicenseAuthorityRoutes(r, cfg, nil)

			req := httptest.NewRequest(http.MethodPost, "/license/activate", strings.NewReader(`{"instance_id":"attacker-1"}`))
			if tt.header != "" {
				req.Header.Set("Internal-Token", tt.header)
			}
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			denied := rec.Code == http.StatusUnauthorized
			if tt.wantAllowed && denied {
				t.Fatalf("valid internal token was rejected with %d", rec.Code)
			}
			if !tt.wantAllowed && !denied {
				t.Fatalf("activation was not denied: got %d, want 401. body: %s", rec.Code, rec.Body.String())
			}
			if denied && !strings.Contains(rec.Body.String(), "internal authority token required") {
				t.Errorf("unexpected denial body: %s", rec.Body.String())
			}
		})
	}
}
