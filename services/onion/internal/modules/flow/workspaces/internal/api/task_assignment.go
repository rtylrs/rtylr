package api

import "context"

// Members may claim their own work, but assigning work to another user remains
// an elevated operation.
func requireTaskAssignmentPermission(ctx context.Context, assigneeID *int64, callerID int64) error {
	if assigneeID == nil {
		return nil
	}
	if callerID != 0 && *assigneeID == callerID {
		return nil
	}
	return requireFlowPermission(ctx, "flow:tasks:assign")
}
