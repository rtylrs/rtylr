package domain

import "context"

type Repository interface {
	ListShifts(ctx context.Context, ownerID int64, params ListShiftsParams) ([]ShiftRecord, int64, error)
	GetShift(ctx context.Context, ownerID, shiftID int64) (*ShiftRecord, error)
	CreateShift(ctx context.Context, ownerID int64, shift ShiftRecord) (*ShiftRecord, error)
	UpdateShift(ctx context.Context, ownerID int64, shift ShiftRecord) error
	ListShiftsByLocation(ctx context.Context, ownerID, locationID int64) ([]ShiftRecord, error)

	ListRosterAssignments(ctx context.Context, ownerID int64, weekStart, weekEnd string) ([]ShiftAssignmentRecord, error)
	CreateShiftAssignment(ctx context.Context, ownerID int64, assignment ShiftAssignmentRecord) (*ShiftAssignmentRecord, error)
	PublishRosterAssignments(ctx context.Context, ownerID int64, weekStart, weekEnd string) error
	ListEmployeeAssignmentsForDate(ctx context.Context, ownerID, employeeID int64, date string) ([]ShiftAssignmentRecord, error)
}

type LeaveConflictChecker interface {
	HasApprovedLeave(ctx context.Context, ownerID, employeeID int64, date string) (bool, error)
}
