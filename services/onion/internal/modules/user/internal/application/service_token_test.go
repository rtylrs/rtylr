package application

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/user/internal/domain"
	"github.com/voxire/onion/internal/platform/utils/jwtutils"
	"github.com/voxire/onion/internal/testutil"
)

func TestValidateToken_HydratesMissingBusinessScope(t *testing.T) {
	db := testutil.OpenDB(t)

	ctx := context.Background()
	ctx = context.WithValue(ctx, jwtutils.JWTSecretKey, "test-secret")
	ctx = context.WithValue(ctx, jwtutils.JWTDurationKey, time.Hour)

	service := NewService(ctx, db, "")
	if !service.Initialize(ctx) {
		t.Fatal("failed to initialize user service jwt helper")
	}

	email := fmt.Sprintf("token-scope+%d@example.com", time.Now().UnixNano())
	userID, err := service.repo.CreateUser(ctx, &domain.User{
		Email:        email,
		PasswordHash: "hash",
		FullName:     "Token Scope",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	businessID, err := service.repo.CreateBusiness(ctx, "Token Scope Business", userID)
	if err != nil {
		t.Fatalf("CreateBusiness: %v", err)
	}
	if err := service.repo.SetUserBusinessID(ctx, userID, businessID); err != nil {
		t.Fatalf("SetUserBusinessID: %v", err)
	}

	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "UPDATE users SET business_id = NULL WHERE id = ?", userID)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM businesses WHERE id = ?", businessID)
		_ = service.repo.DeleteUser(context.Background(), userID)
	})

	token, err := service.JWT.CreateJWT(&AccessTokenClaims{
		UserID:         userID,
		Platform:       "rtylr",
		PlatformUserID: 42,
		PlanID:         9,
	})
	if err != nil {
		t.Fatalf("CreateJWT: %v", err)
	}

	claims, err := service.ValidateToken(ctx, token)
	if err != nil {
		t.Fatalf("ValidateToken: %v", err)
	}
	if claims.BusinessID != businessID {
		t.Fatalf("expected hydrated business %d, got %d", businessID, claims.BusinessID)
	}
	if claims.Platform != "rtylr" || claims.PlatformUserID != 42 || claims.PlanID != 9 {
		t.Fatalf("platform claims were not preserved: %+v", claims)
	}

	refreshedToken, err := service.RefreshToken(ctx, userID)
	if err != nil {
		t.Fatalf("RefreshToken: %v", err)
	}
	refreshedClaims := &AccessTokenClaims{}
	if err := service.JWT.ValidateToken(ctx, refreshedToken, refreshedClaims); err != nil {
		t.Fatalf("raw refreshed token ValidateToken: %v", err)
	}
	if refreshedClaims.BusinessID != businessID {
		t.Fatalf("expected refreshed token business %d, got %d", businessID, refreshedClaims.BusinessID)
	}
	if refreshedClaims.UserID != userID || refreshedClaims.BusinessID != businessID {
		t.Fatalf("expected refreshed token identity, got user=%d business=%d", refreshedClaims.UserID, refreshedClaims.BusinessID)
	}
}
