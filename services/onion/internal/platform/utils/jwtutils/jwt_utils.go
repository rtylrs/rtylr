package jwtutils

import (
	"context"
	"errors"
	"net/url"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/voxire/onion/internal/platform/logger"
)

var (
	ErrInvalidToken = errors.New("token is invalid")
	ErrTokenExpired = errors.New("token has expired")
)

type Claims interface {
	jwt.Claims
	SetIssuer(string)
	SetExpiration(time.Time)
}

type JWTHelper struct {
	SecretKey     string
	Issuer        string
	TokenDuration time.Duration
}

func NewJWTHelper(secretKey, issuer string) *JWTHelper {
	return &JWTHelper{
		SecretKey:     secretKey,
		Issuer:        issuer,
		TokenDuration: 0,
	}
}

func NewJWTHelperWithDuration(secretKey, issuer string, duration time.Duration) *JWTHelper {
	return &JWTHelper{
		SecretKey:     secretKey,
		Issuer:        issuer,
		TokenDuration: duration,
	}
}

func (jwtHelper *JWTHelper) ValidateToken(ctx context.Context, token string, payload jwt.Claims) error {
	if token == "" {
		return ErrInvalidToken
	}

	token, err := url.PathUnescape(token)
	if err != nil {
		logger.Ctx(ctx).Warn("Could not urldecode token")

		return ErrInvalidToken
	}

	keyVerification := func(t *jwt.Token) (interface{}, error) {
		_, ok := t.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, ErrInvalidToken
		}

		return []byte(jwtHelper.SecretKey), nil
	}

	_, err = jwt.ParseWithClaims(token, payload, keyVerification, jwt.WithIssuer(jwtHelper.Issuer))

	if errors.Is(err, jwt.ErrTokenExpired) {
		logger.Ctx(ctx).Warn("Token expired")

		return ErrTokenExpired
	}

	if err != nil {
		logger.Ctx(ctx).Warn("Token is invalid",
			logger.String("error", err.Error()),
		)

		return ErrInvalidToken
	}

	return nil
}

func (jwtHelper *JWTHelper) CreateJWT(claims Claims) (string, error) {
	claims.SetIssuer(jwtHelper.Issuer)

	if jwtHelper.TokenDuration > 0 {
		claims.SetExpiration(time.Now().Add(jwtHelper.TokenDuration))
	} else {
		claims.SetExpiration(time.Time{})
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(jwtHelper.SecretKey))
}

func (jwtHelper *JWTHelper) RefreshToken(claims Claims) (string, error) {
	claims.SetIssuer(jwtHelper.Issuer)

	if jwtHelper.TokenDuration > 0 {
		claims.SetExpiration(time.Now().Add(jwtHelper.TokenDuration))
	} else {
		claims.SetExpiration(time.Time{})
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(jwtHelper.SecretKey))
}

type CtxKey string

const (
	JWTSecretKey   CtxKey = "jwt_secret"
	JWTDurationKey CtxKey = "jwt_duration"
)

func InitializeJWT(ctx context.Context) (context.Context, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return ctx, errors.New("JWT_SECRET environment variable is not set")
	}

	jwtDuration := time.Duration(0)

	if durationStr := os.Getenv("JWT_DURATION"); durationStr != "" {
		if duration, err := time.ParseDuration(durationStr); err == nil {
			jwtDuration = duration
		}
	}

	ctx = context.WithValue(ctx, JWTSecretKey, jwtSecret)
	ctx = context.WithValue(ctx, JWTDurationKey, jwtDuration)

	return ctx, nil
}
