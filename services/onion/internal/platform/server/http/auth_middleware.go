package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
)

const (
	AuthHeaderKey = "Authorization"
	BearerPrefix  = "Bearer "
)

func writeUnauthorizedResponse(w http.ResponseWriter, message string, logoff bool) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)

	responseData := errors.ErrorResponse{
		Error: errors.ServerError{
			Code:    http.StatusUnauthorized,
			Message: message,
			Logoff:  &logoff,
		},
		Status: "error",
	}

	if jsonBytes, err := json.Marshal(responseData); err == nil {
		if _, writeErr := w.Write(jsonBytes); writeErr != nil {
			logger.Global().Error("failed to write unauthorized response",
				logger.Error(writeErr),
			)
		}
	}
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

		return nil, &errors.ServerError{
			Code:    500,
			Message: "Authentication service unavailable",
		}
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
				writeUnauthorizedResponse(w, "Unauthorized", false)

				return
			}

			token := extractToken(authHeader)

			var tokenValidator authPkg.TokenValidator
			if tokenValidatorResolver != nil {
				tokenValidator = tokenValidatorResolver.ResolveTokenValidator(ctx, authPkg.ProductPlatformCtx(ctx))
			}

			authUser, err := processTokenAuthentication(ctx, token, path, r.RemoteAddr, tokenValidator)
			if err != nil {
				writeUnauthorizedResponse(w, "Unauthorized", true)

				return
			}

			newCtx := context.WithValue(ctx, authPkg.AuthUserKey, authUser)
			next.ServeHTTP(w, r.WithContext(newCtx))
		})
	}
}
