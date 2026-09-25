package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/hr_shifts/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_shifts/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

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
