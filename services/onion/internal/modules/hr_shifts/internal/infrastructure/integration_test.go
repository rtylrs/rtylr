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
