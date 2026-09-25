package application

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// OAuthTokenClaims embeds the same fields as the user module's AccessTokenClaims
// so the existing auth middleware can validate OAuth-issued JWTs transparently.
type OAuthTokenClaims struct {
	UserID     int64  `json:"user_id"`
	BusinessID int64  `json:"business_id,omitempty"`
	Role       string `json:"role,omitempty"`
	Scope      string `json:"scope,omitempty"`
	jwt.RegisteredClaims
}

func (c *OAuthTokenClaims) SetIssuer(issuer string) {
	c.Issuer = issuer
}

func (c *OAuthTokenClaims) SetExpiration(exp time.Time) {
	if !exp.IsZero() {
		c.ExpiresAt = jwt.NewNumericDate(exp)
	}
}
