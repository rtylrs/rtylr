package infrastructure_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/subscriptions/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

// seedSeatOwner creates a dedicated owner+business pair for seat counting.
// The shared fixture owner is written to by other suites running in parallel,
// which makes before/after deltas on it flaky.
func seedSeatOwner(t *testing.T, db *sql.DB) (ownerID, businessID int64) {
	t.Helper()
	ctx := context.Background()
	suffix := time.Now().UnixNano()

	res, err := db.ExecContext(ctx,
		`INSERT INTO users (email, username, password_hash) VALUES (?, ?, '')`,
		fmt.Sprintf("seat-owner-%d@test.local", suffix), fmt.Sprintf("seat-owner-%d", suffix))
	if err != nil {
		t.Fatalf("insert seat owner: %v", err)
	}
	ownerID, err = res.LastInsertId()
	if err != nil {
		t.Fatalf("seat owner id: %v", err)
	}
	businessID = ownerID
	if _, err := db.ExecContext(ctx,
		`INSERT INTO businesses (id, name, owner_id) VALUES (?, ?, ?)`,
		businessID, fmt.Sprintf("Seat Count Business %d", suffix), ownerID); err != nil {
		t.Fatalf("insert seat business: %v", err)
	}

	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM user_app_permissions WHERE owner_id = ?", ownerID)
		_, _ = db.ExecContext(ctx, "DELETE FROM businesses WHERE id = ?", businessID)
		_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", ownerID)
	})
	return ownerID, businessID
}

func insertTenantMember(t *testing.T, db *sql.DB, businessID int64, permission string) int64 {
	t.Helper()
	ctx := context.Background()

	suffix := time.Now().UnixNano()
	email := fmt.Sprintf("seat-count-%d-%s@test.local", suffix, permission)
	res, err := db.ExecContext(ctx,
		`INSERT INTO users (email, username, password_hash, business_id) VALUES (?, ?, '', ?)`,
		email, fmt.Sprintf("seat-count-%d", suffix), businessID)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	userID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("last insert id: %v", err)
	}

	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM user_app_permissions WHERE user_id = ?", userID)
		_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", userID)
	})

	if permission == "none" {
		return userID
	}

	var ownerID int64
	if err := db.QueryRowContext(ctx, `SELECT owner_id FROM businesses WHERE id = ?`, businessID).Scan(&ownerID); err != nil {
		t.Fatalf("read business owner: %v", err)
	}
	active := permission == "active"
	if _, err := db.ExecContext(ctx,
		`INSERT INTO user_app_permissions (business_id, owner_id, user_id, apps, role, is_active)
		 VALUES (?, ?, ?, JSON_ARRAY(), 'staff', ?)`,
		businessID, ownerID, userID, active); err != nil {
		t.Fatalf("insert permission: %v", err)
	}
	return userID
}

// CountActiveSeats feeds syncSubscriptionBillingState, so a deactivated member
// must stop counting even though users.business_id keeps pointing at the
// business. Before the fix the third UNION branch counted them forever.
func TestSubscriptions_CountActiveSeats_DeactivatedMemberFreesSeat(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	ownerID, businessID := seedSeatOwner(t, db)

	before, err := repo.CountActiveSeats(ctx, ownerID)
	if err != nil {
		t.Fatalf("CountActiveSeats before: %v", err)
	}
	if before != 1 {
		t.Fatalf("baseline seat count = %d, want 1 (the owner)", before)
	}

	userID := insertTenantMember(t, db, businessID, "active")

	active, err := repo.CountActiveSeats(ctx, ownerID)
	if err != nil {
		t.Fatalf("CountActiveSeats active: %v", err)
	}
	if active != before+1 {
		t.Fatalf("count with active member = %d, want %d", active, before+1)
	}

	if _, err := db.ExecContext(ctx,
		`UPDATE user_app_permissions SET is_active = FALSE WHERE owner_id = ? AND user_id = ?`,
		ownerID, userID); err != nil {
		t.Fatalf("deactivate permission: %v", err)
	}

	after, err := repo.CountActiveSeats(ctx, ownerID)
	if err != nil {
		t.Fatalf("CountActiveSeats after deactivate: %v", err)
	}
	if after != before {
		t.Fatalf("count after deactivate = %d, want %d (seat not freed)", after, before)
	}
}

// Attached users without a permission row must keep counting: only an
// explicitly inactive row removes a seat.
func TestSubscriptions_CountActiveSeats_AttachedUserWithoutPermissionRowStillCounts(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	ownerID, businessID := seedSeatOwner(t, db)

	before, err := repo.CountActiveSeats(ctx, ownerID)
	if err != nil {
		t.Fatalf("CountActiveSeats before: %v", err)
	}

	insertTenantMember(t, db, businessID, "none")

	after, err := repo.CountActiveSeats(ctx, ownerID)
	if err != nil {
		t.Fatalf("CountActiveSeats after: %v", err)
	}
	if after != before+1 {
		t.Fatalf("count with attached permission-less member = %d, want %d", after, before+1)
	}
}
