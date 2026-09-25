package origins

import (
	"path"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/platform/config"
)

func TestAllowedOriginsRequireHTTPSForHostedDomains(t *testing.T) {
	useTestEnvironment(t, "development")

	cases := []struct {
		origin string
		want   bool
	}{
		{origin: "http://localhost:3000", want: true},
		{origin: "https://rtylr.com", want: true},
		{origin: "https://dash.rtylr.com", want: true},
		{origin: "https://voxire.com", want: true},
		{origin: "https://api.voxire.tech", want: true},
		{origin: "http://rtylr.com", want: false},
		{origin: "http://dash.rtylr.com", want: false},
		{origin: "http://voxire.com", want: false},
		{origin: "http://api.voxire.tech", want: false},
	}

	for _, tc := range cases {
		t.Run(tc.origin, func(t *testing.T) {
			if got := IsAllowed(tc.origin); got != tc.want {
				t.Fatalf("IsAllowed(%q) = %v, want %v", tc.origin, got, tc.want)
			}
		})
	}
}

func TestAllowedPatternStringsUseWebSocketGlobSemantics(t *testing.T) {
	useTestEnvironment(t, "development")

	cases := []struct {
		target string
		want   bool
	}{
		{target: "http://localhost", want: true},
		{target: "http://localhost:3000", want: true},
		{target: "https://rtylr.com", want: true},
		{target: "https://dash.rtylr.com", want: true},
		{target: "https://voxire.com", want: true},
		{target: "https://api.voxire.tech", want: true},
		{target: "http://rtylr.com", want: false},
		{target: "http://dash.rtylr.com", want: false},
		{target: "http://voxire.com", want: false},
		{target: "http://api.voxire.tech", want: false},
	}

	patterns := AllowedPatternStrings()
	for _, tc := range cases {
		t.Run(tc.target, func(t *testing.T) {
			got := false
			for _, pattern := range patterns {
				matched, err := path.Match(strings.ToLower(pattern), strings.ToLower(tc.target))
				if err != nil {
					t.Fatalf("pattern %q is invalid: %v", pattern, err)
				}
				if matched {
					got = true
					break
				}
			}
			if got != tc.want {
				t.Fatalf("AllowedPatternStrings match %q = %v, want %v", tc.target, got, tc.want)
			}
		})
	}
}

func TestProductionBuildsExcludeLocalhostOrigins(t *testing.T) {
	useTestEnvironment(t, "production")

	for _, origin := range []string{
		"http://localhost:3000",
		"http://localhost",
		"http://127.0.0.1:5173",
	} {
		if IsAllowed(origin) {
			t.Errorf("production must reject dev origin %q", origin)
		}
	}
	for _, pattern := range AllowedPatternStrings() {
		if strings.Contains(pattern, "localhost") || strings.Contains(pattern, "127.0.0.1") {
			t.Errorf("production pattern list contains %q", pattern)
		}
	}

	// Hosted origins keep working in production.
	if !IsAllowed("https://dash.rtylr.com") {
		t.Error("hosted origin rejected in production")
	}
}

func TestWildcardsMatchExactlyOneSubdomainLabel(t *testing.T) {
	useTestEnvironment(t, "development")

	if !IsAllowed("https://app.voxire.com") {
		t.Error("single-label subdomain should match")
	}
	// A takeover chain two labels deep must not inherit apex trust (#777).
	if IsAllowed("https://evil.attacker.voxire.com") {
		t.Error("multi-label wildcard match allows nested takeover chains")
	}
}

func useTestEnvironment(t *testing.T, appEnv string) {
	t.Helper()

	previous := config.GlobalConfig
	t.Cleanup(func() { config.GlobalConfig = previous })
	config.GlobalConfig = nil
	t.Setenv("APP_ENV", appEnv)
}
