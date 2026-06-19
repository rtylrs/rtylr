package cdn

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/utils/jwtutils"
)

// apiKeyAuth guards write endpoints. It accepts either:
//
//  1. the shared CDN_API_KEY (a service-to-service credential for internal
//     callers / build steps), or
//  2. any user's personal API key that carries the "upload" scope — created
//     and managed in Dash, stored hashed in hub_api_keys.
//  3. a normal signed app JWT, so browser apps can upload directly to the
//     upload runmode without proxying bytes through the API runmode.
//
// When no shared key is configured the middleware is a no-op (local dev only),
// preserving the historical "auth disabled if unset" behaviour. The user-key
// path stays active regardless so DB-backed keys work in every environment.
func (c *CDN) apiKeyAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")

		// Shared key path (back-compat). Empty configured key => dev-open.
		if c.apiKey == "" {
			next.ServeHTTP(w, r)
			return
		}
		// Constant-time compare so a timing side-channel can't reveal the key.
		if token != "" && subtle.ConstantTimeCompare([]byte(token), []byte(c.apiKey)) == 1 {
			next.ServeHTTP(w, r)
			return
		}

		// Per-user key path: a personal `upload`-scoped key from hub_api_keys.
		if token != "" {
			sum := sha256.Sum256([]byte(token))
			if _, ok := c.uploadKeyOwnerByHash(r.Context(), hex.EncodeToString(sum[:])); ok {
				next.ServeHTTP(w, r)
				return
			}

			if validAppJWT(r, token) {
				next.ServeHTTP(w, r)
				return
			}
		}

		http.Error(w, "unauthorized", http.StatusUnauthorized)
	})
}

type uploadJWTClaims struct {
	UserID         int64  `json:"user_id"`
	BusinessID     int64  `json:"business_id,omitempty"`
	Role           string `json:"role,omitempty"`
	Platform       string `json:"platform,omitempty"`
	PlatformUserID int64  `json:"platform_user_id,omitempty"`
	PlanID         int64  `json:"plan_id,omitempty"`
	jwt.RegisteredClaims
}

func (c *uploadJWTClaims) SetIssuer(issuer string) {
	c.Issuer = issuer
}

func (c *uploadJWTClaims) SetExpiration(exp time.Time) {
	c.ExpiresAt = jwt.NewNumericDate(exp)
}

func uploadJWTSecret() string {
	if config.GlobalConfig != nil && config.GlobalConfig.JWTSecret != "" {
		return config.GlobalConfig.JWTSecret
	}
	return os.Getenv("JWT_SECRET")
}

func validAppJWT(r *http.Request, token string) bool {
	secret := uploadJWTSecret()
	if token == "" || secret == "" {
		return false
	}

	var claims uploadJWTClaims
	helper := jwtutils.NewJWTHelper(secret, "onion-platform")
	if err := helper.ValidateToken(r.Context(), token, &claims); err != nil {
		return false
	}

	return claims.UserID > 0
}
