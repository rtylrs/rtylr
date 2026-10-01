package infrastructure_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/team/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

// seedSeatOwner creates a dedicated owner+business pair for seat counting.
// The shared fixture owner is written to by other suites running in parallel
// (invites, deactivations), which makes before/after deltas on it flaky.
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
		_, _ = db.ExecContext(ctx, "DELETE FROM authorization_business_members WHERE business_id = ?", businessID)
		_, _ = db.ExecContext(ctx, "DELETE FROM businesses WHERE id = ?", businessID)
		_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", ownerID)
	})
	return ownerID, businessID
}

// insertTenantMember creates a user and optional normalized business membership.
// Rows are removed on cleanup.
func insertTenantMember(t *testing.T, db *sql.DB, businessID int64, permission string) int64 {
	t.Helper()
	ctx := context.Background()

	suffix := time.Now().UnixNano()
	email := fmt.Sprintf("seat-member-%d-%s@test.local", suffix, permission)
	res, err := db.ExecContext(ctx,
		`INSERT INTO users (email, username, password_hash, business_id) VALUES (?, ?, '', ?)`,
		email, fmt.Sprintf("seat-member-%d", suffix), businessID)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	userID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("last insert id: %v", err)
	}

	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM authorization_business_members WHERE user_id = ?", userID)
		_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", userID)
	})

	if permission == "none" {
		return userID
	}
	status := "active"
	if permission == "inactive" {
		status = "revoked"
	}
	if _, err := db.ExecContext(ctx,
		`INSERT INTO authorization_business_members (business_id, user_id, member_type, status) VALUES (?, ?, 'member', ?)`,
		businessID, userID, status); err != nil {
		t.Fatalf("insert business membership: %v", err)
	}
	return userID
}

// A deactivated member keeps users.business_id, so CountActiveMembers must let
// the permission row decide whether they still occupy a billable seat. Before
// the fix the third UNION branch counted them forever and RemoveMember's recount
// never lowered seat_count.
func TestTeam_CountActiveMembers_DeactivatedMemberFreesSeat(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	ownerID, businessID := seedSeatOwner(t, db)

	before, err := repo.CountActiveMembers(ctx, ownerID)
	if err != nil {
		t.Fatalf("CountActiveMembers before: %v", err)
	}
	if before != 1 {
		t.Fatalf("baseline seat count = %d, want 1 (the owner)", before)
	}

	userID := insertTenantMember(t, db, businessID, "active")

	active, err := repo.CountActiveMembers(ctx, ownerID)
	if err != nil {
		t.Fatalf("CountActiveMembers active: %v", err)
	}
	if active != before+1 {
		t.Fatalf("count with active member = %d, want %d", active, before+1)
	}

	if err := repo.DeactivateUser(ctx, businessID, ownerID, ownerID, userID); err != nil {
		t.Fatalf("DeactivateUser: %v", err)
	}

	after, err := repo.CountActiveMembers(ctx, ownerID)
	if err != nil {
		t.Fatalf("CountActiveMembers after deactivate: %v", err)
	}
	if after != before {
		t.Fatalf("count after deactivate = %d, want %d (seat not freed)", after, before)
	}
}

// Users without a business-membership row do not occupy a business seat.
func TestTeam_CountActiveMembers_UnprovisionedUserDoesNotCount(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	ownerID, businessID := seedSeatOwner(t, db)

	before, err := repo.CountActiveMembers(ctx, ownerID)
	if err != nil {
		t.Fatalf("CountActiveMembers before: %v", err)
	}

	insertTenantMember(t, db, businessID, "none")

	after, err := repo.CountActiveMembers(ctx, ownerID)
	if err != nil {
		t.Fatalf("CountActiveMembers after: %v", err)
	}
	if after != before {
		t.Fatalf("count without membership = %d, want %d", after, before)
	}
}
