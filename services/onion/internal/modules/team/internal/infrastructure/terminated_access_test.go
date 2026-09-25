package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/team/internal/application"
	"github.com/voxire/onion/internal/testutil"
)

// These tests follow the suite's opt-in integration pattern: they need
// DB_HOST plus the fixture user 1002 (active permissions, active employee)
// that the rest of this package already relies on. The employee fixture is
// flipped to terminated for the sweep and restored on cleanup so the shared
// fixture stays intact for other tests.
func TestTerminatedAccessReconciliation_DeactivatesTerminatedOnly(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	service := application.NewService(ctx, db)

	// Fixture precondition: staff 1002 is employed and active. Assert it so a
	// drifted fixture fails loudly here rather than passing vacuously.
	var isActive int
	if err := db.QueryRowContext(ctx,
		"SELECT is_active FROM user_app_permissions WHERE owner_id = ? AND user_id = ?",
		testutil.OwnerID, teamStaffUserID,
	).Scan(&isActive); err != nil {
		t.Fatalf("fixture permission row missing: %v", err)
	}
	if isActive != 1 {
		t.Fatalf("fixture precondition failed: staff %d permission inactive", teamStaffUserID)
	}

	// Mark the fixture employee terminated, sweep, and expect the permission
	// row to flip off. Restore both rows afterwards.
	setStatus := func(status string) {
		t.Helper()
		if _, err := db.ExecContext(ctx,
			"UPDATE hr_employees SET status = ? WHERE user_id = ? AND business_id = ?",
			status, teamStaffUserID, testutil.BusinessID,
		); err != nil {
			t.Fatalf("set employee status: %v", err)
		}
	}
	setStatus("terminated")
	t.Cleanup(func() {
		setStatus("active")
		if _, err := db.ExecContext(ctx,
			"UPDATE user_app_permissions SET is_active = TRUE WHERE owner_id = ? AND user_id = ?",
			testutil.OwnerID, teamStaffUserID,
		); err != nil {
			t.Errorf("restore permission: %v", err)
		}
	})

	sweep := application.NewService(ctx, db)
	if err := sweep.RunTerminatedAccessReconciliation(ctx); err != nil {
		t.Fatalf("RunTerminatedAccessReconciliation: %v", err)
	}

	var after int
	if err := db.QueryRowContext(ctx,
		"SELECT is_active FROM user_app_permissions WHERE owner_id = ? AND user_id = ?",
		testutil.OwnerID, teamStaffUserID,
	).Scan(&after); err != nil {
		t.Fatalf("read permission after sweep: %v", err)
	}
	if after != 0 {
		t.Fatalf("permission still active after reconciliation for terminated employee")
	}

	// Scoped path: restoring employment then revoking by (business, user) must
	// be a no-op for an active employee, proving status gates the revoke.
	setStatus("active")
	if err := service.RevokeAccessForTerminated(ctx, testutil.BusinessID, teamStaffUserID); err != nil {
		t.Fatalf("RevokeAccessForTerminated on active employee: %v", err)
	}
}
