package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/voxire/onion/internal/platform/utils/jwtutils"
)

const webSocketJWTIssuer = "onion-platform"

var (
	ErrWebSocketTokenMissing = errors.New("websocket token missing")
	ErrJWTSecretMissing      = errors.New("jwt secret missing")
	ErrWebSocketUserMissing  = errors.New("websocket token missing user")
)

type webSocketClaims struct {
	UserID         int64  `json:"user_id"`
	BusinessID     int64  `json:"business_id"`
	IsConsoleAdmin bool   `json:"is_console_admin,omitempty"`
	Platform       string `json:"platform,omitempty"`
	PlatformUserID int64  `json:"platform_user_id,omitempty"`
	PlanID         int64  `json:"plan_id,omitempty"`
	jwt.RegisteredClaims
}

func (c *webSocketClaims) SetIssuer(issuer string) {
	c.Issuer = issuer
}

func (c *webSocketClaims) SetExpiration(exp time.Time) {
	if !exp.IsZero() {
		c.ExpiresAt = jwt.NewNumericDate(exp)
	}
}

func WithAuthUser(ctx context.Context, user *AuthUser) context.Context {
	return context.WithValue(ctx, AuthUserKey, user)
}

func WebSocketToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	}
	return ""
}

//nolint:revive // The exported name is explicit at call sites that authenticate WS upgrades.
func AuthUserFromWebSocketRequest(r *http.Request) (*AuthUser, error) {
	if ticket := r.URL.Query().Get("ticket"); ticket != "" {
		user, ok := ConsumeWebSocketTicket(r.Context(), ticket)
		if !ok {
			return nil, ErrWebSocketTicketInvalid
		}
		user.IP = r.RemoteAddr
		return user, nil
	}

	token := WebSocketToken(r)
	if token == "" {
		return nil, ErrWebSocketTokenMissing
	}

	jwtSecret, _ := r.Context().Value(jwtutils.JWTSecretKey).(string)
	if jwtSecret == "" {
		return nil, ErrJWTSecretMissing
	}

	var claims webSocketClaims
	helper := jwtutils.NewJWTHelper(jwtSecret, webSocketJWTIssuer)
	if err := helper.ValidateToken(r.Context(), token, &claims); err != nil {
		return nil, err
	}
	if claims.UserID == 0 {
		return nil, ErrWebSocketUserMissing
	}

	return &AuthUser{
		UserID:         claims.UserID,
		BusinessID:     claims.BusinessID,
		IsConsoleAdmin: claims.IsConsoleAdmin,
		Platform:       claims.Platform,
		PlatformUserID: claims.PlatformUserID,
		PlanID:         claims.PlanID,
		Token:          token,
		IP:             r.RemoteAddr,
	}, nil
}
