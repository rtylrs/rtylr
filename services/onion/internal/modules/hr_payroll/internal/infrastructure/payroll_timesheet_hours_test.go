package infrastructure_test

import (
	"context"
	"database/sql"
	"testing"

	payrollapp "github.com/voxire/onion/internal/modules/hr_payroll/internal/application"
	payrollinfra "github.com/voxire/onion/internal/modules/hr_payroll/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

// This file is the payroll half of the B11-R05 regression: given the rows
// hr_time produces (a 'locked' timesheet - the status LockTimesheet writes when
// the final approval level passes - and its hr_time_entries minutes), the
// summary and TriggerPayrollRun must report non-zero hours and hourly base pay.
//
// hr_time's own DB test (TestApproveTimesheet_FinalLevelLocksTimesheet in
// internal/modules/hr_time/internal/application) drives the real submit ->
// approve lifecycle and proves that end state; the static contract test in
// payroll_timesheet_status_alignment_test.go proves the status set matches.
// The module boundary prevents one test package from importing both modules'
// internal trees, so the seed here writes the same status directly.

// cleanupB11R05Fixture removes every row a payroll-hours regression test writes,
// children first, so a failed run cannot poison the next one.
func cleanupB11R05Fixture(t *testing.T, db *sql.DB, businessID int64) {
	t.Helper()
	t.Cleanup(func() {
		ctx := context.Background()
		for _, stmt := range []string{
			"DELETE FROM hr_payroll_entries WHERE business_id = ?",
			"DELETE FROM hr_payroll_runs WHERE business_id = ?",
			"DELETE FROM hr_payroll_config WHERE business_id = ?",
			"DELETE FROM hr_timesheet_approval_actions WHERE business_id = ?",
			"DELETE FROM hr_timesheets WHERE business_id = ?",
			"DELETE FROM hr_time_entries WHERE business_id = ?",
			"DELETE FROM hr_employees WHERE business_id = ?",
		} {
			if _, err := db.ExecContext(ctx, stmt, businessID); err != nil {
				t.Logf("cleanup %q for business %d: %v", stmt, businessID, err)
			}
		}
	})
}

func insertB11R05Employee(t *testing.T, db *sql.DB, businessID int64, email, hourlyRate string) int64 {
	t.Helper()
	res, err := db.ExecContext(context.Background(), `
		INSERT INTO hr_employees (owner_id, business_id, first_name, last_name, email, status, hourly_rate)
		VALUES (?, ?, 'B11R05', 'Hourly', ?, 'active', ?)`,
		businessID, businessID, email, hourlyRate)
	if err != nil {
		t.Fatalf("insert hourly employee: %v", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("hourly employee id: %v", err)
	}
	return id
}

func insertB11R05PayrollConfig(t *testing.T, db *sql.DB, businessID int64) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(), `
		INSERT INTO hr_payroll_config (owner_id, business_id, period_type, pay_day_of_month, overtime_enabled)
		VALUES (?, ?, 'monthly', 28, FALSE)`,
		businessID, businessID); err != nil {
		t.Fatalf("insert payroll config: %v", err)
	}
}

// insertB11R05Timesheet writes the timesheet row exactly as hr_time's
// CreateTimesheet does, with total_minutes left at its unmaintained default.
func insertB11R05Timesheet(t *testing.T, db *sql.DB, businessID, employeeID int64, periodStart, periodEnd, status string) int64 {
	t.Helper()
	res, err := db.ExecContext(context.Background(), `
		INSERT INTO hr_timesheets (owner_id, business_id, employee_id, period_start, period_end, status, current_approval_level, total_minutes)
		VALUES (?, ?, ?, ?, ?, ?, 0, 0)`,
		businessID, businessID, employeeID, periodStart, periodEnd, status)
	if err != nil {
		t.Fatalf("insert timesheet (%s): %v", status, err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("timesheet id: %v", err)
	}
	return id
}

func insertB11R05Entry(t *testing.T, db *sql.DB, businessID, employeeID, timesheetID int64, date string, minutes int32) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(), `
		INSERT INTO hr_time_entries (owner_id, business_id, employee_id, timesheet_id, date, duration_minutes, is_billable, status, source)
		VALUES (?, ?, ?, ?, ?, ?, FALSE, 'draft', 'manual')`,
		businessID, businessID, employeeID, timesheetID, date, minutes); err != nil {
		t.Fatalf("insert time entry: %v", err)
	}
}

// TestTriggerPayrollRun_HourlyBasePayFromLockedTimesheet is the cross-module
// regression test for B11-R05: a finally approved timesheet (hr_time's 'locked'
// state) with 12 hours in hr_time_entries must pay hourly base pay. Before the
// fix the summary query filtered status='approved' (a status hr_time never
// writes) and read the never-written hr_timesheets.total_minutes, so this
// assertion saw HoursWorked=0 and BasePayCents=0 while still paying only
// overtime premiums.
func TestTriggerPayrollRun_HourlyBasePayFromLockedTimesheet(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	const businessID int64 = 9100761
	const approverID int64 = 9100762
	const periodStart = "2026-09-01"
	const periodEnd = "2026-09-30"

	cleanupB11R05Fixture(t, db, businessID)

	employeeID := insertB11R05Employee(t, db, businessID, "b11r05-hourly@example.com", "30.00")
	insertB11R05PayrollConfig(t, db, businessID)

	timesheetID := insertB11R05Timesheet(t, db, businessID, employeeID, periodStart, periodEnd, "locked")
	insertB11R05Entry(t, db, businessID, employeeID, timesheetID, "2026-09-02", 480)
	insertB11R05Entry(t, db, businessID, employeeID, timesheetID, "2026-09-03", 240)

	// The summary itself must report the entry minutes, not the dead column.
	payrollRepo := payrollinfra.NewRepository(db)
	summaries, err := payrollRepo.ListApprovedTimesheetSummaries(ctx, businessID, periodStart, periodEnd)
	if err != nil {
		t.Fatalf("ListApprovedTimesheetSummaries: %v", err)
	}
	var summaryMinutes int32
	for _, s := range summaries {
		if s.EmployeeID == employeeID {
			summaryMinutes = s.TotalMinutes
		}
	}
	if summaryMinutes != 720 {
		t.Fatalf("summary minutes = %d, want 720 (from hr_time_entries, not hr_timesheets.total_minutes)", summaryMinutes)
	}

	// And a full run must store non-zero hours and hourly base pay.
	payrollService := payrollapp.NewService(ctx, db, nil, nil)
	run, entries, err := payrollService.TriggerPayrollRun(ctx, businessID, approverID, periodStart, periodEnd)
	if err != nil {
		t.Fatalf("TriggerPayrollRun: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("payroll entries = %d, want 1 (only the test business's employee)", len(entries))
	}
	entry := entries[0]
	if entry.HoursWorked != 12 {
		t.Errorf("HoursWorked = %v, want 12", entry.HoursWorked)
	}
	if entry.BasePayCents != 36000 {
		t.Errorf("BasePayCents = %d, want 36000 (12h * $30.00)", entry.BasePayCents)
	}
	if run.TotalGrossCents != 36000 || run.TotalNetCents != 36000 {
		t.Errorf("run totals = gross %d net %d, want 36000/36000", run.TotalGrossCents, run.TotalNetCents)
	}
}

// TestListApprovedTimesheetSummaries_StatusAndMinutesSource pins the two halves
// of the summary query without going through the payroll service: final
// approved ('locked') and legacy ('approved') timesheets contribute the sum of
// their time entries, while an intermediate ('submitted') timesheet is not paid
// even if it has entries.
func TestListApprovedTimesheetSummaries_StatusAndMinutesSource(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	const businessID int64 = 9100763
	const employeeID int64 = 4101

	cleanupB11R05Fixture(t, db, businessID)

	lockedSheet := insertB11R05Timesheet(t, db, businessID, employeeID, "2026-09-01", "2026-09-30", "locked")
	insertB11R05Entry(t, db, businessID, employeeID, lockedSheet, "2026-09-02", 90)
	insertB11R05Entry(t, db, businessID, employeeID, lockedSheet, "2026-09-03", 30)

	legacySheet := insertB11R05Timesheet(t, db, businessID, employeeID, "2026-08-01", "2026-08-31", "approved")
	insertB11R05Entry(t, db, businessID, employeeID, legacySheet, "2026-08-05", 60)

	submittedSheet := insertB11R05Timesheet(t, db, businessID, employeeID, "2026-07-01", "2026-07-31", "submitted")
	insertB11R05Entry(t, db, businessID, employeeID, submittedSheet, "2026-07-05", 600)

	payrollRepo := payrollinfra.NewRepository(db)
	summaries, err := payrollRepo.ListApprovedTimesheetSummaries(ctx, businessID, "2026-07-01", "2026-09-30")
	if err != nil {
		t.Fatalf("ListApprovedTimesheetSummaries: %v", err)
	}
	if len(summaries) != 1 {
		t.Fatalf("summaries = %d, want 1", len(summaries))
	}
	if summaries[0].EmployeeID != employeeID {
		t.Errorf("employee_id = %d, want %d", summaries[0].EmployeeID, employeeID)
	}
	// locked 120 + legacy approved 60; the submitted sheet's 600 minutes are
	// not approved time and must not be paid.
	if summaries[0].TotalMinutes != 180 {
		t.Errorf("total_minutes = %d, want 180 (locked 120 + legacy approved 60; submitted 600 excluded)", summaries[0].TotalMinutes)
	}
}

// A run that covers only part of a timesheet period keeps the query's
// long-standing day-count proration, now applied to real entry minutes.
func TestListApprovedTimesheetSummaries_ProratesPartialPeriodOverlap(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	const businessID int64 = 9100764

	cleanupB11R05Fixture(t, db, businessID)

	timesheetID := insertB11R05Timesheet(t, db, businessID, 4101, "2026-09-01", "2026-09-30", "locked")
	insertB11R05Entry(t, db, businessID, 4101, timesheetID, "2026-09-02", 310)

	payrollRepo := payrollinfra.NewRepository(db)
	// 15 of September's 30 days: 310 * 15/30 = 155.
	summaries, err := payrollRepo.ListApprovedTimesheetSummaries(ctx, businessID, "2026-09-16", "2026-09-30")
	if err != nil {
		t.Fatalf("ListApprovedTimesheetSummaries: %v", err)
	}
	if len(summaries) != 1 {
		t.Fatalf("summaries = %d, want 1", len(summaries))
	}
	if summaries[0].TotalMinutes != 155 {
		t.Errorf("total_minutes = %d, want 155 (310 entry minutes prorated over the 15/30 overlap)", summaries[0].TotalMinutes)
	}
}
