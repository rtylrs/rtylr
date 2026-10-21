package domain_test

import (
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/hr/hr_shifts/internal/domain"
)

func TestIsValidAssignmentStatus_AllValid(t *testing.T) {
	for _, s := range []string{
		domain.AssignmentStatusScheduled,
		domain.AssignmentStatusConfirmed,
		domain.AssignmentStatusSwapped,
		domain.AssignmentStatusCanceled,
	} {
		if !domain.IsValidAssignmentStatus(s) {
			t.Errorf("expected %q to be valid", s)
		}
	}
}

func TestIsValidAssignmentStatus_Invalid(t *testing.T) {
	if domain.IsValidAssignmentStatus("pending") {
		t.Error("expected 'pending' to be invalid")
	}
}

func TestValidateShift_Valid(t *testing.T) {
	s := domain.ShiftRecord{
		Name:      "Morning",
		StartTime: "08:00",
		EndTime:   "16:00",
	}
	if err := domain.ValidateShift(s); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateShift_EmptyName(t *testing.T) {
	s := domain.ShiftRecord{Name: "", StartTime: "08:00", EndTime: "16:00"}
	if err := domain.ValidateShift(s); !errors.Is(err, domain.ErrEmptyShiftName) {
		t.Errorf("expected ErrEmptyShiftName, got %v", err)
	}
}

func TestValidateShift_EmptyStartTime(t *testing.T) {
	s := domain.ShiftRecord{Name: "Night", StartTime: "", EndTime: "06:00"}
	if err := domain.ValidateShift(s); !errors.Is(err, domain.ErrEmptyShiftTime) {
		t.Errorf("expected ErrEmptyShiftTime, got %v", err)
	}
}

func TestValidateShift_EmptyEndTime(t *testing.T) {
	s := domain.ShiftRecord{Name: "Day", StartTime: "09:00", EndTime: ""}
	if err := domain.ValidateShift(s); !errors.Is(err, domain.ErrEmptyShiftTime) {
		t.Errorf("expected ErrEmptyShiftTime, got %v", err)
	}
}

func TestValidateShiftAssignment_Valid(t *testing.T) {
	a := domain.ShiftAssignmentRecord{
		ShiftID:    1,
		EmployeeID: 4101,
		Date:       "2026-06-01",
	}
	if err := domain.ValidateShiftAssignment(a); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateShiftAssignment_ZeroShiftID(t *testing.T) {
	a := domain.ShiftAssignmentRecord{ShiftID: 0, EmployeeID: 4101, Date: "2026-06-01"}
	if err := domain.ValidateShiftAssignment(a); !errors.Is(err, domain.ErrZeroShiftID) {
		t.Errorf("expected ErrZeroShiftID, got %v", err)
	}
}

func TestValidateShiftAssignment_ZeroEmployeeID(t *testing.T) {
	a := domain.ShiftAssignmentRecord{ShiftID: 1, EmployeeID: 0, Date: "2026-06-01"}
	if err := domain.ValidateShiftAssignment(a); !errors.Is(err, domain.ErrZeroEmployeeID) {
		t.Errorf("expected ErrZeroEmployeeID, got %v", err)
	}
}

func TestValidateShiftAssignment_EmptyDate(t *testing.T) {
	a := domain.ShiftAssignmentRecord{ShiftID: 1, EmployeeID: 4101, Date: ""}
	if err := domain.ValidateShiftAssignment(a); !errors.Is(err, domain.ErrEmptyAssignDate) {
		t.Errorf("expected ErrEmptyAssignDate, got %v", err)
	}
}
