package config

import "testing"

// LicenseAuthorityURL was assigned the constant directly and never read an
// environment variable, while its own doc comment named LICENSE_AUTHORITY_URL as
// the override "for our own testing". The result: a self-hosted backend could
// only ever validate against the hosted authority, so an air-gapped install
// could not work at all and self-host licensing could not be tested offline.
// Once the grace window elapses, LicenseEnforcementMiddleware answers 402 on
// every authenticated route, not just billing ones (#1079).
func TestLicenseAuthorityURLIsOverridable(t *testing.T) {
	setLicenseConfigEnv(t)
	GlobalConfig = nil

	t.Setenv("LICENSE_AUTHORITY_URL", "http://127.0.0.1:8123")
	cfg, err := LoadAndValidate()
	if err != nil {
		t.Fatalf("LoadAndValidate: %v", err)
	}
	if cfg.LicenseAuthorityURL != "http://127.0.0.1:8123" {
		t.Errorf("LicenseAuthorityURL = %q, want the value from the environment", cfg.LicenseAuthorityURL)
	}
}

// The override is not a documented customer knob, so an unset or blank value
// must still land on the hosted authority rather than on an empty string.
func TestLicenseAuthorityURLFallsBackToTheDefault(t *testing.T) {
	setLicenseConfigEnv(t)
	GlobalConfig = nil

	for _, value := range []string{"", "   "} {
		t.Setenv("LICENSE_AUTHORITY_URL", value)
		GlobalConfig = nil
		cfg, err := LoadAndValidate()
		if err != nil {
			t.Fatalf("LoadAndValidate with %q: %v", value, err)
		}
		if cfg.LicenseAuthorityURL != DefaultLicenseAuthorityURL {
			t.Errorf("LicenseAuthorityURL = %q for input %q, want the default %q",
				cfg.LicenseAuthorityURL, value, DefaultLicenseAuthorityURL)
		}
	}
}
