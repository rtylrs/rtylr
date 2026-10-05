package application

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AdminClaims is the JWT payload for an authenticated admin session.
type AdminClaims struct {
	AdminID int64  `json:"admin_id"`
	Email   string `json:"email"`
	jwt.RegisteredClaims
}

func (c *AdminClaims) SetIssuer(issuer string) { c.Issuer = issuer }

func (c *AdminClaims) SetExpiration(exp time.Time) {
	if !exp.IsZero() {
		c.ExpiresAt = jwt.NewNumericDate(exp)
	}
}
