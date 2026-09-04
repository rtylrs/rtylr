package application

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/user/internal/domain"
	"github.com/voxire/onion/internal/testutil"
)

func TestCreateBusinessForUserCreatesOwnerAuthorizationMembership(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	service := NewService(ctx, db, "")

	email := fmt.Sprintf("business-owner+%d@example.com", time.Now().UnixNano())
	userID, err := service.repo.CreateUser(ctx, &domain.User{
		Email:        email,
		PasswordHash: "hash",
		FullName:     "Business Owner",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	businessName := fmt.Sprintf("Owner Authorization Business %d", time.Now().UnixNano())
	var businessID int64
	t.Cleanup(func() {
		cleanupCtx := context.Background()
		_, _ = db.ExecContext(cleanupCtx, `
			DELETE FROM authorization_business_members
			WHERE business_id = ? AND user_id = ?`, businessID, userID)
		_, _ = db.ExecContext(cleanupCtx, `
			DELETE FROM hr_employees
			WHERE business_id = ? AND user_id = ?`, businessID, userID)
		_, _ = db.ExecContext(cleanupCtx, "DELETE FROM businesses WHERE id = ?", businessID)
		_ = service.repo.DeleteUser(cleanupCtx, userID)
	})

	if err := service.CreateBusinessForUser(ctx, userID, businessName); err != nil {
		t.Fatalf("CreateBusinessForUser: %v", err)
	}

	if err := db.QueryRowContext(ctx, "SELECT business_id FROM users WHERE id = ?", userID).Scan(&businessID); err != nil {
		t.Fatalf("read created business ID: %v", err)
	}
	if businessID <= 0 {
		t.Fatalf("created business ID = %d, want a positive ID", businessID)
	}

	var roleKey, status string
	var roleID sql.NullInt64
	err = db.QueryRowContext(ctx, `
		SELECT dash_role_key, status, dash_role_id
		FROM authorization_business_members
		WHERE business_id = ? AND user_id = ?`, businessID, userID).Scan(&roleKey, &status, &roleID)
	if err != nil {
		t.Fatalf("read owner authorization membership: %v", err)
	}
	if roleKey != "owner" {
		t.Fatalf("owner dash role key = %q, want owner", roleKey)
	}
	if status != "active" {
		t.Fatalf("owner membership status = %q, want active", status)
	}
	if !roleID.Valid || roleID.Int64 <= 0 {
		t.Fatalf("owner dash role ID = %v, want a positive ID", roleID)
	}
}
