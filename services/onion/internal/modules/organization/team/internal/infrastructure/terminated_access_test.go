package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/organization/team/internal/application"
	"github.com/voxire/onion/internal/modules/organization/team/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

type terminatedEmployeeSourceStub []application.TerminatedEmployee

func (s terminatedEmployeeSourceStub) ListTerminatedEmployees(context.Context) ([]application.TerminatedEmployee, error) {
	return s, nil
}

// These tests follow the suite's opt-in integration pattern: they need
// DB_HOST plus the fixture user 1002 (active permissions, active employee)
// that the rest of this package already relies on. The employee fixture is
// flipped to terminated for the sweep and restored on cleanup so the shared
// fixture stays intact for other tests.
func TestTerminatedAccessReconciliation_DeactivatesTerminatedOnly(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	service := application.NewService(ctx, infrastructure.NewRepository(db)).WithTerminatedEmployeeSource(
		terminatedEmployeeSourceStub{{BusinessID: testutil.BusinessID, UserID: teamStaffUserID}},
	)

	// Fixture precondition: staff 1002 has active business membership.
	var membershipStatus string
	if err := db.QueryRowContext(ctx,
		"SELECT status FROM authorization_business_members WHERE business_id = ? AND user_id = ?",
		testutil.BusinessID, teamStaffUserID,
	).Scan(&membershipStatus); err != nil {
		t.Fatalf("fixture membership missing: %v", err)
	}
	if membershipStatus != "active" {
		t.Fatalf("fixture precondition failed: staff %d membership status is %q", teamStaffUserID, membershipStatus)
	}

	// Mark the fixture employee terminated, sweep, and expect membership to be
	// revoked. Restore the shared fixture afterwards.
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
		if _, err := db.ExecContext(ctx, `
			UPDATE authorization_business_members SET status = 'active'
			WHERE business_id = ? AND user_id = ?`, testutil.BusinessID, teamStaffUserID); err != nil {
			t.Errorf("restore membership: %v", err)
		}
		for _, table := range []string{"authorization_license_assignments", "authorization_role_assignments"} {
			if _, err := db.ExecContext(ctx, "UPDATE "+table+" SET status = 'active', revoked_at = NULL, revoked_by = NULL WHERE business_id = ? AND user_id = ?", testutil.BusinessID, teamStaffUserID); err != nil {
				t.Errorf("restore %s: %v", table, err)
			}
		}
	})

	if err := service.RunTerminatedAccessReconciliation(ctx); err != nil {
		t.Fatalf("RunTerminatedAccessReconciliation: %v", err)
	}

	var after string
	if err := db.QueryRowContext(ctx,
		"SELECT status FROM authorization_business_members WHERE business_id = ? AND user_id = ?",
		testutil.BusinessID, teamStaffUserID,
	).Scan(&after); err != nil {
		t.Fatalf("read membership after sweep: %v", err)
	}
	if after != "revoked" {
		t.Fatalf("membership status after reconciliation = %q, want revoked", after)
	}

}
