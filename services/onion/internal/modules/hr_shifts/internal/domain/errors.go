package domain

import "errors"

var (
	ErrShiftNotFound  = errors.New("shift not found")
	ErrRosterConflict = errors.New("shift assignment conflicts with approved leave or another shift")

	ErrAvailabilityNotFound = errors.New("availability entry not found")
	// ErrAvailabilityKindAmbiguous is returned when a write sets both a weekday
	// and a specific date, or neither. Exactly one identifies the kind of row.
	ErrAvailabilityKindAmbiguous = errors.New("exactly one of day_of_week and specific_date must be set")
	ErrAvailabilityDayOutOfRange = errors.New("day_of_week must be between 0 (Monday) and 6 (Sunday)")
	ErrAvailabilityDateInvalid   = errors.New("specific_date must be a valid YYYY-MM-DD date")
	ErrAvailabilityTimeInvalid   = errors.New("start_time and end_time must be HH:MM 24-hour times")
	// ErrStaffNotLinked is returned when an employee has no linked hr_staff row,
	// so their availability cannot be resolved.
	ErrStaffNotLinked = errors.New("employee has no linked staff record")
	// ErrPermissionDenied is returned when a caller tries to read or write another
	// staff member's availability without manage_schedules.
	ErrPermissionDenied = errors.New("permission denied")
)
