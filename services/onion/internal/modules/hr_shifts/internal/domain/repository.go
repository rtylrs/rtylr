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
}

type LeaveConflictChecker interface {
	HasApprovedLeave(ctx context.Context, businessID, employeeID int64, date string) (bool, error)
}
