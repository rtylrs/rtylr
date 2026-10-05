package cdn

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/tenant"
	"github.com/voxire/onion/internal/platform/utils/jwtutils"
)

type principalContextKey struct{}

// principal is who an authenticated CDN request runs as. businessID is the
// tenant every tenant-scoped query is filtered by; it is 0 when the request
// carries no tenant, and a 0 businessID is never treated as "all tenants".
type principal struct {
	userID     int64
	businessID int64
	// service marks the shared CDN_API_KEY credential. That secret is a
	// deployment operator credential, not a tenant credential: it is not
	// bound to a business, and holding it is already equivalent to database
	// access, so destructive platform operations stay available to it.
	service bool
}

func withPrincipal(ctx context.Context, p principal) context.Context {
	return context.WithValue(ctx, principalContextKey{}, p)
}

func principalFrom(ctx context.Context) (principal, bool) {
	p, ok := ctx.Value(principalContextKey{}).(principal)
	return p, ok
}

// mutationScope is the tenant a destructive operation (purge) or a
// credential-minting operation (sign) is confined to. ok is false when the
// request carries no principal at all, which must be a denial.
type mutationScope struct {
	tenant   int64
	platform bool
}

// read is the delivery scope this principal implies. The platform credential
// is a deployment operator secret rather than a tenant, so it reads every row;
// a tenant reads its own rows plus the publicly flagged ones.
func (s mutationScope) read() readScope {
	if s.platform {
		return signedRead()
	}
	return tenantRead(s.tenant)
}

// owns reports whether a resolved file row belongs to this principal. A row
// with no resolved business belongs to nobody and is never owned.
func (s mutationScope) owns(f *File) bool {
	if s.platform {
		return true
	}
	return f.BusinessID > 0 && f.BusinessID == s.tenant
}

func callerScope(r *http.Request) (mutationScope, bool) {
	p, ok := principalFrom(r.Context())
	if !ok {
		return mutationScope{}, false
	}
	if p.service {
		return mutationScope{platform: true}, true
	}
	if p.businessID > 0 {
		return mutationScope{tenant: p.businessID}, true
	}
	return mutationScope{}, false
}

// requireTenantScope returns the tenant a request may write under. The shared
// service credential is refused: a file recorded against a business_id of 0
// could never be read, purged or attributed again, so the service key is not
// allowed to create one.
func requireTenantScope(w http.ResponseWriter, r *http.Request) (int64, bool) {
	p, ok := principalFrom(r.Context())
	if !ok || p.service || p.businessID <= 0 {
		http.Error(w, "no tenant scope for this credential", http.StatusForbidden)
		return 0, false
	}
	return p.businessID, true
}

// apiKeyAuth guards write endpoints. It accepts either:
//
//  1. the shared CDN_API_KEY (a service-to-service credential for internal
//     callers / build steps),
//  2. any user's personal API key that carries the "upload" scope — created
//     and managed in Dash, stored hashed in hub_api_keys,
//  3. a normal signed app JWT, so browser apps can upload directly to the
//     upload runmode without proxying bytes through the API runmode.
//
// Paths 2 and 3 are resolved to a business through
// tenant.ResolveBusinessID, which verifies the caller actually has access to
// it. A token that is syntactically valid but names no reachable business is
// rejected: accepting it is what let any logged-in user act on any tenant's
// files.
//
// When no shared key is configured the middleware is open only in an explicit
// development environment. That open path authenticates nobody, so the request
// reaches the handler with no principal and every tenant-scoped handler denies
// it; only anonymous public delivery stays reachable.
func (c *CDN) apiKeyAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")

		// An explicitly configured development environment may intentionally run
		// without a shared service key. Empty/unknown environments fail closed.
		if c.apiKey == "" && cdnDevelopmentEnvironment() && token == "" {
			next.ServeHTTP(w, r.WithContext(withPrincipal(r.Context(), principal{})))
			return
		}
		// Constant-time compare so a timing side-channel can't reveal the key.
		if c.apiKey != "" && token != "" && subtle.ConstantTimeCompare([]byte(token), []byte(c.apiKey)) == 1 {
			next.ServeHTTP(w, r.WithContext(withPrincipal(r.Context(), principal{service: true})))
			return
		}

		// Per-user key path: a personal `upload`-scoped key from hub_api_keys.
		if token != "" {
			sum := sha256.Sum256([]byte(token))
			if ownerID, keyBusinessID, ok := c.uploadKeyOwnerByHash(r.Context(), hex.EncodeToString(sum[:])); ok {
				if p, ok := c.tenantPrincipal(r.Context(), keyBusinessID, ownerID); ok {
					next.ServeHTTP(w, r.WithContext(withPrincipal(r.Context(), p)))
					return
				}
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			if claims, ok := validAppJWT(r, token); ok {
				if p, ok := c.tenantPrincipal(r.Context(), claims.BusinessID, claims.UserID); ok {
					next.ServeHTTP(w, r.WithContext(withPrincipal(r.Context(), p)))
					return
				}
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
		}

		http.Error(w, "unauthorized", http.StatusUnauthorized)
	})
}

// tenantPrincipal maps an authenticated identity onto the business its requests
// are scoped by. It fails closed: an unresolvable business is a denial, never
// a usable id of 0.
func (c *CDN) tenantPrincipal(ctx context.Context, businessID, userID int64) (principal, bool) {
	if userID <= 0 {
		return principal{}, false
	}
	resolved, err := tenant.ResolveBusinessID(ctx, businessID, userID)
	if err != nil || resolved <= 0 {
		// Logged because a token that verifies but names no reachable business
		// is either a stale session or an attempt to reach another tenant's
		// files, and the two are indistinguishable from the denial alone.
		logger.Ctx(ctx).Warn("cdn: credential resolved to no business scope",
			logger.Any("user_id", userID),
			logger.Any("claimed_business_id", businessID),
			logger.Error(err),
		)
		return principal{}, false
	}
	return principal{userID: userID, businessID: resolved}, true
}

// deliveryAuth resolves the read scope for the public delivery routes. A
// request with no bearer token is anonymous and sees only assets explicitly
// flagged public. A bearer token that does not resolve to a business falls back
// to the anonymous scope rather than erroring, so a stale token on a public
// image degrades to the public surface instead of breaking the asset — it can
// never widen it.
func (c *CDN) deliveryAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scope := anonymousRead
		if token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "); token != "" {
			sum := sha256.Sum256([]byte(token))
			if ownerID, keyBusinessID, ok := c.uploadKeyOwnerByHash(r.Context(), hex.EncodeToString(sum[:])); ok {
				if p, ok := c.tenantPrincipal(r.Context(), keyBusinessID, ownerID); ok {
					scope = tenantRead(p.businessID)
				}
			} else if claims, ok := validAppJWT(r, token); ok {
				if p, ok := c.tenantPrincipal(r.Context(), claims.BusinessID, claims.UserID); ok {
					scope = tenantRead(p.businessID)
				}
			}
		}
		next.ServeHTTP(w, r.WithContext(withDeliveryScope(r.Context(), scope)))
	})
}

type deliveryScopeContextKey struct{}

func withDeliveryScope(ctx context.Context, scope readScope) context.Context {
	return context.WithValue(ctx, deliveryScopeContextKey{}, scope)
}

func deliveryScopeFrom(ctx context.Context) readScope {
	scope, ok := ctx.Value(deliveryScopeContextKey{}).(readScope)
	if !ok {
		return anonymousRead
	}
	return scope
}

func cdnDevelopmentEnvironment() bool {
	env := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
	if config.GlobalConfig != nil && strings.TrimSpace(config.GlobalConfig.AppEnv) != "" {
		env = strings.ToLower(strings.TrimSpace(config.GlobalConfig.AppEnv))
	}
	switch env {
	case "dev", "development", "local", "test":
		return true
	default:
		return false
	}
}

// ValidateSecurityConfig lets the upload runmode fail at startup rather than
// waiting for the first write request to reveal a missing production secret.
func ValidateSecurityConfig(appEnv, apiKey string) error {
	if strings.TrimSpace(apiKey) == "" {
		env := strings.ToLower(strings.TrimSpace(appEnv))
		switch env {
		case "dev", "development", "local", "test":
			return nil
		default:
			return fmt.Errorf("CDN_API_KEY is required outside development environments")
		}
	}
	return nil
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

func validAppJWT(r *http.Request, token string) (uploadJWTClaims, bool) {
	secret := uploadJWTSecret()
	if token == "" || secret == "" {
		return uploadJWTClaims{}, false
	}

	var claims uploadJWTClaims
	helper := jwtutils.NewJWTHelper(secret, "onion-platform")
	if err := helper.ValidateToken(r.Context(), token, &claims); err != nil {
		return uploadJWTClaims{}, false
	}

	return claims, claims.UserID > 0
}
