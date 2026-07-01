package config

import (
	"os"
	"testing"
)

func TestTrinoConfig_DisabledByDefault(t *testing.T) {
	t.Setenv(TrinoEnabledEnvVar, "")
	GlobalConfig = nil

	if IsTrinoEnabled() {
		t.Fatal("expected trino disabled by default")
	}
	if IsTrinoConfigured() {
		t.Fatal("expected trino not configured when disabled")
	}
}

func TestTrinoConfig_EnabledRequiresUser(t *testing.T) {
	t.Setenv(TrinoEnabledEnvVar, "true")
	t.Setenv(TrinoHostEnvVar, "trino.example.com")
	t.Setenv(TrinoUserEnvVar, "")
	GlobalConfig = nil

	if IsTrinoConfigured() {
		t.Fatal("expected trino not configured without user")
	}
}

func TestTrinoConfig_EnabledWithMinimumSettings(t *testing.T) {
	t.Setenv(TrinoEnabledEnvVar, "true")
	t.Setenv(TrinoHostEnvVar, "trino.example.com")
	t.Setenv(TrinoPortEnvVar, "8443")
	t.Setenv(TrinoUserEnvVar, "analytics")
	t.Setenv(TrinoCatalogEnvVar, "iceberg")
	t.Setenv(TrinoSchemaEnvVar, "rtylr")
	t.Setenv(TrinoTLSEnvVar, "true")
	GlobalConfig = nil

	if !IsTrinoEnabled() {
		t.Fatal("expected trino enabled")
	}
	if !IsTrinoConfigured() {
		t.Fatal("expected trino configured")
	}
	if GetTrinoHost() != "trino.example.com" {
		t.Fatalf("host = %q", GetTrinoHost())
	}
	if GetTrinoPort() != "8443" {
		t.Fatalf("port = %q", GetTrinoPort())
	}
	if GetTrinoUser() != "analytics" {
		t.Fatalf("user = %q", GetTrinoUser())
	}
	if GetTrinoCatalog() != "iceberg" {
		t.Fatalf("catalog = %q", GetTrinoCatalog())
	}
	if GetTrinoSchema() != "rtylr" {
		t.Fatalf("schema = %q", GetTrinoSchema())
	}
	if !IsTrinoTLS() {
		t.Fatal("expected TLS enabled")
	}
}

func TestTrinoConfig_DefaultsWhenUnset(t *testing.T) {
	for _, key := range []string{
		TrinoHostEnvVar,
		TrinoPortEnvVar,
		TrinoSourceEnvVar,
	} {
		_ = os.Unsetenv(key)
	}
	GlobalConfig = nil

	if GetTrinoHost() != DefaultTrinoHost {
		t.Fatalf("host default = %q", GetTrinoHost())
	}
	if GetTrinoPort() != DefaultTrinoPort {
		t.Fatalf("port default = %q", GetTrinoPort())
	}
	if GetTrinoSource() != DefaultTrinoSource {
		t.Fatalf("source default = %q", GetTrinoSource())
	}
}
