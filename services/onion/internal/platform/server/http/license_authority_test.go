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
	r := chi.NewRouter()
	cfg := &platformconfig.AppConfig{
		RunMode:        "license",
		DeploymentMode: "hosted",
	}

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

	// Test POST /license/activate without instance_id
	req = httptest.NewRequest(http.MethodPost, "/license/activate", strings.NewReader(`{"license_key":"abc"}`))
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 from /license/activate without instance_id, got %d", rec.Code)
	}
}

func TestLicenseActivationDoesNotUseConfiguredKeyAsAnonymousFallback(t *testing.T) {
	previousConfig := platformconfig.GlobalConfig
	platformconfig.GlobalConfig = nil
	t.Cleanup(func() { platformconfig.GlobalConfig = previousConfig })
	r := chi.NewRouter()
	cfg := &platformconfig.AppConfig{
		RunMode:        "license",
		DeploymentMode: "self_hosted",
		LicenseKey:     "server-side-license-key",
	}
	RegisterLicenseAuthorityRoutes(r, cfg, nil)

	req := httptest.NewRequest(http.MethodPost, "/license/activate", strings.NewReader(`{"instance_id":"attacker-1"}`))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("activation without an explicit key = %d, want 400: %s", rec.Code, rec.Body.String())
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
