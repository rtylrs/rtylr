package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/hr/hr_shifts/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/hr_shifts/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

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
