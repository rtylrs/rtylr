package workspaceadapters

import (
	"context"

	flowworkspaces "github.com/voxire/onion/internal/modules/flow/workspaces"
	hrcore "github.com/voxire/onion/internal/modules/hr/hr_core"
	team "github.com/voxire/onion/internal/modules/organization/team"
)

type flowWorkdayRosterAdapter struct {
	core *hrcore.Module
	team *team.Module
}

func (a flowWorkdayRosterAdapter) ResolveWorkdayRoster(ctx context.Context, businessID int64, candidateUserIDs []int64, includeAll bool) (flowworkspaces.WorkdayRoster, error) {
	candidateIDs := make(map[int64]struct{}, len(candidateUserIDs))
	for _, userID := range candidateUserIDs {
		candidateIDs[userID] = struct{}{}
	}
	include := func(userID int64) bool {
		if includeAll {
			return true
		}
		_, ok := candidateIDs[userID]
		return ok
	}

	memberIDs, err := a.team.ListActiveMemberUserIDs(ctx, businessID)
	if err != nil {
		return flowworkspaces.WorkdayRoster{}, err
	}
	roster := flowworkspaces.WorkdayRoster{EmployeeUserIDs: make(map[int64]int64)}
	for _, userID := range memberIDs {
		roster.BusinessMemberUserIDs = append(roster.BusinessMemberUserIDs, userID)
		if include(userID) {
			roster.ActiveMemberUserIDs = append(roster.ActiveMemberUserIDs, userID)
		}
	}
	employees, err := a.core.ListActiveEmployees(ctx, businessID)
	if err != nil {
		return flowworkspaces.WorkdayRoster{}, err
	}
	for _, employee := range employees {
		if include(employee.UserID) {
			roster.EmployeeUserIDs[employee.ID] = employee.UserID
		}
	}
	return roster, nil
}

func NewWorkdayRosterSource(core *hrcore.Module, members *team.Module) flowworkspaces.WorkdayRosterSource {
	return flowWorkdayRosterAdapter{core: core, team: members}
}
