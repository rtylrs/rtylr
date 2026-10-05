package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/hr/hr_shifts/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/hr_shifts/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

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
