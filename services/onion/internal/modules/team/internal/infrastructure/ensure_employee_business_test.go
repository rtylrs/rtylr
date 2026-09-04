package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/team/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

// EnsureEmployeeForUser took the business from the invited user's own row:
//
//	SELECT ?, COALESCE(u.business_id, 0), u.id, ...
//
// users.business_id is still NULL at invite time, so it collapsed to 0 and every
// invited member's hr_employees record was scoped to no business. The dash TEAM
// widget reported 0 staff against four active members, and anything scoping HR
// employees by business missed them entirely (#1060).
//
// The business is now a parameter. This proves it is the one stored, and that
// the invitee's own row is not consulted.
//
// The three ids below are deliberately all different. scripts/test-db.sh
// installs a BEFORE INSERT trigger on every table carrying both business_id and
// owner_id which rewrites business_id to owner_id whenever business_id is 0, so
// in onion_test a row written with business_id = 0 comes back looking correct.
// Asserting a business that is neither zero nor the owner is the only way this
// assertion means anything here. It also means this test cannot demonstrate the
// old behaviour: the trigger would have repaired it. The evidence for that is
// the production data in #1060.
func TestEnsureEmployeeForUserScopesToThePassedBusiness(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	repo := infrastructure.NewRepository(db)

	const (
		userID     int64 = 990601
		ownerID    int64 = 990600
		businessID int64 = 990602
	)

	// A brand new user with users.business_id NULL, which is the invite-time
	// state the defect depended on.
	if _, err := db.ExecContext(ctx,
		"INSERT INTO users (id, email, password_hash, full_name, business_id) VALUES (?, ?, ?, ?, NULL)",
		userID, "ensure-employee-probe@test.local", "fixture-auth-disabled", "Probe Person",
	); err != nil {
		t.Skipf("cannot seed probe user: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM hr_employees WHERE user_id = ?", userID)
		_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", userID)
	})

	var nullBusiness *int64
	if err := db.QueryRowContext(ctx, "SELECT business_id FROM users WHERE id = ?", userID).Scan(&nullBusiness); err != nil {
		t.Fatalf("read seeded user: %v", err)
	}
	if nullBusiness != nil {
		t.Fatalf("probe precondition failed: users.business_id is %d, want NULL", *nullBusiness)
	}

	if err := repo.EnsureEmployeeForUser(ctx, ownerID, businessID, userID); err != nil {
		t.Fatalf("EnsureEmployeeForUser: %v", err)
	}

	var storedBusinessID, storedOwnerID int64
	if err := db.QueryRowContext(ctx,
		"SELECT business_id, owner_id FROM hr_employees WHERE user_id = ?", userID,
	).Scan(&storedBusinessID, &storedOwnerID); err != nil {
		t.Fatalf("read employee row: %v", err)
	}
	if storedBusinessID != businessID {
		t.Errorf("hr_employees.business_id = %d, want %d; the record is not scoped to the business the member was invited into",
			storedBusinessID, businessID)
	}
	if storedOwnerID != ownerID {
		t.Errorf("hr_employees.owner_id = %d, want %d", storedOwnerID, ownerID)
	}
}
