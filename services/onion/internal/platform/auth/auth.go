package auth

import (
	"context"
	"time"
)

type ContextKey string

const (
	AuthUserKey ContextKey = "auth_user"
)

//nolint:revive // stutter is preferred over changing struct name across hundreds of files
type AuthUser struct {
	UserID         int64  `json:"user_id"`
	BusinessID     int64  `json:"business_id"`
	Role           string `json:"role"`
	Platform       string `json:"platform,omitempty"`
	PlatformUserID int64  `json:"platform_user_id,omitempty"`
	PlanID         int64  `json:"plan_id,omitempty"`
	Token          string `json:"token"`
	IP             string `json:"ip"`
}

type TokenClaims struct {
	UserID         int64
	BusinessID     int64
	Role           string
	Platform       string
	PlatformUserID int64
	PlanID         int64
	ExpiresAt      time.Time
}

type TokenValidator interface {
	ValidateToken(ctx context.Context, token string) (int64, error)
	ValidateTokenFull(ctx context.Context, token string) (*TokenClaims, error)
}

func GetAuthUser(ctx context.Context) (*AuthUser, bool) {
	value := ctx.Value(AuthUserKey)
	if value == nil {
		return nil, false
	}

	authUser, ok := value.(*AuthUser)

	return authUser, ok
}

func IsAuthenticated(ctx context.Context) bool {
	_, ok := GetAuthUser(ctx)

	return ok
}

func CopyRequestValues(base, request context.Context) context.Context {
	if v := request.Value(AuthUserKey); v != nil {
		base = context.WithValue(base, AuthUserKey, v)
	}

	if v := request.Value(deviceCtxKey{}); v != nil {
		base = context.WithValue(base, deviceCtxKey{}, v)
	}

	if v := request.Value(productPlatformCtxKey{}); v != nil {
		base = context.WithValue(base, productPlatformCtxKey{}, v)
	}

	return base
}
