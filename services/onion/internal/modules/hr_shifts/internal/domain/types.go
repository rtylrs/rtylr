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
	ID         int64  `json:"id"`
	OwnerID    int64  `json:"owner_id"`
	BusinessID int64  `json:"business_id"`
	LocationID int64  `json:"location_id"`
	Name       string `json:"name"`
	StartTime  string `json:"start_time"`
	EndTime    string `json:"end_time"`
	Color      string `json:"color"`

	// The fields below were carried by the retired shifts module, which
	// registered the same /hr/shifts routes against this same table.
	StaffID              int64     `json:"staff_id"`
	DepartmentID         int64     `json:"department_id"`
	Date                 string    `json:"date"`
	BreakDurationMinutes int32     `json:"break_duration_minutes"`
	Status               string    `json:"status"`
	ClockInAt            time.Time `json:"clock_in_at"`
	ClockOutAt           time.Time `json:"clock_out_at"`
	HoursWorked          float64   `json:"hours_worked"`
	OvertimeHours        float64   `json:"overtime_hours"`
	Notes                string    `json:"notes"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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
	ID         int64     `json:"id"`
	OwnerID    int64     `json:"owner_id"`
	BusinessID int64     `json:"business_id"`
	EmployeeID int64     `json:"employee_id"`
	ShiftID    int64     `json:"shift_id"`
	Date       string    `json:"date"`
	Status     string    `json:"status"`
	Published  bool      `json:"published"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
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
	EmployeeID   int64  `json:"employee_id"`
	Date         string `json:"date"`
	ConflictType string `json:"conflict_type"`
	Description  string `json:"description"`
}
