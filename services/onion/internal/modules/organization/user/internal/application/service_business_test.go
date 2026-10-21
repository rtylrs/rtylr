package application

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/organization/user/internal/domain"
	"github.com/voxire/onion/internal/testutil"
)

func TestCreateBusinessForUserResolvesOwnerFromBusiness(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	service := newTestService(db, "")

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

	var ownerID int64
	err = db.QueryRowContext(ctx, `SELECT owner_id FROM businesses WHERE id = ?`, businessID).Scan(&ownerID)
	if err != nil {
		t.Fatalf("read business owner: %v", err)
	}
	if ownerID != userID {
		t.Fatalf("business owner = %d, want %d", ownerID, userID)
	}
	var membershipCount int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM authorization_business_members
		WHERE business_id = ? AND user_id = ?`, businessID, userID).Scan(&membershipCount); err != nil {
		t.Fatalf("count owner membership rows: %v", err)
	}
	if membershipCount != 0 {
		t.Fatalf("owner has %d separate membership rows, want owner derived from business ownership", membershipCount)
	}
}
