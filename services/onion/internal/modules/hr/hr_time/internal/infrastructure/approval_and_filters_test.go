package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/hr/hr_time/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/hr_time/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

// Timesheet filters and approval writes are covered here because both defects
// live in the SQL layer: the old queries filtered employee_id unconditionally
// (so a status-only list matched employee 0) and advanced the approval level
// with no status/level predicate.

const (
	filterTestBusinessID int64 = 9100792
	filterTestEmployeeA  int64 = 4401
	filterTestEmployeeB  int64 = 4402
)

// TestListTimesheets_StatusFilterWithoutEmployeeReturnsEveryEmployee is the
// B44-R01 regression: GET /hr/timesheets?status=submitted (what the approvals
// screen sends) has no employee_id, which used to decode to 0 and filter
// employee_id = 0 -- always an empty list, with the status filter ignored.
func TestListTimesheets_StatusFilterWithoutEmployeeReturnsEveryEmployee(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	t.Cleanup(func() {
		cleanupCtx := context.Background()
		for _, stmt := range []string{
			"DELETE FROM hr_timesheet_approval_actions WHERE business_id = ?",
			"DELETE FROM hr_timesheets WHERE business_id = ?",
		} {
			if _, err := db.ExecContext(cleanupCtx, stmt, filterTestBusinessID); err != nil {
				t.Logf("cleanup %q for business %d: %v", stmt, filterTestBusinessID, err)
			}
		}
	})

	submitted, err := repo.UpsertTimesheet(ctx, filterTestBusinessID, domain.TimesheetRecord{
		EmployeeID:  filterTestEmployeeA,
		PeriodStart: "2026-10-01",
		PeriodEnd:   "2026-10-31",
		Status:      domain.TimesheetStatusSubmitted,
	})
	if err != nil {
		t.Fatalf("UpsertTimesheet(submitted): %v", err)
	}
	if _, err := repo.UpsertTimesheet(ctx, filterTestBusinessID, domain.TimesheetRecord{
		EmployeeID:  filterTestEmployeeB,
		PeriodStart: "2026-10-01",
		PeriodEnd:   "2026-10-31",
		Status:      domain.TimesheetStatusDraft,
	}); err != nil {
		t.Fatalf("UpsertTimesheet(draft): %v", err)
	}

	sheets, total, err := repo.ListTimesheets(ctx, filterTestBusinessID, domain.ListTimesheetsParams{
		Status: ptr(domain.TimesheetStatusSubmitted),
		Limit:  50,
	})
	if err != nil {
		t.Fatalf("ListTimesheets(status=submitted): %v", err)
	}
	if total != 1 || len(sheets) != 1 {
		t.Fatalf("status-only list = %d rows / total %d, want 1/1; absent employee_id must mean all employees (B44-R01)", len(sheets), total)
	}
	if sheets[0].ID != submitted.ID || sheets[0].EmployeeID != filterTestEmployeeA {
		t.Errorf("status-only list returned timesheet %d (employee %d), want %d (employee %d)",
			sheets[0].ID, sheets[0].EmployeeID, submitted.ID, filterTestEmployeeA)
	}

	// An explicit employee filter still narrows the list, and no filters at
	// all returns both employees' rows.
	byEmployee, byEmployeeTotal, err := repo.ListTimesheets(ctx, filterTestBusinessID, domain.ListTimesheetsParams{
		EmployeeID: ptr(filterTestEmployeeB),
		Limit:      50,
	})
	if err != nil {
		t.Fatalf("ListTimesheets(employee=%d): %v", filterTestEmployeeB, err)
	}
	if byEmployeeTotal != 1 || len(byEmployee) != 1 || byEmployee[0].EmployeeID != filterTestEmployeeB {
		t.Errorf("employee filter returned %d rows / total %d, want employee %d only", len(byEmployee), byEmployeeTotal, filterTestEmployeeB)
	}

	all, allTotal, err := repo.ListTimesheets(ctx, filterTestBusinessID, domain.ListTimesheetsParams{Limit: 50})
	if err != nil {
		t.Fatalf("ListTimesheets(no filters): %v", err)
	}
	if allTotal != 2 || len(all) != 2 {
		t.Errorf("unfiltered list = %d rows / total %d, want 2/2", len(all), allTotal)
	}
}

// TestAdvanceApprovalLevel_RejectsARepeatedLevel is the B44-R02 regression: the
// increment carries the level the caller read, so a retried or overlapping
// approval that still carries the old level affects 0 rows and cannot skip a
// level of the chain.
func TestAdvanceApprovalLevel_RejectsARepeatedLevel(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const businessID int64 = 9100793
	const approverID int64 = 4403

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

	ts, err := repo.UpsertTimesheet(ctx, businessID, domain.TimesheetRecord{
		EmployeeID:  4401,
		PeriodStart: "2026-10-01",
		PeriodEnd:   "2026-10-31",
		Status:      domain.TimesheetStatusDraft,
	})
	if err != nil {
		t.Fatalf("UpsertTimesheet: %v", err)
	}
	submitted, err := repo.SubmitTimesheet(ctx, businessID, ts.ID)
	if err != nil {
		t.Fatalf("SubmitTimesheet: %v", err)
	}
	if submitted.CurrentApprovalLevel != 1 {
		t.Fatalf("submitted level = %d, want 1", submitted.CurrentApprovalLevel)
	}

	advanced, err := repo.AdvanceApprovalLevel(ctx, businessID, ts.ID, submitted.CurrentApprovalLevel, approverID)
	if err != nil {
		t.Fatalf("AdvanceApprovalLevel(first): %v", err)
	}
	if advanced.CurrentApprovalLevel != 2 {
		t.Fatalf("level after first approval = %d, want 2", advanced.CurrentApprovalLevel)
	}

	// The retry still believes the timesheet is at level 1.
	if _, err := repo.AdvanceApprovalLevel(ctx, businessID, ts.ID, submitted.CurrentApprovalLevel, approverID); !errors.Is(err, domain.ErrTimesheetNotPending) {
		t.Fatalf("retried approval error = %v, want ErrTimesheetNotPending", err)
	}

	got, err := repo.GetTimesheet(ctx, businessID, ts.ID)
	if err != nil {
		t.Fatalf("GetTimesheet: %v", err)
	}
	if got.CurrentApprovalLevel != 2 {
		t.Errorf("level after rejected retry = %d, want 2 (a level was skipped)", got.CurrentApprovalLevel)
	}

	var approvedActions int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM hr_timesheet_approval_actions
		WHERE business_id = ? AND timesheet_id = ? AND action = 'approved'`,
		businessID, ts.ID).Scan(&approvedActions); err != nil {
		t.Fatalf("count approved actions: %v", err)
	}
	if approvedActions != 1 {
		t.Errorf("approved actions = %d, want 1", approvedActions)
	}
}

// TestAdvanceApprovalLevel_RejectsUnsubmittedTimesheet pins the status half of
// the guard: a draft (level 0) must not be advanced by an approve request.
func TestAdvanceApprovalLevel_RejectsUnsubmittedTimesheet(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const businessID int64 = 9100794

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

	ts, err := repo.UpsertTimesheet(ctx, businessID, domain.TimesheetRecord{
		EmployeeID:  4401,
		PeriodStart: "2026-10-01",
		PeriodEnd:   "2026-10-31",
		Status:      domain.TimesheetStatusDraft,
	})
	if err != nil {
		t.Fatalf("UpsertTimesheet: %v", err)
	}

	if _, err := repo.AdvanceApprovalLevel(ctx, businessID, ts.ID, 0, 4403); !errors.Is(err, domain.ErrTimesheetNotPending) {
		t.Fatalf("advance of a draft error = %v, want ErrTimesheetNotPending", err)
	}

	got, err := repo.GetTimesheet(ctx, businessID, ts.ID)
	if err != nil {
		t.Fatalf("GetTimesheet: %v", err)
	}
	if got.Status != domain.TimesheetStatusDraft || got.CurrentApprovalLevel != 0 {
		t.Errorf("draft changed to status %q level %d; want %q/0", got.Status, got.CurrentApprovalLevel, domain.TimesheetStatusDraft)
	}
}

// TestLockTimesheetWithApprovalAction_WritesLockAndActionTogether covers the
// B08-R02 final-approval defect: the lock and its final_approved action must
// land together, and the lock must refuse a row that is not submitted.
func TestLockTimesheetWithApprovalAction_WritesLockAndActionTogether(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const businessID int64 = 9100795
	const approverID int64 = 4403

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

	ts, err := repo.UpsertTimesheet(ctx, businessID, domain.TimesheetRecord{
		EmployeeID:  4401,
		PeriodStart: "2026-10-01",
		PeriodEnd:   "2026-10-31",
		Status:      domain.TimesheetStatusDraft,
	})
	if err != nil {
		t.Fatalf("UpsertTimesheet: %v", err)
	}

	// A draft must not be finalizable, and the refused attempt must not leave
	// an action row behind.
	if _, err := repo.LockTimesheetWithApprovalAction(ctx, businessID, ts.ID, approverID, 0); !errors.Is(err, domain.ErrTimesheetNotPending) {
		t.Fatalf("final approval of a draft error = %v, want ErrTimesheetNotPending", err)
	}
	stillDraft, err := repo.GetTimesheet(ctx, businessID, ts.ID)
	if err != nil {
		t.Fatalf("GetTimesheet after refused final approval: %v", err)
	}
	if stillDraft.Status != domain.TimesheetStatusDraft {
		t.Errorf("status after refused final approval = %q, want %q", stillDraft.Status, domain.TimesheetStatusDraft)
	}

	if _, err := repo.SubmitTimesheet(ctx, businessID, ts.ID); err != nil {
		t.Fatalf("SubmitTimesheet: %v", err)
	}

	locked, err := repo.LockTimesheetWithApprovalAction(ctx, businessID, ts.ID, approverID, 1)
	if err != nil {
		t.Fatalf("LockTimesheetWithApprovalAction: %v", err)
	}
	if locked.Status != "locked" {
		t.Fatalf("status after final approval = %q, want %q", locked.Status, "locked")
	}

	var finalActions int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM hr_timesheet_approval_actions
		WHERE business_id = ? AND timesheet_id = ? AND action = 'final_approved'`,
		businessID, ts.ID).Scan(&finalActions); err != nil {
		t.Fatalf("count final_approved actions: %v", err)
	}
	if finalActions != 1 {
		t.Fatalf("final_approved actions = %d, want 1 (lock and action must land together)", finalActions)
	}

	// A retried final approval is a conflict, not a second action row.
	if _, err := repo.LockTimesheetWithApprovalAction(ctx, businessID, ts.ID, approverID, 1); !errors.Is(err, domain.ErrTimesheetNotPending) {
		t.Fatalf("retried final approval error = %v, want ErrTimesheetNotPending", err)
	}
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM hr_timesheet_approval_actions
		WHERE business_id = ? AND timesheet_id = ? AND action = 'final_approved'`,
		businessID, ts.ID).Scan(&finalActions); err != nil {
		t.Fatalf("count final_approved actions after retry: %v", err)
	}
	if finalActions != 1 {
		t.Errorf("final_approved actions after retry = %d, want 1", finalActions)
	}
}
