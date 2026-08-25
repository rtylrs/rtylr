package application

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/user/internal/domain"
	"github.com/voxire/onion/internal/platform/utils/jwtutils"
	"github.com/voxire/onion/internal/testutil"
)

// TestValidateToken_RejectsTokenFromBeforePasswordChange covers the epoch
// revocation mechanism (#762): a token issued before a password change stops
// validating once the stored epoch moves past the one embedded in the token.
func TestValidateToken_RejectsTokenFromBeforePasswordChange(t *testing.T) {
	db := testutil.OpenDB(t)

	ctx := context.Background()
	ctx = context.WithValue(ctx, jwtutils.JWTSecretKey, "test-secret")
	ctx = context.WithValue(ctx, jwtutils.JWTDurationKey, time.Hour)

	service := NewService(ctx, db, "")
	if !service.Initialize(ctx) {
		t.Fatal("failed to initialize user service jwt helper")
	}

	email := fmt.Sprintf("token-epoch+%d@example.com", time.Now().UnixNano())
	userID, err := service.repo.CreateUser(ctx, &domain.User{
		Email:        email,
		PasswordHash: "hash",
		FullName:     "Token Epoch",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	t.Cleanup(func() {
		_ = service.repo.DeleteUser(context.Background(), userID)
	})

	token, err := service.GenerateToken(ctx, userID)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	if _, err := service.ValidateToken(ctx, token); err != nil {
		t.Fatalf("token should validate before any bump: %v", err)
	}

	if err := service.repo.UpdateUserPassword(ctx, userID, "new-hash"); err != nil {
		t.Fatalf("UpdateUserPassword: %v", err)
	}

	if _, err := service.ValidateToken(ctx, token); !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("token issued before password change: got err=%v, want ErrInvalidCredentials", err)
	}
}

// TestHydrateAccessTokenClaims_ClearsStaleConsoleAdmin covers issue #762
// item 3: a demoted console admin's outstanding claim must be cleared during
// validation instead of being carried to token expiry.
func TestHydrateAccessTokenClaims_ClearsStaleConsoleAdmin(t *testing.T) {
	db := testutil.OpenDB(t)

	ctx := context.Background()
	ctx = context.WithValue(ctx, jwtutils.JWTSecretKey, "test-secret")
	ctx = context.WithValue(ctx, jwtutils.JWTDurationKey, time.Hour)

	service := NewService(ctx, db, "")
	if !service.Initialize(ctx) {
		t.Fatal("failed to initialize user service jwt helper")
	}

	email := fmt.Sprintf("stale-admin+%d@example.com", time.Now().UnixNano())
	userID, err := service.repo.CreateUser(ctx, &domain.User{
		Email:        email,
		PasswordHash: "hash",
		FullName:     "Stale Admin",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	t.Cleanup(func() {
		_ = service.repo.DeleteUser(context.Background(), userID)
	})

	claims := &AccessTokenClaims{UserID: userID, Role: "console_admin"}
	service.hydrateAccessTokenClaims(ctx, claims)
	if claims.Role == "console_admin" {
		t.Fatal("stale console_admin role must be cleared for non-admin users")
	}
	if claims.TokenEpoch != 0 {
		t.Errorf("TokenEpoch = %d, want 0 for untouched user", claims.TokenEpoch)
	}
}

// TestLinkGoogleUser_RejectsUnverifiedEmail covers issue #815: an OAuth
// identity whose provider has not verified the email must never link to (or
// provision) a local account, even when a same-email local user exists.
func TestLinkGoogleUser_RejectsUnverifiedEmail(t *testing.T) {
	db := testutil.OpenDB(t)

	ctx := context.Background()
	ctx = context.WithValue(ctx, jwtutils.JWTSecretKey, "test-secret")
	ctx = context.WithValue(ctx, jwtutils.JWTDurationKey, time.Hour)

	service := NewService(ctx, db, "")
	if !service.Initialize(ctx) {
		t.Fatal("failed to initialize user service jwt helper")
	}

	victimEmail := fmt.Sprintf("oauth-victim+%d@example.com", time.Now().UnixNano())
	victimID, err := service.repo.CreateUser(ctx, &domain.User{
		Email:        victimEmail,
		PasswordHash: "hash",
		FullName:     "Victim",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	t.Cleanup(func() {
		_ = service.repo.DeleteUser(context.Background(), victimID)
	})

	result, err := service.linkGoogleUser(ctx, &GoogleUser{
		Email:         victimEmail,
		Name:          "Attacker",
		EmailVerified: false,
	}, "", netip.MustParseAddr("203.0.113.9"), "ua")
	if !errors.Is(err, ErrGoogleEmailUnverified) {
		t.Fatalf("err = %v, want ErrGoogleEmailUnverified", err)
	}
	if result != nil {
		t.Fatalf("result = %+v, want nil", result)
	}

	// The verified path still links the existing local account.
	result, err = service.linkGoogleUser(ctx, &GoogleUser{
		Email:         victimEmail,
		Name:          "Victim",
		EmailVerified: true,
	}, "", netip.MustParseAddr("203.0.113.9"), "ua")
	if err != nil {
		t.Fatalf("verified link failed: %v", err)
	}
	if result.IsNew || result.UserID != victimID {
		t.Errorf("linked result = %+v, want existing user %d", result, victimID)
	}
}
