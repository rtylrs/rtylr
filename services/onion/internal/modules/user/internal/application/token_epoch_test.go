package application

import (
	"context"
	"errors"
	"fmt"
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
