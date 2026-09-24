package trino

import (
	"context"
	"errors"
	"net/url"
	"sync"
	"testing"

	"github.com/voxire/onion/internal/platform/config"
)

func resetForTest(t *testing.T) {
	t.Helper()
	Close()
	clientInstance = nil
	clientOnce = sync.Once{}
	httpOnce = sync.Once{}
}

func TestConnect_DisabledIsNoOp(t *testing.T) {
	resetForTest(t)
	t.Setenv(config.TrinoEnabledEnvVar, "false")
	config.GlobalConfig = nil

	if err := Connect(context.Background()); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	if Available() {
		t.Fatal("expected trino unavailable when disabled")
	}
}

func TestConnect_EnabledButMisconfigured(t *testing.T) {
	resetForTest(t)
	t.Setenv(config.TrinoEnabledEnvVar, "true")
	t.Setenv(config.TrinoUserEnvVar, "")
	config.GlobalConfig = nil

	if err := Connect(context.Background()); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	if Available() {
		t.Fatal("expected trino unavailable when misconfigured")
	}
}

func TestQuery_ReturnsDisabledWhenOff(t *testing.T) {
	resetForTest(t)
	t.Setenv(config.TrinoEnabledEnvVar, "false")
	config.GlobalConfig = nil

	rows, err := Query(context.Background(), "SELECT 1")
	if rows != nil {
		defer rows.Close()
		if rowsErr := rows.Err(); rowsErr != nil {
			t.Fatalf("rows.Err: %v", rowsErr)
		}
	}
	if !errors.Is(err, ErrDisabled) {
		t.Fatalf("expected ErrDisabled, got %v", err)
	}
}

func TestQuery_ReturnsNotConnectedWhenEnabledButUnreachable(t *testing.T) {
	resetForTest(t)
	t.Setenv(config.TrinoEnabledEnvVar, "true")
	t.Setenv(config.TrinoHostEnvVar, "192.0.2.1")
	t.Setenv(config.TrinoPortEnvVar, "8080")
	t.Setenv(config.TrinoUserEnvVar, "analytics")
	config.GlobalConfig = nil

	if err := Connect(context.Background()); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	if Available() {
		t.Fatal("expected trino unavailable when connection fails")
	}

	rows, err := Query(context.Background(), "SELECT 1")
	if rows != nil {
		defer rows.Close()
		if rowsErr := rows.Err(); rowsErr != nil {
			t.Fatalf("rows.Err: %v", rowsErr)
		}
	}
	if !errors.Is(err, ErrNotConnected) {
		t.Fatalf("expected ErrNotConnected, got %v", err)
	}
}

func TestBuildConfigEscapesTrinoCredentials(t *testing.T) {
	t.Setenv(config.TrinoEnabledEnvVar, "true")
	t.Setenv(config.TrinoHostEnvVar, "trino.example.com")
	t.Setenv(config.TrinoPortEnvVar, "8443")
	t.Setenv(config.TrinoUserEnvVar, "analytics/user")
	t.Setenv(config.TrinoPasswordEnvVar, "p/secret?with@reserved:characters")
	t.Setenv(config.TrinoTLSEnvVar, "true")
	config.GlobalConfig = nil
	t.Cleanup(func() { config.GlobalConfig = nil })

	cfg, err := buildConfig()
	if err != nil {
		t.Fatalf("buildConfig: %v", err)
	}
	if _, err := cfg.FormatDSN(); err != nil {
		t.Fatalf("FormatDSN rejected reserved credentials: %v", err)
	}

	parsed, err := url.Parse(cfg.ServerURI)
	if err != nil {
		t.Fatalf("parse server URI: %v", err)
	}
	if got := parsed.User.Username(); got != "analytics/user" {
		t.Fatalf("username = %q, want analytics/user", got)
	}
	if got, ok := parsed.User.Password(); !ok || got != "p/secret?with@reserved:characters" {
		t.Fatalf("password = %q, present=%v", got, ok)
	}
	if parsed.Host != "trino.example.com:8443" {
		t.Fatalf("host = %q, want trino.example.com:8443", parsed.Host)
	}
}
