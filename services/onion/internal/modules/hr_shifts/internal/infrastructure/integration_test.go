package infrastructure_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/hr_shifts/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_shifts/internal/infrastructure"
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

// TestRosterFlow_UnchangedByMerge walks the roster workflow the surviving module
// owns, which the retired shifts module had no equivalent of, to prove the merge
// left it working.
func TestRosterFlow_UnchangedByMerge(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const (
		employeeID int64 = 4108
		weekStart        = "2026-09-07"
		weekEnd          = "2026-09-14"
		shiftDate        = "2026-09-08"
	)

	shift, err := repo.CreateShift(ctx, testutil.OwnerID, domain.ShiftRecord{
		LocationID: 5, Name: "Roster Flow Shift",
		StartTime: "08:00:00", EndTime: "16:00:00", Color: "#654321",
	})
	if err != nil {
		t.Fatalf("CreateShift: %v", err)
	}
	t.Cleanup(func() { deleteShiftRow(t, db, shift.ID) })

	t.Cleanup(func() {
		if _, err := db.ExecContext(context.Background(),
			"DELETE FROM hr_shift_assignments WHERE owner_id = ? AND employee_id = ? AND date = ?",
			testutil.OwnerID, employeeID, shiftDate); err != nil {
			t.Errorf("cleanup of assignment: %v", err)
		}
	})

	assignment, err := repo.CreateShiftAssignment(ctx, testutil.OwnerID, domain.ShiftAssignmentRecord{
		EmployeeID: employeeID,
		ShiftID:    shift.ID,
		Date:       shiftDate,
	})
	if err != nil {
		t.Fatalf("CreateShiftAssignment: %v", err)
	}
	if assignment.ID == 0 {
		t.Fatal("created assignment has no ID")
	}
	if assignment.Status != domain.AssignmentStatusScheduled {
		t.Errorf("assignment status = %q, want scheduled", assignment.Status)
	}
	// The create response has to carry the stored timestamps. Reconstructing the
	// record in Go instead of reading it back left these zero, which serialises
	// as a nonsense Unix value.
	if assignment.CreatedAt.IsZero() {
		t.Error("assignment CreatedAt is zero, want the stored timestamp")
	}
	if assignment.UpdatedAt.IsZero() {
		t.Error("assignment UpdatedAt is zero, want the stored timestamp")
	}

	// The roster listing must find it, and it starts out unpublished.
	assignments, err := repo.ListRosterAssignments(ctx, testutil.OwnerID, weekStart, weekEnd)
	if err != nil {
		t.Fatalf("ListRosterAssignments: %v", err)
	}
	var before *domain.ShiftAssignmentRecord
	for i := range assignments {
		if assignments[i].ID == assignment.ID {
			before = &assignments[i]
		}
	}
	if before == nil {
		t.Fatalf("assignment %d missing from the roster for its own week", assignment.ID)
	}
	if before.Published {
		t.Error("assignment is published before the roster was published")
	}

	if err := repo.PublishRosterAssignments(ctx, testutil.OwnerID, weekStart, weekEnd); err != nil {
		t.Fatalf("PublishRosterAssignments: %v", err)
	}

	after, err := repo.ListRosterAssignments(ctx, testutil.OwnerID, weekStart, weekEnd)
	if err != nil {
		t.Fatalf("ListRosterAssignments after publish: %v", err)
	}
	var published bool
	for _, a := range after {
		if a.ID == assignment.ID {
			published = a.Published
		}
	}
	if !published {
		t.Errorf("assignment %d is not published after publishing its week", assignment.ID)
	}

	// Conflict detection backs the roster builder, so it has to keep seeing the
	// assignment for that employee and date.
	forDate, err := repo.ListEmployeeAssignmentsForDate(ctx, testutil.OwnerID, employeeID, shiftDate)
	if err != nil {
		t.Fatalf("ListEmployeeAssignmentsForDate: %v", err)
	}
	if len(forDate) == 0 {
		t.Error("expected the assignment to be visible to conflict detection")
	}
}

func TestListShifts_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	shifts, total, err := repo.ListShifts(ctx, testutil.OwnerID, domain.ListShiftsParams{Limit: 50, Offset: 0})
	if err != nil {
		t.Fatalf("ListShifts: %v", err)
	}
	if total < 0 {
		t.Errorf("total count is negative: %d", total)
	}
	_ = shifts
}

func TestListShifts_AllBelongToOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	shifts, _, err := repo.ListShifts(ctx, testutil.OwnerID, domain.ListShiftsParams{Limit: 50, Offset: 0})
	if err != nil {
		t.Fatalf("ListShifts: %v", err)
	}
	for _, s := range shifts {
		if s.OwnerID != testutil.OwnerID {
			t.Errorf("shift %d has owner_id %d, want %d", s.ID, s.OwnerID, testutil.OwnerID)
		}
	}
}

func TestListShifts_PaginationOffset_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, _, err := repo.ListShifts(ctx, testutil.OwnerID, domain.ListShiftsParams{Limit: 10, Offset: 100})
	if err != nil {
		t.Fatalf("ListShifts with large offset: %v", err)
	}
}

func TestCreateShift_AndGetShift(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.ShiftRecord{
		LocationID: 0,
		Name:       "Test Morning Shift",
		StartTime:  "08:00",
		EndTime:    "16:00",
		Color:      "#336699",
	}

	created, err := repo.CreateShift(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("CreateShift: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("created shift has no ID")
	}
	if created.Name != input.Name {
		t.Errorf("shift name = %q, want %q", created.Name, input.Name)
	}

	got, err := repo.GetShift(ctx, testutil.OwnerID, created.ID)
	if err != nil {
		t.Fatalf("GetShift after create: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("GetShift id = %d, want %d", got.ID, created.ID)
	}
}

func TestGetShift_NotFound_ReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetShift(ctx, testutil.OwnerID, 9999999)
	if err == nil {
		t.Fatal("expected error for non-existent shift, got nil")
	}
}

func TestUpdateShift_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.ShiftRecord{
		LocationID: 0,
		Name:       "Update Test Shift",
		StartTime:  "09:00",
		EndTime:    "17:00",
		Color:      "#ff0000",
	}
	created, err := repo.CreateShift(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("CreateShift for update test: %v", err)
	}

	created.Name = "Updated Shift Name"
	if err := repo.UpdateShift(ctx, testutil.OwnerID, *created); err != nil {
		t.Fatalf("UpdateShift: %v", err)
	}

	got, err := repo.GetShift(ctx, testutil.OwnerID, created.ID)
	if err != nil {
		t.Fatalf("GetShift after update: %v", err)
	}
	if got.Name != "Updated Shift Name" {
		t.Errorf("updated name = %q, want %q", got.Name, "Updated Shift Name")
	}
}

func TestListRosterAssignments_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	assignments, err := repo.ListRosterAssignments(ctx, testutil.OwnerID, "2025-01-06", "2025-01-12")
	if err != nil {
		t.Fatalf("ListRosterAssignments: %v", err)
	}
	_ = assignments
}

func TestListShiftsByLocation_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	shifts, err := repo.ListShiftsByLocation(ctx, testutil.OwnerID, 1)
	if err != nil {
		t.Fatalf("ListShiftsByLocation: %v", err)
	}
	_ = shifts
}

func TestListEmployeeAssignmentsForDate_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	assignments, err := repo.ListEmployeeAssignmentsForDate(ctx, testutil.OwnerID, 4101, "2025-06-01")
	if err != nil {
		t.Fatalf("ListEmployeeAssignmentsForDate: %v", err)
	}
	_ = assignments
}

func TestCreateShiftAssignment_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	shift, err := repo.CreateShift(ctx, testutil.OwnerID, domain.ShiftRecord{
		Name:      "Assignment Test Shift",
		StartTime: "07:00",
		EndTime:   "15:00",
		Color:     "#00ff00",
	})
	if err != nil {
		t.Fatalf("CreateShift for assignment test: %v", err)
	}

	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM hr_shift_assignments WHERE owner_id = ? AND employee_id = ? AND date = ?", testutil.OwnerID, int64(4101), "2025-07-01")
	})

	assignment := domain.ShiftAssignmentRecord{
		EmployeeID: 4101,
		ShiftID:    shift.ID,
		Date:       "2025-07-01",
	}
	created, err := repo.CreateShiftAssignment(ctx, testutil.OwnerID, assignment)
	if err != nil {
		t.Fatalf("CreateShiftAssignment: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("created assignment has no ID")
	}
}

func TestPublishRosterAssignments_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Publish an empty week -- should be a no-op but not error.
	err := repo.PublishRosterAssignments(ctx, testutil.OwnerID, "2025-06-02", "2025-06-08")
	if err != nil {
		t.Fatalf("PublishRosterAssignments: %v", err)
	}
}

func TestCreateShift_OwnerIDIsSet(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	created, err := repo.CreateShift(ctx, testutil.OwnerID, domain.ShiftRecord{
		Name:      "Owner ID Check Shift",
		StartTime: "06:00",
		EndTime:   "14:00",
		Color:     "#123456",
	})
	if err != nil {
		t.Fatalf("CreateShift: %v", err)
	}
	if created.OwnerID != testutil.OwnerID {
		t.Errorf("OwnerID = %d, want %d", created.OwnerID, testutil.OwnerID)
	}
}

func TestCreateShift_StartsWithPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	created, err := repo.CreateShift(ctx, testutil.OwnerID, domain.ShiftRecord{
		Name:      "Positive ID Shift",
		StartTime: "10:00",
		EndTime:   "18:00",
		Color:     "#aabbcc",
	})
	if err != nil {
		t.Fatalf("CreateShift: %v", err)
	}
	if created.ID <= 0 {
		t.Errorf("expected positive ID, got %d", created.ID)
	}
}

func TestUpdateShift_ColorChanges(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	created, err := repo.CreateShift(ctx, testutil.OwnerID, domain.ShiftRecord{
		Name:      "Color Change Shift",
		StartTime: "08:00",
		EndTime:   "16:00",
		Color:     "#000000",
	})
	if err != nil {
		t.Fatalf("CreateShift: %v", err)
	}

	created.Color = "#ffffff"
	if err := repo.UpdateShift(ctx, testutil.OwnerID, *created); err != nil {
		t.Fatalf("UpdateShift: %v", err)
	}

	got, err := repo.GetShift(ctx, testutil.OwnerID, created.ID)
	if err != nil {
		t.Fatalf("GetShift: %v", err)
	}
	if got.Color != "#ffffff" {
		t.Errorf("Color = %q, want '#ffffff'", got.Color)
	}
}

func TestListShifts_CountMatchesTotalField(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	shifts, total, err := repo.ListShifts(ctx, testutil.OwnerID, domain.ListShiftsParams{Limit: 1000, Offset: 0})
	if err != nil {
		t.Fatalf("ListShifts: %v", err)
	}
	if int64(len(shifts)) > total {
		t.Errorf("returned %d shifts but total=%d", len(shifts), total)
	}
}

func TestListShiftsByLocation_AllBelongToOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	shifts, err := repo.ListShiftsByLocation(ctx, testutil.OwnerID, 1)
	if err != nil {
		t.Fatalf("ListShiftsByLocation: %v", err)
	}
	for _, s := range shifts {
		if s.OwnerID != testutil.OwnerID {
			t.Errorf("shift %d has OwnerID=%d, want %d", s.ID, s.OwnerID, testutil.OwnerID)
		}
	}
}

func TestCreateShiftAssignment_AssignmentHasScheduledStatus(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	shift, err := repo.CreateShift(ctx, testutil.OwnerID, domain.ShiftRecord{
		Name:      "Status Check Shift",
		StartTime: "09:00",
		EndTime:   "17:00",
		Color:     "#112233",
	})
	if err != nil {
		t.Fatalf("CreateShift: %v", err)
	}

	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM hr_shift_assignments WHERE owner_id = ? AND employee_id = ? AND date = ?", testutil.OwnerID, int64(4102), "2025-08-01")
	})

	assignment := domain.ShiftAssignmentRecord{
		EmployeeID: 4102,
		ShiftID:    shift.ID,
		Date:       "2025-08-01",
	}
	created, err := repo.CreateShiftAssignment(ctx, testutil.OwnerID, assignment)
	if err != nil {
		t.Fatalf("CreateShiftAssignment: %v", err)
	}
	if created.Status != "scheduled" {
		t.Errorf("expected status 'scheduled', got %q", created.Status)
	}
}

func TestListRosterAssignments_AllBelongToOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	assignments, err := repo.ListRosterAssignments(ctx, testutil.OwnerID, "2025-01-06", "2025-01-12")
	if err != nil {
		t.Fatalf("ListRosterAssignments: %v", err)
	}
	for _, a := range assignments {
		if a.OwnerID != testutil.OwnerID {
			t.Errorf("assignment %d has OwnerID=%d, want %d", a.ID, a.OwnerID, testutil.OwnerID)
		}
	}
}

func TestGetShift_ReturnsCorrectName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	created, err := repo.CreateShift(ctx, testutil.OwnerID, domain.ShiftRecord{
		Name:      "Name Verify Shift",
		StartTime: "07:00",
		EndTime:   "15:00",
		Color:     "#445566",
	})
	if err != nil {
		t.Fatalf("CreateShift: %v", err)
	}

	got, err := repo.GetShift(ctx, testutil.OwnerID, created.ID)
	if err != nil {
		t.Fatalf("GetShift: %v", err)
	}
	if got.Name != "Name Verify Shift" {
		t.Errorf("Name = %q, want 'Name Verify Shift'", got.Name)
	}
}

func TestListEmployeeAssignmentsForDate_ReturnsSlice(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	assignments, err := repo.ListEmployeeAssignmentsForDate(ctx, testutil.OwnerID, 4101, "2025-09-01")
	if err != nil {
		t.Fatalf("ListEmployeeAssignmentsForDate: %v", err)
	}
	if assignments == nil {
		t.Fatal("expected non-nil slice")
	}
}

func TestCreateShiftAssignment_EmployeeIDStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	shift, err := repo.CreateShift(ctx, testutil.OwnerID, domain.ShiftRecord{
		Name:      "Employee ID Store Shift",
		StartTime: "11:00",
		EndTime:   "19:00",
		Color:     "#667788",
	})
	if err != nil {
		t.Fatalf("CreateShift: %v", err)
	}

	const empID int64 = 4103
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM hr_shift_assignments WHERE owner_id = ? AND employee_id = ? AND date = ?", testutil.OwnerID, empID, "2025-10-01")
	})
	created, err := repo.CreateShiftAssignment(ctx, testutil.OwnerID, domain.ShiftAssignmentRecord{
		EmployeeID: empID,
		ShiftID:    shift.ID,
		Date:       "2025-10-01",
	})
	if err != nil {
		t.Fatalf("CreateShiftAssignment: %v", err)
	}
	if created.EmployeeID != empID {
		t.Errorf("EmployeeID = %d, want %d", created.EmployeeID, empID)
	}
}

// TestCreateShiftAssignment_DuplicateReturnsRosterConflict covers the unique
// key on (owner_id, employee_id, date): a second assignment for the same
// employee and day must surface as the roster conflict the endpoint maps to
// 409, not as a raw duplicate-key error that becomes a 500.
func TestCreateShiftAssignment_DuplicateReturnsRosterConflict(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const (
		employeeID int64 = 4111
		shiftDate        = "2026-11-03"
	)

	shift, err := repo.CreateShift(ctx, testutil.OwnerID, domain.ShiftRecord{
		Name:      "Duplicate Assignment Shift",
		StartTime: "09:00",
		EndTime:   "17:00",
		Color:     "#334455",
	})
	if err != nil {
		t.Fatalf("CreateShift: %v", err)
	}
	t.Cleanup(func() {
		deleteShiftRow(t, db, shift.ID)
		if _, err := db.ExecContext(context.Background(),
			"DELETE FROM hr_shift_assignments WHERE owner_id = ? AND employee_id = ? AND date = ?",
			testutil.OwnerID, employeeID, shiftDate); err != nil {
			t.Errorf("cleanup of assignment: %v", err)
		}
	})

	assignment := domain.ShiftAssignmentRecord{
		EmployeeID: employeeID,
		ShiftID:    shift.ID,
		Date:       shiftDate,
	}
	if _, err := repo.CreateShiftAssignment(ctx, testutil.OwnerID, assignment); err != nil {
		t.Fatalf("first CreateShiftAssignment: %v", err)
	}

	if _, err := repo.CreateShiftAssignment(ctx, testutil.OwnerID, assignment); !errors.Is(err, domain.ErrRosterConflict) {
		t.Fatalf("duplicate CreateShiftAssignment error = %v, want ErrRosterConflict", err)
	}
}
