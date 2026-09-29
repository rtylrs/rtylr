package http

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

const (
	AuthHeaderKey = "Authorization"
	BearerPrefix  = "Bearer "
)

func writeUnauthorizedResponse(ctx context.Context, w http.ResponseWriter, message string) {
	WriteOneKitError(ctx, w, onekiterrors.New(http.StatusUnauthorized, message))
}

func extractToken(authHeader string) string {
	if strings.HasPrefix(authHeader, BearerPrefix) {
		return strings.TrimPrefix(authHeader, BearerPrefix)
	}

	return authHeader
}

func processTokenAuthentication(ctx context.Context, token, path, remoteAddr string, tokenValidator authPkg.TokenValidator) (*authPkg.AuthUser, error) {
	if tokenValidator == nil {
		logger.Ctx(ctx).Error("TokenValidator is not available for token validation")

		return nil, onekiterrors.New(500, "Authentication service unavailable")
	}

	claims, err := tokenValidator.ValidateTokenFull(ctx, token)
	if err != nil {
		logger.Ctx(ctx).Warn("Invalid token",
			logger.Error(err),
			logger.String("path", path),
			logger.String("ip", remoteAddr),
		)

		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	authUser := &authPkg.AuthUser{
		UserID:         claims.UserID,
		BusinessID:     claims.BusinessID,
		Role:           claims.Role,
		Platform:       claims.Platform,
		PlatformUserID: claims.PlatformUserID,
		PlanID:         claims.PlanID,
		Token:          token,
		IP:             remoteAddr,
	}

	if authSuccessLogsEnabled() {
		logger.Ctx(ctx).Info("Authenticated request",
			logger.Any("user_id", authUser.UserID),
			logger.String("ip", remoteAddr),
			logger.String("path", path),
		)
	}

	return authUser, nil
}

func authSuccessLogsEnabled() bool {
	return envFlag("AUTH_SUCCESS_LOGS") || envFlag("AUTH_REQUEST_LOGS")
}

func GetAuthMiddlewareFunc(accessPolicy *RouteAccessPolicy, tokenValidatorResolver authPkg.TokenValidatorResolver, accessResolvers ...authPkg.EffectiveAccessResolver) MiddlewareFunc {
	// MetricsHandler owns the /metrics basic-auth check. Mark the route
	// anonymous here so the bearer gate (and the downstream subscription and
	// license gates, which consult the same policy) let that handler decide
	// whether the scraper is authorized.
	if accessPolicy != nil {
		accessPolicy.RegisterAnonymousRoutes("/metrics")
	}

	var accessResolver authPkg.EffectiveAccessResolver
	if len(accessResolvers) > 0 {
		accessResolver = accessResolvers[0]
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			userAgent := r.Header.Get("User-Agent")

			productPlatform := authPkg.ResolveProductPlatform(r)
			ctx := authPkg.WithProductPlatform(r.Context(), productPlatform)
			r = r.WithContext(ctx)

			if accessPolicy != nil && accessPolicy.IsAllowedAnonymously(path) {
				next.ServeHTTP(w, r)

				return
			}

			if accessPolicy != nil && accessPolicy.IsAllowedForAgent(userAgent, path) {
				next.ServeHTTP(w, r)

				return
			}

			authHeader := r.Header.Get(AuthHeaderKey)
			if authHeader == "" {
				writeUnauthorizedResponse(ctx, w, "Unauthorized")

				return
			}

			token := extractToken(authHeader)

			var tokenValidator authPkg.TokenValidator
			if tokenValidatorResolver != nil {
				tokenValidator = tokenValidatorResolver.ResolveTokenValidator(ctx, authPkg.ProductPlatformCtx(ctx))
			}

			authUser, err := processTokenAuthentication(ctx, token, path, r.RemoteAddr, tokenValidator)
			if err != nil {
				writeUnauthorizedResponse(ctx, w, "Unauthorized")

				return
			}
			if accessResolver != nil {
				effectiveAccess, accessErr := accessResolver.GetEffectiveAccess(ctx, authUser.BusinessID, authUser.UserID)
				if accessErr != nil {
					logger.Ctx(ctx).Warn("Effective authorization resolution failed", logger.Error(accessErr), logger.String("path", path))
					writeUnauthorizedResponse(ctx, w, "Authorization unavailable")
					return
				}
				authUser.EffectiveAccess = effectiveAccess
				if appID, isAppPath := appForPath(path); isAppPath && !authPkg.HasAppAccess(authUser, appID) {
					WriteOneKitError(ctx, w, onekiterrors.PermissionDenied())
					return
				}
				if appID, permission, isFeaturePath := featurePermissionForRequest(r.Method, path); isFeaturePath && !hasFeaturePermission(authUser, appID, permission) {
					WriteOneKitError(ctx, w, onekiterrors.PermissionDenied())
					return
				}
				if statusResolver, ok := accessResolver.(authPkg.ProvisioningStatusResolver); ok {
					var status *authorization.ProvisioningStatus
					var statusErr error
					if businessResolver, scoped := accessResolver.(authPkg.BusinessProvisioningStatusResolver); scoped {
						status, statusErr = businessResolver.GetProvisioningStatusForBusiness(ctx, authUser.BusinessID, authUser.UserID)
					} else {
						status, statusErr = statusResolver.GetProvisioningStatus(ctx, authUser.UserID)
					}
					if statusErr != nil {
						logger.Ctx(ctx).Warn("Provisioning status resolution failed", logger.Error(statusErr), logger.String("path", path))
						writeUnauthorizedResponse(ctx, w, "Authorization unavailable")
						return
					}
					authUser.ProvisioningStatus = status
					if status != nil && !isProvisioningPath(path) {
						if !status.EmailVerified {
							writeUnauthorizedResponse(ctx, w, "Email verification required")
							return
						}
						if status.PasswordChangeRequired {
							writeUnauthorizedResponse(ctx, w, "Account activation required")
							return
						}
						if status.MFARequired {
							if !status.MFAEnrolled && !isMFAPath(r, false) {
								writeUnauthorizedResponse(ctx, w, "MFA enrollment required")
								return
							}
							if status.MFAEnrolled && !isMFAPath(r, true) {
								mfaResolver, ok := accessResolver.(authPkg.MFAVerifiedResolver)
								if !ok {
									writeUnauthorizedResponse(ctx, w, "MFA verification unavailable")
									return
								}
								var verified bool
								var verifyErr error
								if businessResolver, scoped := accessResolver.(authPkg.BusinessMFAVerifiedResolver); scoped {
									verified, verifyErr = businessResolver.IsMFAVerifiedForBusiness(ctx, authUser.BusinessID, authUser.UserID)
								} else {
									verified, verifyErr = mfaResolver.IsMFAVerified(ctx, authUser.UserID)
								}
								if verifyErr != nil || !verified {
									writeUnauthorizedResponse(ctx, w, "MFA verification required")
									return
								}
							}
						}
					}
				}
			}

			newCtx := context.WithValue(ctx, authPkg.AuthUserKey, authUser)
			next.ServeHTTP(w, r.WithContext(newCtx))
		})
	}
}

func hasFeaturePermission(user *authPkg.AuthUser, appID authorization.AppID, permission string) bool {
	if appID == authorization.AppHub {
		// Owners and admins are granted every dashboard permission by the
		// resolver, but a legacy account with no normalized membership row yet
		// resolves to no effective access. Fall back to the same owner/admin
		// boundary the concrete handlers enforce so those accounts are not
		// denied at the edge before their handler can apply that rule.
		return authPkg.HasDashboardPermission(user, permission) || authPkg.IsBusinessAdmin(user)
	}
	return authPkg.HasAppPermission(user, appID, permission)
}

func appForPath(path string) (authorization.AppID, bool) {
	normalized := strings.ToLower(strings.TrimRight(path, "/"))
	if normalized == "/hub/menu" || strings.HasPrefix(normalized, "/hub/menu/") {
		return authorization.AppMenu, true
	}
	for _, app := range []authorization.AppID{
		authorization.AppPOS,
		authorization.AppERP,
		authorization.AppHR,
		authorization.AppCRM,
		authorization.AppFinance,
		authorization.AppFlow,
		authorization.AppInsights,
		authorization.AppMenu,
		authorization.AppRecruit,
	} {
		prefix := "/" + string(app)
		if normalized == prefix || strings.HasPrefix(normalized, prefix+"/") {
			return app, true
		}
	}
	return "", false
}

// isMFAPath reports whether a request may bypass the MFA gate.
//
// Reads and proof endpoints (state, verify, recovery) always bypass: they are
// how a user satisfies the gate in the first place.
//
// Enrollment bypasses only the "enrollment required" branch, because it is the
// only way out of that state. Once a factor exists the same endpoints are a
// factor rotation, so they must clear the verification gate like any other
// mutation. Everything else — PATCH /2fa, POST /invalidate-sessions — is a
// security-setting mutation and gets no exemption at all: exempting them let a
// stolen bearer token disable or replace the second factor outright.
func isMFAPath(r *http.Request, mfaEnrolled bool) bool {
	switch strings.ToLower(strings.TrimRight(r.URL.Path, "/")) {
	case "/hub/security/v1/state",
		"/hub/security/v1/mfa/verify",
		"/hub/security/v1/mfa/recovery":
		return true
	case "/hub/security/v1/mfa/enroll",
		"/hub/security/v1/mfa/enroll/confirm":
		return !mfaEnrolled
	default:
		return false
	}
}

func isProvisioningPath(path string) bool {
	switch strings.ToLower(strings.TrimRight(path, "/")) {
	case "/hub/user/v1/users/me",
		"/hub/user/v1/users/password",
		"/common/auth/v1/verify-email",
		"/common/auth/v1/resend-verification":
		return true
	default:
		return false
	}
}
