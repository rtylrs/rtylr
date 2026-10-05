package teamadapters

import (
	"context"

	hrcore "github.com/voxire/onion/internal/modules/hr/hr_core"
	"github.com/voxire/onion/internal/modules/hr/staff"
	"github.com/voxire/onion/internal/modules/organization/team"
)

type teamHRProvisionerAdapter struct {
	core  *hrcore.Module
	staff *staff.Module
}

func (a teamHRProvisionerAdapter) EnsureEmployeeForUser(ctx context.Context, ownerID, businessID, userID int64, firstName, lastName, email string) error {
	return a.core.EnsureEmployeeForUser(ctx, ownerID, businessID, userID, firstName, lastName, email)
}

func (a teamHRProvisionerAdapter) EnsureStaffForUser(ctx context.Context, ownerID, businessID, userID int64, role string) error {
	return a.staff.EnsureStaffForUser(ctx, ownerID, businessID, userID, role)
}

func (a teamHRProvisionerAdapter) DeleteProvisionedMember(ctx context.Context, ownerID, businessID, userID int64) error {
	if err := a.staff.DeleteStaffForUser(ctx, ownerID, businessID, userID); err != nil {
		return err
	}
	return a.core.DeleteEmployeeForUser(ctx, ownerID, businessID, userID)
}

func (a teamHRProvisionerAdapter) ListTerminatedEmployees(ctx context.Context) ([]team.TerminatedEmployee, error) {
	rows, err := a.core.ListTerminatedEmployees(ctx)
	if err != nil {
		return nil, err
	}
	employees := make([]team.TerminatedEmployee, 0, len(rows))
	for _, row := range rows {
		employees = append(employees, team.TerminatedEmployee{BusinessID: row.BusinessID, UserID: row.UserID})
	}
	return employees, nil
}

func NewHRMemberProvisioner(core *hrcore.Module, staffModule *staff.Module) team.MemberProvisioner {
	return teamHRProvisionerAdapter{core: core, staff: staffModule}
}

func NewTerminatedEmployeeSource(core *hrcore.Module, staffModule *staff.Module) team.TerminatedEmployeeSource {
	return teamHRProvisionerAdapter{core: core, staff: staffModule}
}
