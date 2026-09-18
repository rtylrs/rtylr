package http

import (
	"crypto/subtle"
	"net/http"
	"os"
	"strings"

	"github.com/voxire/onion/internal/platform/config"
)

// PartialsAPIKeyHeader is the request header Partials sends to identify itself
// to internal endpoints.
const PartialsAPIKeyHeader = "X-API-KEY"

// PartialsAPIKeyEnvVar is the environment variable (sourced from Secrets
// Manager in production) that holds the shared internal key.
const PartialsAPIKeyEnvVar = "PARTIALS_API_KEY"

func partialsAPIKey() string {
	if config.GlobalConfig != nil {
		return config.GlobalConfig.PartialsAPIKey
	}
	return strings.TrimSpace(os.Getenv(PartialsAPIKeyEnvVar))
}

// hasValidPartialsAPIKey reports whether the request carries the configured
// internal key. It returns false when no key is configured so the bypass is
// impossible to trigger by accident.
func hasValidPartialsAPIKey(r *http.Request) bool {
	expected := partialsAPIKey()
	if expected == "" {
		return false
	}

	actual := r.Header.Get(PartialsAPIKeyHeader)
	if actual == "" {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(actual), []byte(expected)) == 1
}

// PartialsBypassHandler sends requests that present a valid X-API-KEY straight
// to the bare router, skipping every global middleware (security, rate limit,
// auth, subscription, license, idempotency). Requests without the key — or with
// a wrong one — go through the normal middleware-protected handler.
//
// The bypass is only active while PARTIALS_API_KEY is configured, and it is
// intended for internal endpoints only: it grants the caller the same reach as
// an unauthenticated request to every route, with no user identity attached.
func PartialsBypassHandler(bare, protected http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hasValidPartialsAPIKey(r) {
			bare.ServeHTTP(w, r)

			return
		}

		protected.ServeHTTP(w, r)
	})
}
