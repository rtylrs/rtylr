package application

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/hr_time/internal/domain"
	"github.com/voxire/onion/internal/testutil"
)

// singleLevelApprovalChain makes the first approval the final one, so the
// lifecycle reaches the level at which hr_time locks the timesheet.
type singleLevelApprovalChain struct{}

func (singleLevelApprovalChain) GetApprovalChain(context.Context, int64, string) (*domain.ApprovalChainInfo, error) {
	return &domain.ApprovalChainInfo{TotalLevels: 1}, nil
}

// TestApproveTimesheet_FinalLevelLocksTimesheet is the hr_time half of the
// B11-R05 regression. hr_payroll only pays a timesheet whose status represents
// approved time, and this pins the state the real submit -> approve lifecycle
// leaves behind: 'locked', not 'approved' (the ApproveTimesheet query that
// writes 'approved' has no caller). It also pins where the payable minutes
// live: hr_time_entries, since hr_timesheets.total_minutes is never updated
// after the insert-only CreateTimesheet, so payroll must sum the entries.
func TestApproveTimesheet_FinalLevelLocksTimesheet(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	const businessID int64 = 9100765
	const employeeID int64 = 4101
	const approverID int64 = 9100766

	t.Cleanup(func() {
		cleanupCtx := context.Background()
		for _, stmt := range []string{
			"DELETE FROM hr_timesheet_approval_actions WHERE business_id = ?",
			"DELETE FROM hr_time_entries WHERE business_id = ?",
			"DELETE FROM hr_timesheets WHERE business_id = ?",
		} {
			if _, err := db.ExecContext(cleanupCtx, stmt, businessID); err != nil {
				t.Logf("cleanup %q for business %d: %v", stmt, businessID, err)
			}
		}
	})

	service := NewService(ctx, db, singleLevelApprovalChain{})

	ts, err := service.GetOrCreateTimesheet(ctx, businessID, employeeID, "2026-10-01", "2026-10-31")
	if err != nil {
		t.Fatalf("GetOrCreateTimesheet: %v", err)
	}

	for _, minutes := range []int32{480, 240} {
		if _, err := service.CreateTimeEntry(ctx, businessID, domain.TimeEntryRecord{
			EmployeeID:      employeeID,
			TimesheetID:     ts.ID,
			Date:            "2026-10-05",
			DurationMinutes: minutes,
			Status:          domain.EntryStatusDraft,
			Source:          domain.EntrySourceManual,
		}); err != nil {
			t.Fatalf("CreateTimeEntry(%d): %v", minutes, err)
		}
	}

	submitted, err := service.SubmitTimesheet(ctx, businessID, ts.ID)
	if err != nil {
		t.Fatalf("SubmitTimesheet: %v", err)
	}
	if submitted.Status != domain.TimesheetStatusSubmitted {
		t.Fatalf("after submit status = %q, want %q", submitted.Status, domain.TimesheetStatusSubmitted)
	}

	approved, err := service.ApproveTimesheet(ctx, businessID, ts.ID, approverID)
	if err != nil {
		t.Fatalf("ApproveTimesheet: %v", err)
	}
	if approved.Status != "locked" {
		t.Fatalf("final approval status = %q, want \"locked\"; hr_payroll's summary accepts this as approved time (B11-R05)", approved.Status)
	}

	// The payable minutes exist as entry rows; the timesheet's own
	// total_minutes column is deliberately not asserted because the fix keeps
	// payroll reading hr_time_entries instead of maintaining it.
	var entryMinutes int64
	if err := db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(duration_minutes), 0)
		FROM hr_time_entries
		WHERE business_id = ? AND employee_id = ? AND timesheet_id = ?`,
		businessID, employeeID, ts.ID).Scan(&entryMinutes); err != nil {
		t.Fatalf("sum entry minutes: %v", err)
	}
	if entryMinutes != 720 {
		t.Errorf("entry minutes = %d, want 720", entryMinutes)
	}
}
