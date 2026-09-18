package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/voxire/onion/internal/platform/config"
)

func withPartialsAPIKey(t *testing.T, key string) {
	t.Helper()
	previous := config.GlobalConfig
	config.GlobalConfig = &config.AppConfig{PartialsAPIKey: key}
	t.Cleanup(func() { config.GlobalConfig = previous })
}

func TestPartialsBypassRoutesValidKeyToBareHandler(t *testing.T) {
	withPartialsAPIKey(t, "pka_test_key")

	bareHit := false
	bare := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) { bareHit = true })
	protected := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})

	req := httptest.NewRequest(http.MethodGet, "/hub/user/v1/users/me", nil)
	req.Header.Set(PartialsAPIKeyHeader, "pka_test_key")
	rec := httptest.NewRecorder()

	PartialsBypassHandler(bare, protected).ServeHTTP(rec, req)

	if !bareHit {
		t.Fatalf("expected the bare handler to run for a valid %s", PartialsAPIKeyHeader)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected bare handler to own the response, got status %d", rec.Code)
	}
}

func TestPartialsBypassFallsThroughWithoutValidKey(t *testing.T) {
	withPartialsAPIKey(t, "pka_test_key")

	cases := map[string]string{
		"missing header": "",
		"wrong key":      "pka_not_the_key",
	}

	for name, header := range cases {
		t.Run(name, func(t *testing.T) {
			protectedHit := false
			bare := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {})
			protected := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				protectedHit = true
				w.WriteHeader(http.StatusUnauthorized)
			})

			req := httptest.NewRequest(http.MethodGet, "/hub/user/v1/users/me", nil)
			if header != "" {
				req.Header.Set(PartialsAPIKeyHeader, header)
			}
			rec := httptest.NewRecorder()

			PartialsBypassHandler(bare, protected).ServeHTTP(rec, req)

			if !protectedHit {
				t.Fatalf("expected the protected handler to run")
			}
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("expected protected status, got %d", rec.Code)
			}
		})
	}
}

func TestPartialsBypassDisabledWhenKeyUnset(t *testing.T) {
	withPartialsAPIKey(t, "")

	protectedHit := false
	bare := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {})
	protected := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		protectedHit = true
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set(PartialsAPIKeyHeader, "anything")
	rec := httptest.NewRecorder()

	PartialsBypassHandler(bare, protected).ServeHTTP(rec, req)

	if !protectedHit {
		t.Fatalf("expected bypass to stay disabled when no key is configured")
	}
}

func TestHasValidPartialsAPIKeyReadsEnvWhenConfigMissing(t *testing.T) {
	previous := config.GlobalConfig
	config.GlobalConfig = nil
	t.Cleanup(func() { config.GlobalConfig = previous })
	t.Setenv(PartialsAPIKeyEnvVar, "pka_env_key")

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set(PartialsAPIKeyHeader, "pka_env_key")

	if !hasValidPartialsAPIKey(req) {
		t.Fatalf("expected env-configured key to validate")
	}
}
