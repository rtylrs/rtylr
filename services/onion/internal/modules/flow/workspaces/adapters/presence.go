package workspaceadapters

import (
	"context"

	"github.com/voxire/onion/internal/modules/flow/workspaces"
	hrcore "github.com/voxire/onion/internal/modules/hr/hr_core"
	hrleave "github.com/voxire/onion/internal/modules/hr/hr_leave"
)

type flowPresenceAdapter struct {
	core  *hrcore.Module
	leave *hrleave.Module
}

func (a flowPresenceAdapter) EmployeeIDByUserID(ctx context.Context, businessID, userID int64) (int64, error) {
	return a.core.GetEmployeeIDByUserID(ctx, businessID, userID)
}

func (a flowPresenceAdapter) HasPermission(ctx context.Context, businessID, employeeID int64, permission string) (bool, error) {
	return a.core.HasPermission(ctx, businessID, employeeID, permission)
}

func (a flowPresenceAdapter) ListActiveEmployees(ctx context.Context, businessID int64) ([]workspaces.PresenceEmployee, error) {
	employees, err := a.core.ListActiveEmployees(ctx, businessID)
	if err != nil {
		return nil, err
	}
	result := make([]workspaces.PresenceEmployee, 0, len(employees))
	for _, employee := range employees {
		result = append(result, workspaces.PresenceEmployee{
			ID:        employee.ID,
			UserID:    employee.UserID,
			FirstName: employee.FirstName,
			LastName:  employee.LastName,
		})
	}
	return result, nil
}

func (a flowPresenceAdapter) ApprovedLeaveEmployeeIDs(ctx context.Context, businessID int64, date string) (map[int64]bool, error) {
	return a.leave.ApprovedLeaveEmployeeIDs(ctx, businessID, date)
}

func NewPresenceSource(core *hrcore.Module, leave *hrleave.Module) workspaces.PresenceSource {
	return flowPresenceAdapter{core: core, leave: leave}
}
