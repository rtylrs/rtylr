package domain

import (
	"errors"
	"strings"
)

var (
	ErrEmptyShiftName  = errors.New("shift name is required")
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

// ValidateShift checks that required shift fields are present.
func ValidateShift(s ShiftRecord) error {
	if strings.TrimSpace(s.Name) == "" {
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
