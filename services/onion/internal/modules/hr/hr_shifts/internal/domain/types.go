package domain

import "time"

// ShiftKind distinguishes the two kinds of row the hr_shifts table holds.
type ShiftKind string

const (
	// ShiftKindAny does not restrict a listing to either kind.
	ShiftKindAny ShiftKind = ""
	// ShiftKindTemplate is a reusable shift for a location, with no staff member.
	ShiftKindTemplate ShiftKind = "template"
	// ShiftKindAttendance is one staff member's worked shift on a given date.
	ShiftKindAttendance ShiftKind = "attendance"
)

// IsValidShiftKind reports whether kind is one this module recognises.
func IsValidShiftKind(kind ShiftKind) bool {
	switch kind {
	case ShiftKindAny, ShiftKindTemplate, ShiftKindAttendance:
		return true
	}
	return false
}

// ShiftRecord is a row in hr_shifts. The table stores two kinds of record and
// this struct covers both, because a single table cannot be split without a
// migration and both kinds are read through the same endpoint.
//
// A template row carries LocationID, Name and Color and leaves StaffID zero.
// An attendance row carries StaffID, Date and the clock and hours fields, and
// leaves LocationID zero and Name empty. Kind reports which one a row is.
type ShiftRecord struct {
	ID         int64
	OwnerID    int64
	BusinessID int64
	LocationID int64
	Name       string
	StartTime  string
	EndTime    string
	Color      string

	// The fields below were carried by the retired shifts module, which
	// registered the same /hr/shifts routes against this same table.
	StaffID              int64
	DepartmentID         int64
	Date                 string
	BreakDurationMinutes int32
	Status               string
	ClockInAt            time.Time
	ClockOutAt           time.Time
	HoursWorked          float64
	OvertimeHours        float64
	Notes                string

	CreatedAt time.Time
	UpdatedAt time.Time
}

// Kind reports which kind of row this is. StaffID is the discriminator: the
// column is NULL on template rows and set on attendance rows.
func (s ShiftRecord) Kind() ShiftKind {
	if s.StaffID == 0 {
		return ShiftKindTemplate
	}
	return ShiftKindAttendance
}

type ShiftAssignmentRecord struct {
	ID         int64
	OwnerID    int64
	BusinessID int64
	EmployeeID int64
	ShiftID    int64
	Date       string
	Status     string
	Published  bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// ListShiftsParams filters a shift listing. Every filter is optional; a zero
// value means "do not filter on this field", so the default listing returns
// both kinds of row.
type ListShiftsParams struct {
	Limit  int32
	Offset int32

	StaffID    int64
	Date       string
	LocationID int64
	Kind       ShiftKind
}

type RosterParams struct {
	LocationID int64
	WeekStart  string
}

type ConflictInfo struct {
	EmployeeID   int64
	Date         string
	ConflictType string
	Description  string
}
