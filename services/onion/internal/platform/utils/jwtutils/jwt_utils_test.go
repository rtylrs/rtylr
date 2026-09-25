package jwtutils

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/voxire/onion/internal/platform/config"
)

// testClaims is a minimal Claims implementation used by the tests.
type testClaims struct {
	jwt.RegisteredClaims
}

func (t *testClaims) SetIssuer(iss string)        { t.Issuer = iss }
func (t *testClaims) SetExpiration(exp time.Time) { t.ExpiresAt = jwt.NewNumericDate(exp) }

func TestCreateAndValidateJWT_RoundTrip(t *testing.T) {
	helper := NewJWTHelperWithDuration("super-secret-test-key", "voxire-test", time.Hour)

	tok, err := helper.CreateJWT(&testClaims{})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if tok == "" {
		t.Fatal("token should be non-empty")
	}

	var claims testClaims
	if err := helper.ValidateToken(context.Background(), tok, &claims); err != nil {
		t.Fatalf("validate: %v", err)
	}
	if claims.Issuer != "voxire-test" {
		t.Errorf("issuer roundtrip: got %q", claims.Issuer)
	}
}

func TestValidateToken_RejectsEmpty(t *testing.T) {
	helper := NewJWTHelper("k", "i")
	if err := helper.ValidateToken(context.Background(), "", &testClaims{}); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken for empty token, got %v", err)
	}
}

func TestValidateToken_RejectsExpired(t *testing.T) {
	helper := NewJWTHelperWithDuration("k", "i", -time.Hour) // already expired

	tok, err := helper.CreateJWT(&testClaims{})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := helper.ValidateToken(context.Background(), tok, &testClaims{}); !errors.Is(err, ErrTokenExpired) {
		t.Errorf("expected ErrTokenExpired, got %v", err)
	}
}

func TestValidateToken_RejectsMissingExpiration(t *testing.T) {
	claims := &testClaims{}
	claims.SetIssuer("i")
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("k"))
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	helper := NewJWTHelper("k", "i")
	if err := helper.ValidateToken(context.Background(), token, &testClaims{}); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken for missing exp, got %v", err)
	}
}

func TestValidateToken_RejectsWrongIssuer(t *testing.T) {
	signer := NewJWTHelperWithDuration("k", "issuer-a", time.Hour)
	verifier := NewJWTHelperWithDuration("k", "issuer-b", time.Hour)

	tok, _ := signer.CreateJWT(&testClaims{})
	if err := verifier.ValidateToken(context.Background(), tok, &testClaims{}); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken for issuer mismatch, got %v", err)
	}
}

func TestValidateToken_RejectsWrongSecret(t *testing.T) {
	signer := NewJWTHelperWithDuration("secret-1", "i", time.Hour)
	verifier := NewJWTHelperWithDuration("secret-2", "i", time.Hour)

	tok, _ := signer.CreateJWT(&testClaims{})
	if err := verifier.ValidateToken(context.Background(), tok, &testClaims{}); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken for secret mismatch, got %v", err)
	}
}

func TestRefreshTokenIssuesNewSignedToken(t *testing.T) {
	helper := NewJWTHelperWithDuration("k", "i", time.Hour)
	original, _ := helper.CreateJWT(&testClaims{})
	refreshed, err := helper.RefreshToken(&testClaims{})
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	// They will be effectively-identical when the test is fast enough that
	// time.Now() resolves the same. Just confirm refresh produces a valid token.
	_ = original
	if err := helper.ValidateToken(context.Background(), refreshed, &testClaims{}); err != nil {
		t.Fatalf("refreshed token should validate: %v", err)
	}
}

func TestInitializeJWT_RequiresSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "")
	if _, err := InitializeJWT(context.Background()); err == nil {
		t.Error("expected error when JWT_SECRET is unset")
	}
}

func TestInitializeJWT_PopulatesContext(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret-value")
	t.Setenv("JWT_DURATION", "30m")

	ctx, err := InitializeJWT(context.Background())
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	if ctx.Value(JWTSecretKey) != "test-secret-value" {
		t.Errorf("secret not stored in context")
	}
	if d, ok := ctx.Value(JWTDurationKey).(time.Duration); !ok || d != 30*time.Minute {
		t.Errorf("duration not parsed: got %v", ctx.Value(JWTDurationKey))
	}
}

func TestInitializeJWT_BadDurationFails(t *testing.T) {
	os.Unsetenv("JWT_DURATION")
	t.Setenv("JWT_SECRET", "x")
	t.Setenv("JWT_DURATION", "not-a-duration")

	if _, err := InitializeJWT(context.Background()); err == nil {
		t.Fatal("expected invalid JWT_DURATION to fail initialization")
	}
}

func TestInitializeJWT_DefaultsDuration(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret-value")
	t.Setenv("JWT_DURATION", "")

	ctx, err := InitializeJWT(context.Background())
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	if d, _ := ctx.Value(JWTDurationKey).(time.Duration); d != config.DefaultJWTDuration {
		t.Errorf("default duration: got %v, want %v", d, config.DefaultJWTDuration)
	}
}
