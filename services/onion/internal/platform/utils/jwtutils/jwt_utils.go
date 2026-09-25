package jwtutils

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/voxire/onion/internal/platform/config"
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
		TokenDuration: config.DefaultJWTDuration,
	}
}

func NewJWTHelperWithDuration(secretKey, issuer string, duration time.Duration) *JWTHelper {
	if duration == 0 {
		duration = config.DefaultJWTDuration
	}
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

	_, err = jwt.ParseWithClaims(token, payload, keyVerification,
		jwt.WithIssuer(jwtHelper.Issuer), jwt.WithExpirationRequired())

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
	var jwtSecret string
	jwtDuration := config.DefaultJWTDuration

	if config.GlobalConfig != nil {
		jwtSecret = config.GlobalConfig.JWTSecret
		jwtDuration = config.GlobalConfig.JWTDuration
	} else {
		jwtSecret = os.Getenv("JWT_SECRET")
		if rawDuration := os.Getenv("JWT_DURATION"); rawDuration != "" {
			parsed, err := time.ParseDuration(rawDuration)
			if err != nil || parsed <= 0 {
				if err != nil {
					return ctx, errors.New("JWT_DURATION is invalid")
				}
				return ctx, errors.New("JWT_DURATION must be greater than zero")
			}
			jwtDuration = parsed
		}
	}
	if jwtDuration <= 0 {
		jwtDuration = config.DefaultJWTDuration
	}

	if jwtSecret == "" {
		return ctx, errors.New("JWT_SECRET environment variable is not set")
	}

	ctx = context.WithValue(ctx, JWTSecretKey, jwtSecret)
	ctx = context.WithValue(ctx, JWTDurationKey, jwtDuration)

	return ctx, nil
}

// JWTMiddleware injects the JWT secret and duration from env into every
// request context so handlers can call ctx.Value(JWTSecretKey).
func JWTMiddleware() func(http.Handler) http.Handler {
	var secret string
	duration := config.DefaultJWTDuration

	if config.GlobalConfig != nil {
		secret = config.GlobalConfig.JWTSecret
		duration = config.GlobalConfig.JWTDuration
	} else {
		secret = os.Getenv("JWT_SECRET")
		if rawDuration := os.Getenv("JWT_DURATION"); rawDuration != "" {
			if parsed, err := time.ParseDuration(rawDuration); err == nil && parsed > 0 {
				duration = parsed
			}
		}
	}
	if duration <= 0 {
		duration = config.DefaultJWTDuration
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), JWTSecretKey, secret)
			ctx = context.WithValue(ctx, JWTDurationKey, duration)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
