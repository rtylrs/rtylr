package http

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	authPkg "github.com/voxire/onion/internal/platform/auth"
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

func GetAuthMiddlewareFunc(accessPolicy *RouteAccessPolicy, tokenValidatorResolver authPkg.TokenValidatorResolver) MiddlewareFunc {
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

			newCtx := context.WithValue(ctx, authPkg.AuthUserKey, authUser)
			next.ServeHTTP(w, r.WithContext(newCtx))
		})
	}
}
