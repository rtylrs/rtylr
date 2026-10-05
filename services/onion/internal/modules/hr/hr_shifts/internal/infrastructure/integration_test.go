package infrastructure_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/hr/hr_shifts/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/hr_shifts/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

// deleteShiftRow removes a row directly, so a test cleans up even if the code
// path under test is the thing that is broken.
func deleteShiftRow(t *testing.T, db *sql.DB, id int64) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(),
		"DELETE FROM hr_shifts WHERE id = ? AND owner_id = ?", id, testutil.OwnerID); err != nil {
		t.Errorf("cleanup of shift %d: %v", id, err)
	}
}

// TestMergedShift_FullFieldRoundTrip is the merge's central guarantee: one
// endpoint now covers both kinds of row, and every field the retired shifts
// module tracked survives create, list, get, update and delete.
func TestMergedShift_FullFieldRoundTrip(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.ShiftRecord{
		StaffID:              4101,
		DepartmentID:         77,
		Date:                 "2026-06-15",
		StartTime:            "09:30:00",
		EndTime:              "17:45:00",
		BreakDurationMinutes: 45,
		Status:               "scheduled",
		Notes:                "merged round trip",
	}

	created, err := repo.CreateShift(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("CreateShift: %v", err)
	}
	t.Cleanup(func() { deleteShiftRow(t, db, created.ID) })

	// Create must persist every attendance field, not just the template ones.
	if created.StaffID != input.StaffID {
		t.Errorf("StaffID = %d, want %d", created.StaffID, input.StaffID)
	}
	if created.DepartmentID != input.DepartmentID {
		t.Errorf("DepartmentID = %d, want %d", created.DepartmentID, input.DepartmentID)
	}
	if created.Date != input.Date {
		t.Errorf("Date = %q, want %q", created.Date, input.Date)
	}
	if created.StartTime != input.StartTime {
		t.Errorf("StartTime = %q, want %q", created.StartTime, input.StartTime)
	}
	if created.EndTime != input.EndTime {
		t.Errorf("EndTime = %q, want %q", created.EndTime, input.EndTime)
	}
	if created.BreakDurationMinutes != input.BreakDurationMinutes {
		t.Errorf("BreakDurationMinutes = %d, want %d", created.BreakDurationMinutes, input.BreakDurationMinutes)
	}
	if created.Status != input.Status {
		t.Errorf("Status = %q, want %q", created.Status, input.Status)
	}
	if created.Notes != input.Notes {
		t.Errorf("Notes = %q, want %q", created.Notes, input.Notes)
	}
	if created.Kind() != domain.ShiftKindAttendance {
		t.Errorf("Kind = %q, want attendance", created.Kind())
	}

	// The clock and hours columns are written by the staff module's clock-in and
	// clock-out queries, not by this create path, so set them directly and prove
	// the merged read path surfaces them.
	if _, err := db.ExecContext(ctx,
		"UPDATE hr_shifts SET clock_in_at = ?, clock_out_at = ?, hours_worked = ?, overtime_hours = ? WHERE id = ?",
		"2026-06-15 09:30:00", "2026-06-15 17:45:00", "7.50", "1.25", created.ID); err != nil {
		t.Fatalf("seed clock columns: %v", err)
	}

	got, err := repo.GetShift(ctx, testutil.OwnerID, created.ID)
	if err != nil {
		t.Fatalf("GetShift: %v", err)
	}
	if got.HoursWorked != 7.5 {
		t.Errorf("HoursWorked = %v, want 7.5", got.HoursWorked)
	}
	if got.OvertimeHours != 1.25 {
		t.Errorf("OvertimeHours = %v, want 1.25", got.OvertimeHours)
	}
	if got.ClockInAt.IsZero() {
		t.Error("ClockInAt is zero, want the seeded timestamp")
	}
	if got.ClockOutAt.IsZero() {
		t.Error("ClockOutAt is zero, want the seeded timestamp")
	}

	// The row must be reachable through a filtered listing.
	listed, total, err := repo.ListShifts(ctx, testutil.OwnerID, domain.ListShiftsParams{
		Limit: 100, StaffID: input.StaffID, Date: input.Date,
	})
	if err != nil {
		t.Fatalf("ListShifts: %v", err)
	}
	if total < 1 {
		t.Errorf("total = %d, want at least 1", total)
	}
	var found bool
	for _, s := range listed {
		if s.ID == created.ID {
			found = true
			if s.Notes != input.Notes {
				t.Errorf("listed Notes = %q, want %q", s.Notes, input.Notes)
			}
		}
	}
	if !found {
		t.Errorf("shift %d missing from a listing filtered by its own staff_id and date", created.ID)
	}

	// Update every mutable field.
	updated := *got
	updated.Notes = "updated note"
	updated.Status = "completed"
	updated.BreakDurationMinutes = 15
	updated.DepartmentID = 88
	updated.StartTime = "10:00:00"
	updated.EndTime = "18:00:00"
	updated.Date = "2026-06-16"
	if err := repo.UpdateShift(ctx, testutil.OwnerID, updated); err != nil {
		t.Fatalf("UpdateShift: %v", err)
	}

	reread, err := repo.GetShift(ctx, testutil.OwnerID, created.ID)
	if err != nil {
		t.Fatalf("GetShift after update: %v", err)
	}
	if reread.Notes != "updated note" {
		t.Errorf("Notes = %q, want 'updated note'", reread.Notes)
	}
	if reread.Status != "completed" {
		t.Errorf("Status = %q, want completed", reread.Status)
	}
	if reread.BreakDurationMinutes != 15 {
		t.Errorf("BreakDurationMinutes = %d, want 15", reread.BreakDurationMinutes)
	}
	if reread.DepartmentID != 88 {
		t.Errorf("DepartmentID = %d, want 88", reread.DepartmentID)
	}
	if reread.StartTime != "10:00:00" {
		t.Errorf("StartTime = %q, want 10:00:00", reread.StartTime)
	}
	if reread.Date != "2026-06-16" {
		t.Errorf("Date = %q, want 2026-06-16", reread.Date)
	}
	if reread.StaffID != input.StaffID {
		t.Errorf("StaffID = %d, want it unchanged at %d", reread.StaffID, input.StaffID)
	}

	// Delete is new on this module; it came from the retired shifts module.
	if err := repo.DeleteShift(ctx, testutil.OwnerID, created.ID); err != nil {
		t.Fatalf("DeleteShift: %v", err)
	}
	if _, err := repo.GetShift(ctx, testutil.OwnerID, created.ID); !errors.Is(err, domain.ErrShiftNotFound) {
		t.Errorf("after delete, GetShift error = %v, want ErrShiftNotFound", err)
	}
}

// TestMergedShift_KindFilterSeparatesRows proves the two kinds of row sharing
// this table stay distinguishable, which is what stops a template showing up as
// an attendance record.
func TestMergedShift_KindFilterSeparatesRows(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	template, err := repo.CreateShift(ctx, testutil.OwnerID, domain.ShiftRecord{
		LocationID: 4242, Name: "Kind Filter Template",
		StartTime: "06:00:00", EndTime: "12:00:00", Color: "#abcdef",
	})
	if err != nil {
		t.Fatalf("CreateShift template: %v", err)
	}
	t.Cleanup(func() { deleteShiftRow(t, db, template.ID) })

	attendance, err := repo.CreateShift(ctx, testutil.OwnerID, domain.ShiftRecord{
		StaffID: 4109, Date: "2026-06-20",
		StartTime: "06:00:00", EndTime: "12:00:00",
	})
	if err != nil {
		t.Fatalf("CreateShift attendance: %v", err)
	}
	t.Cleanup(func() { deleteShiftRow(t, db, attendance.ID) })

	templates, _, err := repo.ListShifts(ctx, testutil.OwnerID, domain.ListShiftsParams{
		Limit: 1000, Kind: domain.ShiftKindTemplate,
	})
	if err != nil {
		t.Fatalf("ListShifts template: %v", err)
	}
	for _, s := range templates {
		if s.StaffID != 0 {
			t.Errorf("shift %d in the template listing has staff_id %d", s.ID, s.StaffID)
		}
		if s.ID == attendance.ID {
			t.Errorf("attendance row %d leaked into the template listing", s.ID)
		}
	}

	attendances, _, err := repo.ListShifts(ctx, testutil.OwnerID, domain.ListShiftsParams{
		Limit: 1000, Kind: domain.ShiftKindAttendance,
	})
	if err != nil {
		t.Fatalf("ListShifts attendance: %v", err)
	}
	for _, s := range attendances {
		if s.StaffID == 0 {
			t.Errorf("shift %d in the attendance listing has no staff_id", s.ID)
		}
		if s.ID == template.ID {
			t.Errorf("template row %d leaked into the attendance listing", s.ID)
		}
	}
}
