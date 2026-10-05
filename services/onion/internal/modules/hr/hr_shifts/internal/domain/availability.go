package domain

import (
	"context"
	"time"
)

// AvailabilityKind distinguishes the two kinds of row hr_staff_availability holds.
type AvailabilityKind string

const (
	// AvailabilityKindRecurring is a standing weekly pattern, keyed by weekday.
	AvailabilityKindRecurring AvailabilityKind = "rule"
	// AvailabilityKindException is a single-date override of the weekly pattern.
	AvailabilityKindException AvailabilityKind = "exception"
)

// Weekday numbering used by hr_staff_availability.day_of_week. Monday-first so
// the value indexes the same week the roster uses, since a roster week starts on
// a Monday. This is deliberately not time.Weekday, where Sunday is 0.
const (
	weekdayMondayFirstMin = 0
	weekdayMondayFirstMax = 6
)

// AvailabilityRecord is a row in hr_staff_availability. Exactly one of DayOfWeek
// and SpecificDate is set; Kind reports which.
//
// StaffID references hr_staff.id, not hr_employees.id. Those are separate tables
// bridged by the nullable hr_staff.employee_id, which matters because shift
// assignments key on employee_id. See StaffResolver.
type AvailabilityRecord struct {
	ID         int64
	OwnerID    int64
	BusinessID int64
	StaffID    int64

	// DayOfWeek is 0 = Monday through 6 = Sunday on a recurring rule, and -1 on a
	// dated exception. SpecificDate is an ISO-8601 date on an exception and empty
	// on a recurring rule.
	DayOfWeek    int32
	SpecificDate string

	// StartTime and EndTime are HH:MM 24-hour. An EndTime earlier than StartTime
	// means the window crosses midnight and ends the following day, which is how
	// availability for a night shift is expressed. Do not assume End > Start.
	StartTime string
	EndTime   string

	// IsAvailable false marks a deliberate block, which is how a dated exception
	// withdraws availability the recurring pattern would otherwise grant.
	IsAvailable bool

	CreatedAt time.Time
	UpdatedAt time.Time
}

// NoDayOfWeek is the DayOfWeek value carried by a dated exception, which has no
// weekday of its own.
const NoDayOfWeek int32 = -1

// Kind reports which kind of row this is.
func (a AvailabilityRecord) Kind() AvailabilityKind {
	if a.SpecificDate != "" {
		return AvailabilityKindException
	}
	return AvailabilityKindRecurring
}

// ListAvailabilityParams filters an availability listing. Every filter is
// optional. DateFrom and DateTo constrain dated exceptions only and leave
// recurring rules in the result, because a recurring rule carries no date;
// narrow those with DayOfWeek instead.
type ListAvailabilityParams struct {
	Limit  int32
	Offset int32

	StaffID   int64
	DayOfWeek *int32
	DateFrom  string
	DateTo    string
}

// ValidateAvailability checks the invariants the contract states but that `.onk`
// has no decorator to express, so they cannot be enforced at the schema layer.
func ValidateAvailability(a AvailabilityRecord) error {
	hasDay := a.DayOfWeek != NoDayOfWeek
	hasDate := a.SpecificDate != ""

	if hasDay == hasDate {
		return ErrAvailabilityKindAmbiguous
	}
	if hasDay && (a.DayOfWeek < weekdayMondayFirstMin || a.DayOfWeek > weekdayMondayFirstMax) {
		return ErrAvailabilityDayOutOfRange
	}
	if hasDate {
		if _, err := time.Parse(time.DateOnly, a.SpecificDate); err != nil {
			return ErrAvailabilityDateInvalid
		}
	}
	if !isHourMinute(a.StartTime) || !isHourMinute(a.EndTime) {
		return ErrAvailabilityTimeInvalid
	}
	return nil
}

// isHourMinute reports whether s is an HH:MM 24-hour time. It accepts HH:MM:SS
// too, because MySQL hands a TIME column back in that form.
func isHourMinute(s string) bool {
	if _, err := time.Parse("15:04", s); err == nil {
		return true
	}
	_, err := time.Parse("15:04:05", s)
	return err == nil
}

// StaffResolver maps between the identities availability and shift assignments
// use. Availability keys on hr_staff.id; assignments key on hr_employees.id.
type StaffResolver interface {
	// StaffIDByUserID returns the caller's own hr_staff row, used to decide
	// whether a write is self-service.
	StaffIDByUserID(ctx context.Context, businessID, userID int64) (int64, error)
	// StaffIDByEmployeeID crosses the nullable hr_staff.employee_id bridge to
	// find whose availability an assignment should be checked against. It
	// returns ErrStaffNotLinked when no staff row is linked to the employee.
	StaffIDByEmployeeID(ctx context.Context, businessID, employeeID int64) (int64, error)
}

// PermissionProvider gates writes against another staff member's availability.
// It mirrors the interface hr_payroll uses, so the gate style stays uniform
// across HR modules.
type PermissionProvider interface {
	GetEmployeeIDByUserID(ctx context.Context, ownerID, userID int64) (int64, error)
	HasPermission(ctx context.Context, ownerID, employeeID int64, permissionKey string) (bool, error)
}

// PermissionManageSchedules is the permission key required to write another
// staff member's availability. Reading or writing your own never needs it.
const PermissionManageSchedules = "manage_schedules"
