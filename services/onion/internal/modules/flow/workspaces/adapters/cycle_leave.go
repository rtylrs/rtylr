package workspaceadapters

import (
	"context"
	"time"

	flowworkspaces "github.com/voxire/onion/internal/modules/flow/workspaces"
	hrcore "github.com/voxire/onion/internal/modules/hr/hr_core"
	hrleave "github.com/voxire/onion/internal/modules/hr/hr_leave"
)

// flowCycleLeaveAdapter answers Flow's cycle capacity question, which dates is
// each user on approved leave, from HR's own module APIs. Flow keys people by
// user id and HR keys leave by employee id; hr_core's active employee list is
// the link, as it is for presence.
type flowCycleLeaveAdapter struct {
	core  *hrcore.Module
	leave *hrleave.Module
}

func (a flowCycleLeaveAdapter) ApprovedLeaveDates(ctx context.Context, businessID int64, userIDs []int64, from, to time.Time) (map[int64]map[string]bool, error) {
	wanted := make(map[int64]bool, len(userIDs))
	for _, id := range userIDs {
		wanted[id] = true
	}
	employees, err := a.core.ListActiveEmployees(ctx, businessID)
	if err != nil {
		return nil, err
	}
	userByEmployee := make(map[int64]int64, len(employees))
	for _, e := range employees {
		if wanted[e.UserID] {
			userByEmployee[e.ID] = e.UserID
		}
	}
	out := make(map[int64]map[string]bool, len(userIDs))
	if len(userByEmployee) == 0 {
		return out, nil
	}
	ranges, err := a.leave.ApprovedLeaveRanges(ctx, businessID, from.Format("2006-01-02"), to.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	for _, r := range ranges {
		userID, ok := userByEmployee[r.EmployeeID]
		if !ok {
			continue
		}
		start, startErr := time.Parse("2006-01-02", leaveDate(r.StartDate))
		end, endErr := time.Parse("2006-01-02", leaveDate(r.EndDate))
		if startErr != nil || endErr != nil {
			continue
		}
		if start.Before(from) {
			start = from
		}
		if end.After(to) {
			end = to
		}
		if out[userID] == nil {
			out[userID] = map[string]bool{}
		}
		for day := start; !day.After(end); day = day.AddDate(0, 0, 1) {
			out[userID][day.Format("2006-01-02")] = true
		}
	}
	return out, nil
}

// leaveDate keeps the date part of a stored value that may carry a time.
func leaveDate(value string) string {
	if len(value) > 10 {
		return value[:10]
	}
	return value
}

func NewCycleLeaveSource(core *hrcore.Module, leave *hrleave.Module) flowworkspaces.CycleLeaveSource {
	return flowCycleLeaveAdapter{core: core, leave: leave}
}
