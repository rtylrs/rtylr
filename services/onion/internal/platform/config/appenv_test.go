package config

import (
	"bytes"
	"log/slog"
	"os"
	"strings"
	"testing"
)

// withAppEnv isolates the process environment for one case: it drops any
// loaded configuration and sets APP_ENV, so the test exercises the same
// resolution path a real run takes when LoadAndValidate has not run.
func withAppEnv(t *testing.T, env string) {
	t.Helper()
	previous := GlobalConfig
	t.Cleanup(func() { GlobalConfig = previous })
	GlobalConfig = nil
	t.Setenv(AppEnvEnvVar, env)
}

// The production-ish spellings are the ones the previous equality check got
// wrong. PROD and Production failed only because of case folding; the rest
// failed because "equals production or prod" cannot answer a question about a
// value it has never seen.
func TestIsProductionDefaultsClosedForUnrecognisedValues(t *testing.T) {
	cases := []struct {
		name string
		env  string
		want bool
	}{
		{name: "production", env: "production", want: true},
		{name: "prod", env: "prod", want: true},
		{name: "PROD", env: "PROD", want: true},
		{name: "Production", env: "Production", want: true},
		{name: "PRODUCTION", env: "PRODUCTION", want: true},
		{name: "production-eu", env: "production-eu", want: true},
		{name: "padded prod", env: " prod ", want: true},
		{name: "padded empty", env: "   ", want: true},
		{name: "empty", env: "", want: true},
		{name: "nonsense", env: "banana", want: true},
		{name: "staging", env: "staging", want: true},
		{name: "dev", env: "dev", want: false},
		{name: "development", env: "development", want: false},
		{name: "Development", env: "Development", want: false},
		{name: "padded development", env: " development ", want: false},
		{name: "local", env: "local", want: false},
		{name: "test", env: "test", want: false},
		{name: "TEST", env: "TEST", want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			withAppEnv(t, tc.env)
			if got := IsProduction(); got != tc.want {
				t.Errorf("IsProduction() with APP_ENV=%q = %v, want %v", tc.env, got, tc.want)
			}
		})
	}
}

// An unset APP_ENV is the exact state the finding's exploit relies on: the
// equality check answered "not production" for a variable that was never set.
func TestUnsetAppEnvIsProduction(t *testing.T) {
	previous := GlobalConfig
	t.Cleanup(func() { GlobalConfig = previous })
	GlobalConfig = nil
	t.Setenv(AppEnvEnvVar, "placeholder")
	if err := os.Unsetenv(AppEnvEnvVar); err != nil {
		t.Fatalf("unset %s: %v", AppEnvEnvVar, err)
	}

	if !IsProduction() {
		t.Error("an unset APP_ENV must resolve to production")
	}
	if IsNonProduction() {
		t.Error("an unset APP_ENV must not resolve to a development environment")
	}
	if IsDeclaredProduction() {
		t.Error("IsDeclaredProduction must stay false when APP_ENV is unset")
	}
}

// A loaded configuration wins over the environment, but an empty AppEnv in the
// configuration must not mask a real environment value the way a plain
// GlobalConfig.AppEnv read would.
func TestIsProductionPrefersLoadedConfigWithoutMaskingEmptyAppEnv(t *testing.T) {
	previous := GlobalConfig
	t.Cleanup(func() { GlobalConfig = previous })
	t.Setenv(AppEnvEnvVar, "development")

	GlobalConfig = &AppConfig{AppEnv: "production"}
	if !IsProduction() {
		t.Error("a loaded APP_ENV=production must win over the environment")
	}

	GlobalConfig = &AppConfig{AppEnv: "  "}
	if !IsNonProduction() {
		t.Error("a blank AppEnv in the loaded config must fall back to the environment")
	}
}

func TestIsDeclaredProductionSeparatesDeclarationFromAbsence(t *testing.T) {
	cases := []struct {
		env  string
		want bool
	}{
		{env: "production", want: true},
		{env: "prod", want: true},
		{env: " PROD ", want: true},
		{env: "production-eu", want: false},
		{env: "staging", want: false},
		{env: "development", want: false},
		{env: "", want: false},
	}

	for _, tc := range cases {
		t.Run(tc.env, func(t *testing.T) {
			withAppEnv(t, tc.env)
			if got := IsDeclaredProduction(); got != tc.want {
				t.Errorf("IsDeclaredProduction() with APP_ENV=%q = %v, want %v", tc.env, got, tc.want)
			}
		})
	}
}

func TestNormalizeAppEnvTrimsAndFolds(t *testing.T) {
	cases := map[string]string{
		"production":  "production",
		"  PROD  ":    "prod",
		"Production":  "production",
		"\tDeV\n":     "dev",
		"":            "",
		"   ":         "",
		"banana":      "banana",
		"prod-uk-1  ": "prod-uk-1",
	}

	for raw, want := range cases {
		if got := NormalizeAppEnv(raw); got != want {
			t.Errorf("NormalizeAppEnv(%q) = %q, want %q", raw, got, want)
		}
	}
}

// The boot assertion is the only thing that tells an operator their APP_ENV
// was not understood, so it has to fire on exactly the surprising states and
// stay quiet otherwise.
func TestAssertAppEnvIsUnderstoodReportsUnexpectedEnvironments(t *testing.T) {
	cases := []struct {
		name       string
		env        string
		wantLogged bool
	}{
		{name: "unset", env: "", wantLogged: true},
		{name: "whitespace", env: "  ", wantLogged: true},
		{name: "nonsense", env: "banana", wantLogged: true},
		{name: "staging", env: "staging", wantLogged: true},
		{name: "production", env: "production", wantLogged: false},
		{name: "development", env: "development", wantLogged: false},
		{name: "local", env: "local", wantLogged: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			previousDefault := slog.Default()
			slog.SetDefault(slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			})))
			slog.Info("unrelated")
			t.Cleanup(func() { slog.SetDefault(previousDefault) })

			assertAppEnvIsUnderstood(tc.env)

			logged := buf.String()
			if tc.wantLogged {
				if !strings.Contains(logged, "APP_ENV") {
					t.Errorf("expected APP_ENV to be reported, got: %s", logged)
				}
				return
			}
			if strings.Contains(logged, "APP_ENV is not set") || strings.Contains(logged, "not a recognised environment") {
				t.Errorf("unexpected APP_ENV complaint for %q, got: %s", tc.env, logged)
			}
		})
	}
}

// LoadAndValidate must not refuse to boot on an APP_ENV it does not
// recognise. The default is already the safe one, so an unknown label is a
// warning to read, not an outage to create.
func TestLoadAndValidateAcceptsUnrecognisedAppEnv(t *testing.T) {
	setLicenseConfigEnv(t)
	t.Setenv(AppEnvEnvVar, "production-eu")

	if _, err := LoadAndValidate(); err != nil {
		t.Fatalf("an unrecognised APP_ENV must not fail validation: %v", err)
	}
	if !IsProduction() {
		t.Error("an unrecognised APP_ENV must resolve to production")
	}
}
