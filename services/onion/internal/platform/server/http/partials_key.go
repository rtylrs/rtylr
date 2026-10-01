package http

import (
	"crypto/subtle"
	"net/http"
	"os"
	"strings"

	authPkg "github.com/voxire/onion/internal/platform/auth"
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

// partialsActor is the identity the X-API-KEY bypass acts as. The console
// endpoints read the authenticated user from the request context, so the bypass
// has to supply one; it always assumes the console_admin role.
func partialsActor() *authPkg.AuthUser {
	var userID int64
	if config.GlobalConfig != nil {
		userID = config.GlobalConfig.PartialsActorUserID
	}

	return &authPkg.AuthUser{
		UserID: userID,
		IsConsoleAdmin: true,
		Token:  "partials-api-key",
	}
}

// PartialsBypassHandler sends requests that present a valid X-API-KEY straight
// to the bare router, skipping every global middleware (security, rate limit,
// auth, subscription, license, idempotency), and attaches a console_admin
// identity so console endpoints can authorize. Requests without the key — or
// with a wrong one — go through the normal middleware-protected handler.
//
// The bypass is only active while PARTIALS_API_KEY is configured. The actor id
// comes from PARTIALS_ACTOR_USER_ID (0 when unset), which is what audit and
// license writes record.
func PartialsBypassHandler(bare, protected http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hasValidPartialsAPIKey(r) {
			ctx := authPkg.WithAuthUser(r.Context(), partialsActor())
			bare.ServeHTTP(w, r.WithContext(ctx))

			return
		}

		protected.ServeHTTP(w, r)
	})
}
