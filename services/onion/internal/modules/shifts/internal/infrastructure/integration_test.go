package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/shifts/internal/domain"
	"github.com/voxire/onion/internal/modules/shifts/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

const fixtureStaffIDForShifts int64 = 4101 // Alice

func TestListShifts_ReturnsNoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	shifts, err := repo.ListShifts(ctx, testutil.OwnerID, 0, "")
	if err != nil {
		t.Fatalf("shifts.ListShifts: unexpected error: %v", err)
	}
	if shifts == nil {
		t.Fatal("shifts.ListShifts: expected non-nil slice, got nil")
	}
}

func TestListShifts_WrongOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	shifts, err := repo.ListShifts(ctx, 0, 0, "")
	if err != nil {
		t.Fatalf("shifts.ListShifts(ownerID=0): unexpected error: %v", err)
	}
	if len(shifts) != 0 {
		t.Fatalf("shifts.ListShifts(ownerID=0): expected 0 rows, got %d", len(shifts))
	}
}

func TestCreateShift_ReturnsInsertedID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.ShiftInput{
		OwnerID:   testutil.OwnerID,
		StaffID:   fixtureStaffIDForShifts,
		Date:      "2026-06-01",
		StartTime: "09:00",
		EndTime:   "17:00",
	}

	id, err := repo.CreateShift(ctx, input)
	if err != nil {
		t.Fatalf("shifts.CreateShift: unexpected error: %v", err)
	}
	if id <= 0 {
		t.Fatalf("shifts.CreateShift: expected positive ID, got %d", id)
	}
}

func TestCreateShift_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.ShiftInput{
		OwnerID:   testutil.OwnerID,
		StaffID:   fixtureStaffIDForShifts,
		Date:      "2026-06-02",
		StartTime: "08:00",
		EndTime:   "16:00",
	}

	id, err := repo.CreateShift(ctx, input)
	if err != nil {
		t.Fatalf("shifts.CreateShift: %v", err)
	}

	shifts, err := repo.ListShifts(ctx, testutil.OwnerID, 0, "")
	if err != nil {
		t.Fatalf("shifts.ListShifts: %v", err)
	}

	var found bool
	for _, s := range shifts {
		if s.ID == id {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("created shift id=%d not found in list", id)
	}
}

func TestListShifts_FilterByStaffID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.CreateShift(ctx, domain.ShiftInput{
		OwnerID:   testutil.OwnerID,
		StaffID:   fixtureStaffIDForShifts,
		Date:      "2026-06-03",
		StartTime: "07:00",
		EndTime:   "15:00",
	})
	if err != nil {
		t.Fatalf("seed shift: %v", err)
	}

	shifts, err := repo.ListShifts(ctx, testutil.OwnerID, fixtureStaffIDForShifts, "")
	if err != nil {
		t.Fatalf("shifts.ListShifts(staffID filter): %v", err)
	}

	for _, s := range shifts {
		if s.StaffID != fixtureStaffIDForShifts {
			t.Errorf("expected staffID=%d, got %d", fixtureStaffIDForShifts, s.StaffID)
		}
	}
}

func TestListShifts_FilterByDate(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const targetDate = "2026-06-10"

	_, err := repo.CreateShift(ctx, domain.ShiftInput{
		OwnerID:   testutil.OwnerID,
		StaffID:   fixtureStaffIDForShifts,
		Date:      targetDate,
		StartTime: "09:00",
		EndTime:   "17:00",
	})
	if err != nil {
		t.Fatalf("seed shift: %v", err)
	}

	shifts, err := repo.ListShifts(ctx, testutil.OwnerID, 0, targetDate)
	if err != nil {
		t.Fatalf("shifts.ListShifts(date filter): %v", err)
	}

	for _, s := range shifts {
		d := s.Date.Format("2006-01-02")
		if d != targetDate {
			t.Errorf("expected date=%s, got %s (shift id=%d)", targetDate, d, s.ID)
		}
	}
}

func TestDeleteShift_RemovedFromList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateShift(ctx, domain.ShiftInput{
		OwnerID:   testutil.OwnerID,
		StaffID:   fixtureStaffIDForShifts,
		Date:      "2026-07-01",
		StartTime: "09:00",
		EndTime:   "17:00",
	})
	if err != nil {
		t.Fatalf("shifts.CreateShift: %v", err)
	}

	err = repo.DeleteShift(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("shifts.DeleteShift: %v", err)
	}

	shifts, err := repo.ListShifts(ctx, testutil.OwnerID, 0, "")
	if err != nil {
		t.Fatalf("shifts.ListShifts after delete: %v", err)
	}

	for _, s := range shifts {
		if s.ID == id {
			t.Fatalf("deleted shift id=%d still present in list", id)
		}
	}
}

func TestUpdateShift_StatusChanges(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateShift(ctx, domain.ShiftInput{
		OwnerID:   testutil.OwnerID,
		StaffID:   fixtureStaffIDForShifts,
		Date:      "2026-07-15",
		StartTime: "09:00",
		EndTime:   "17:00",
	})
	if err != nil {
		t.Fatalf("shifts.CreateShift: %v", err)
	}

	err = repo.UpdateShift(ctx, testutil.OwnerID, domain.ShiftUpdateInput{
		ID:        id,
		Date:      "2026-07-15",
		StartTime: "09:00",
		EndTime:   "17:00",
		Status:    "completed",
	})
	if err != nil {
		t.Fatalf("shifts.UpdateShift: %v", err)
	}

	shifts, err := repo.ListShifts(ctx, testutil.OwnerID, 0, "")
	if err != nil {
		t.Fatalf("shifts.ListShifts after update: %v", err)
	}

	for _, s := range shifts {
		if s.ID == id {
			if s.Status != "completed" {
				t.Errorf("expected status=completed, got %q", s.Status)
			}
			return
		}
	}
	t.Fatalf("updated shift id=%d not found in list", id)
}

func TestListShifts_EachHasPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.CreateShift(ctx, domain.ShiftInput{
		OwnerID:   testutil.OwnerID,
		StaffID:   fixtureStaffIDForShifts,
		Date:      "2026-08-01",
		StartTime: "09:00",
		EndTime:   "17:00",
	})
	if err != nil {
		t.Fatalf("seed shift: %v", err)
	}

	shifts, err := repo.ListShifts(ctx, testutil.OwnerID, 0, "")
	if err != nil {
		t.Fatalf("shifts.ListShifts: %v", err)
	}

	for _, s := range shifts {
		if s.ID <= 0 {
			t.Errorf("shift has non-positive ID: %d", s.ID)
		}
	}
}

func TestListShifts_ScopedToOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.CreateShift(ctx, domain.ShiftInput{
		OwnerID:   testutil.OwnerID,
		StaffID:   fixtureStaffIDForShifts,
		Date:      "2026-08-02",
		StartTime: "09:00",
		EndTime:   "17:00",
	})
	if err != nil {
		t.Fatalf("seed shift: %v", err)
	}

	other, err := repo.ListShifts(ctx, testutil.OwnerID+9999, 0, "")
	if err != nil {
		t.Fatalf("shifts.ListShifts(other owner): %v", err)
	}
	for _, s := range other {
		if s.OwnerID == testutil.OwnerID {
			t.Errorf("other-owner list contains ownerID=%d shift", testutil.OwnerID)
		}
	}
}

func TestCreateMultipleShifts_CountIncreases(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	before, err := repo.ListShifts(ctx, testutil.OwnerID, 0, "")
	if err != nil {
		t.Fatalf("shifts.ListShifts before: %v", err)
	}

	for i := 0; i < 3; i++ {
		_, err := repo.CreateShift(ctx, domain.ShiftInput{
			OwnerID:   testutil.OwnerID,
			StaffID:   fixtureStaffIDForShifts,
			Date:      "2026-09-01",
			StartTime: "09:00",
			EndTime:   "17:00",
		})
		if err != nil {
			t.Fatalf("shifts.CreateShift[%d]: %v", i, err)
		}
	}

	after, err := repo.ListShifts(ctx, testutil.OwnerID, 0, "")
	if err != nil {
		t.Fatalf("shifts.ListShifts after: %v", err)
	}

	if len(after) <= len(before) {
		t.Fatalf("expected list to grow: before=%d after=%d", len(before), len(after))
	}
}
