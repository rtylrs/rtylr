package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/hr_time/internal/domain"
	"github.com/voxire/onion/internal/testutil"
)

// twoLevelApprovalChain keeps the first approval below TotalLevels, which is
// the state B08-R02 needs: the old service advanced the level without ever
// asking whether the timesheet had been submitted.
type twoLevelApprovalChain struct{}

func (twoLevelApprovalChain) GetApprovalChain(context.Context, int64, string) (*domain.ApprovalChainInfo, error) {
	return &domain.ApprovalChainInfo{
		TotalLevels: 2,
		Levels: []domain.ApprovalLevelInfo{
			{LevelOrder: 1, ApproverRoleID: 1},
			{LevelOrder: 2, ApproverRoleID: 2},
		},
	}, nil
}

// TestApproveTimesheet_RejectsDraftBeforeAnyWrite is the B08-R02 service half:
// approving a never-submitted draft used to advance it (and, on a second call
// against a one-level chain, lock it), because the only guard compared
// CurrentApprovalLevel to TotalLevels.
func TestApproveTimesheet_RejectsDraftBeforeAnyWrite(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	const businessID int64 = 9100781
	const employeeID int64 = 4201
	const approverID int64 = 4202

	t.Cleanup(func() {
		cleanupCtx := context.Background()
		for _, stmt := range []string{
			"DELETE FROM hr_timesheet_approval_actions WHERE business_id = ?",
			"DELETE FROM hr_timesheets WHERE business_id = ?",
		} {
			if _, err := db.ExecContext(cleanupCtx, stmt, businessID); err != nil {
				t.Logf("cleanup %q for business %d: %v", stmt, businessID, err)
			}
		}
	})

	service := NewService(ctx, db, twoLevelApprovalChain{}, newAllowApprovalPermissions(map[int64]int64{approverID: 4203}))
	ts, err := service.GetOrCreateTimesheet(ctx, businessID, employeeID, "2026-10-01", "2026-10-31")
	if err != nil {
		t.Fatalf("GetOrCreateTimesheet: %v", err)
	}

	if _, err := service.ApproveTimesheet(ctx, businessID, ts.ID, approverID); !errors.Is(err, domain.ErrTimesheetNotPending) {
		t.Fatalf("ApproveTimesheet(draft) error = %v, want ErrTimesheetNotPending", err)
	}

	got, err := service.GetTimesheet(ctx, businessID, ts.ID)
	if err != nil {
		t.Fatalf("GetTimesheet: %v", err)
	}
	if got.Status != domain.TimesheetStatusDraft || got.CurrentApprovalLevel != 0 {
		t.Errorf("draft approve left status %q level %d, want %q/0", got.Status, got.CurrentApprovalLevel, domain.TimesheetStatusDraft)
	}

	var actions int
	if err := db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM hr_timesheet_approval_actions WHERE business_id = ? AND timesheet_id = ?",
		businessID, ts.ID,
	).Scan(&actions); err != nil {
		t.Fatalf("count approval actions: %v", err)
	}
	if actions != 0 {
		t.Errorf("draft approve recorded %d approval actions, want 0", actions)
	}
}

// TestApproveTimesheet_FinalApprovalRecordsActionWithLock asserts the
// final-approval path writes the lock and its audit row as one unit, and that
// a retry after locking is refused instead of recording a second final
// approval.
func TestApproveTimesheet_FinalApprovalRecordsActionWithLock(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	const businessID int64 = 9100782
	const employeeID int64 = 4203
	const approverID int64 = 4204

	t.Cleanup(func() {
		cleanupCtx := context.Background()
		for _, stmt := range []string{
			"DELETE FROM hr_timesheet_approval_actions WHERE business_id = ?",
			"DELETE FROM hr_timesheets WHERE business_id = ?",
		} {
			if _, err := db.ExecContext(cleanupCtx, stmt, businessID); err != nil {
				t.Logf("cleanup %q for business %d: %v", stmt, businessID, err)
			}
		}
	})

	service := NewService(ctx, db, singleLevelApprovalChain{}, newAllowApprovalPermissions(map[int64]int64{approverID: 4205}))
	ts, err := service.GetOrCreateTimesheet(ctx, businessID, employeeID, "2026-10-01", "2026-10-31")
	if err != nil {
		t.Fatalf("GetOrCreateTimesheet: %v", err)
	}
	if _, err := service.SubmitTimesheet(ctx, businessID, ts.ID); err != nil {
		t.Fatalf("SubmitTimesheet: %v", err)
	}

	approved, err := service.ApproveTimesheet(ctx, businessID, ts.ID, approverID)
	if err != nil {
		t.Fatalf("ApproveTimesheet: %v", err)
	}
	if approved.Status != "locked" {
		t.Fatalf("final approval status = %q, want \"locked\"", approved.Status)
	}

	finalActions := func() int {
		t.Helper()
		var count int
		if err := db.QueryRowContext(ctx, `
			SELECT COUNT(*) FROM hr_timesheet_approval_actions
			WHERE business_id = ? AND timesheet_id = ? AND action = 'final_approved'`,
			businessID, ts.ID).Scan(&count); err != nil {
			t.Fatalf("count final_approved actions: %v", err)
		}
		return count
	}

	if got := finalActions(); got != 1 {
		t.Fatalf("final_approved actions = %d, want 1 (lock and action must land together)", got)
	}

	if _, err := service.ApproveTimesheet(ctx, businessID, ts.ID, approverID); !errors.Is(err, domain.ErrTimesheetNotPending) {
		t.Fatalf("re-approve of a locked timesheet error = %v, want ErrTimesheetNotPending", err)
	}
	if got := finalActions(); got != 1 {
		t.Errorf("final_approved actions after retry = %d, want 1", got)
	}
}
