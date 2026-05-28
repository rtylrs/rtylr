package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/hr_shifts/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_shifts/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

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
