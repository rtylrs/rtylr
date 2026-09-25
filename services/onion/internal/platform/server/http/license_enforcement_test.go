package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/license"
)

type stubValidator struct {
	report license.StatusReport
	calls  int
}

func (s *stubValidator) Current(_ context.Context) license.StatusReport {
	s.calls++
	return s.report
}

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func withConfig(t *testing.T, cfg *config.AppConfig) func() {
	t.Helper()
	prev := config.GlobalConfig
	config.GlobalConfig = cfg
	return func() { config.GlobalConfig = prev }
}

func TestLicenseEnforcementHostedIsNoOp(t *testing.T) {
	defer withConfig(t, &config.AppConfig{DeploymentMode: license.ModeHosted})()

	validator := &stubValidator{report: license.StatusReport{Valid: false, Status: license.StatusExpired}}
	mw := LicenseEnforcementMiddleware(validator, NewRouteAccessPolicy())

	rec := httptest.NewRecorder()
	mw(okHandler()).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/erp/product/v1/products", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("hosted mode must pass through, got %d", rec.Code)
	}
	if validator.calls != 0 {
		t.Fatalf("hosted mode must not consult the validator, calls=%d", validator.calls)
	}
}

func TestLicenseEnforcementSelfHostInvalidBlocks(t *testing.T) {
	defer withConfig(t, &config.AppConfig{DeploymentMode: license.ModeSelfHosted})()

	validator := &stubValidator{report: license.StatusReport{Valid: false, Status: license.StatusExpired}}
	mw := LicenseEnforcementMiddleware(validator, NewRouteAccessPolicy())

	rec := httptest.NewRecorder()
	mw(okHandler()).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/erp/product/v1/products", nil))

	if rec.Code != http.StatusPaymentRequired {
		t.Fatalf("self-host invalid must return 402, got %d", rec.Code)
	}
}

func TestLicenseEnforcementSelfHostValidPasses(t *testing.T) {
	defer withConfig(t, &config.AppConfig{DeploymentMode: license.ModeSelfHosted})()

	validator := &stubValidator{report: license.StatusReport{Valid: true, Status: license.StatusValid}}
	mw := LicenseEnforcementMiddleware(validator, NewRouteAccessPolicy())

	rec := httptest.NewRecorder()
	mw(okHandler()).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/erp/product/v1/products", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("self-host valid must pass through, got %d", rec.Code)
	}
}

func TestLicenseEnforcementUtilityRoutesAlwaysPass(t *testing.T) {
	defer withConfig(t, &config.AppConfig{DeploymentMode: license.ModeSelfHosted})()

	validator := &stubValidator{report: license.StatusReport{Valid: false, Status: license.StatusMissing}}
	mw := LicenseEnforcementMiddleware(validator, NewRouteAccessPolicy())

	for path := range licenseUtilityRoutes {
		rec := httptest.NewRecorder()
		mw(okHandler()).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("utility route %q must pass even when invalid, got %d", path, rec.Code)
		}
	}
}

func TestLicenseEnforcementAnonymousRoutesPass(t *testing.T) {
	defer withConfig(t, &config.AppConfig{DeploymentMode: license.ModeSelfHosted})()

	policy := NewRouteAccessPolicy()
	policy.RegisterAnonymousRoutes("/login")

	validator := &stubValidator{report: license.StatusReport{Valid: false, Status: license.StatusMissing}}
	mw := LicenseEnforcementMiddleware(validator, policy)

	rec := httptest.NewRecorder()
	mw(okHandler()).ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/login", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("anonymous route must pass even when invalid, got %d", rec.Code)
	}
}
