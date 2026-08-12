package domain

import "context"

type Repository interface {
	ListShifts(ctx context.Context, businessID int64, params ListShiftsParams) ([]ShiftRecord, int64, error)
	GetShift(ctx context.Context, businessID, shiftID int64) (*ShiftRecord, error)
	CreateShift(ctx context.Context, businessID int64, shift ShiftRecord) (*ShiftRecord, error)
	UpdateShift(ctx context.Context, businessID int64, shift ShiftRecord) error
	DeleteShift(ctx context.Context, businessID, shiftID int64) error
	ListShiftsByLocation(ctx context.Context, businessID, locationID int64) ([]ShiftRecord, error)

	ListRosterAssignments(ctx context.Context, businessID int64, weekStart, weekEnd string) ([]ShiftAssignmentRecord, error)
	CreateShiftAssignment(ctx context.Context, businessID int64, assignment ShiftAssignmentRecord) (*ShiftAssignmentRecord, error)
	PublishRosterAssignments(ctx context.Context, businessID int64, weekStart, weekEnd string) error
	ListEmployeeAssignmentsForDate(ctx context.Context, businessID, employeeID int64, date string) ([]ShiftAssignmentRecord, error)

	ListAvailability(ctx context.Context, businessID int64, params ListAvailabilityParams) ([]AvailabilityRecord, int64, error)
	GetAvailability(ctx context.Context, businessID, id int64) (*AvailabilityRecord, error)
	UpsertAvailability(ctx context.Context, businessID int64, entry AvailabilityRecord) (*AvailabilityRecord, error)
	DeleteAvailability(ctx context.Context, businessID, id int64) error
	// ListAvailabilityForStaffDate returns the rows that govern one staff member
	// on one date: the dated exceptions for that date plus the recurring rules
	// for its weekday. Callers resolve precedence, since an exception overrides
	// the pattern.
	ListAvailabilityForStaffDate(ctx context.Context, businessID, staffID int64, date string, dayOfWeek int32) ([]AvailabilityRecord, error)
}

type LeaveConflictChecker interface {
	HasApprovedLeave(ctx context.Context, businessID, employeeID int64, date string) (bool, error)
}
