package domain

import (
	"errors"
	"strings"
)

var (
	ErrEmptyShiftName  = errors.New("shift name is required")
	ErrEmptyShiftDate  = errors.New("shift date is required for a staff shift")
	ErrEmptyShiftTime  = errors.New("shift start and end times are required")
	ErrEmptyAssignDate = errors.New("assignment date is required")
	ErrZeroShiftID     = errors.New("shift ID must be non-zero")
	ErrZeroEmployeeID  = errors.New("employee ID must be non-zero")
)

// IsValidAssignmentStatus returns true when the status is recognised.
func IsValidAssignmentStatus(status string) bool {
	switch status {
	case AssignmentStatusScheduled, AssignmentStatusConfirmed, AssignmentStatusSwapped, AssignmentStatusCanceled:
		return true
	}
	return false
}

// ValidateShift checks that the fields required for this shift's kind are
// present. A template row needs a name; an attendance row needs a date instead,
// because it belongs to a staff member on a specific day and carries no name.
// Both kinds need a start and end time.
func ValidateShift(s ShiftRecord) error {
	if s.Kind() == ShiftKindAttendance {
		if strings.TrimSpace(s.Date) == "" {
			return ErrEmptyShiftDate
		}
	} else if strings.TrimSpace(s.Name) == "" {
		return ErrEmptyShiftName
	}
	if strings.TrimSpace(s.StartTime) == "" || strings.TrimSpace(s.EndTime) == "" {
		return ErrEmptyShiftTime
	}
	return nil
}

// ValidateShiftAssignment checks required assignment fields.
func ValidateShiftAssignment(a ShiftAssignmentRecord) error {
	if a.ShiftID == 0 {
		return ErrZeroShiftID
	}
	if a.EmployeeID == 0 {
		return ErrZeroEmployeeID
	}
	if strings.TrimSpace(a.Date) == "" {
		return ErrEmptyAssignDate
	}
	return nil
}
