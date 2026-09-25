package domain

import "context"

type LeaveConflictChecker interface {
	HasApprovedLeave(ctx context.Context, businessID, employeeID int64, date string) (bool, error)
}
