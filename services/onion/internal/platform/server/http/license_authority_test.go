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
