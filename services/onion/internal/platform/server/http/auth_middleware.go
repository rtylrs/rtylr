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
					status, statusErr := statusResolver.GetProvisioningStatus(ctx, authUser.UserID)
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
							if !status.MFAEnrolled && !isMFAPath(path) {
								writeUnauthorizedResponse(ctx, w, "MFA enrollment required")
								return
							}
							if status.MFAEnrolled && !isMFAPath(path) {
								mfaResolver, ok := accessResolver.(authPkg.MFAVerifiedResolver)
								if !ok {
									writeUnauthorizedResponse(ctx, w, "MFA verification unavailable")
									return
								}
								verified, verifyErr := mfaResolver.IsMFAVerified(ctx, authUser.UserID)
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
		return authPkg.HasDashboardPermission(user, permission)
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

func isMFAPath(path string) bool {
	switch strings.ToLower(strings.TrimRight(path, "/")) {
	case "/hub/security/v1/state",
		"/hub/security/v1/2fa",
		"/hub/security/v1/mfa/enroll",
		"/hub/security/v1/mfa/enroll/confirm",
		"/hub/security/v1/mfa/verify",
		"/hub/security/v1/mfa/recovery":
		return true
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
