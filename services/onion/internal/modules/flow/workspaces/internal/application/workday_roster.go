package application

import (
	"context"
	"errors"
)

var ErrWorkdayRosterUnavailable = errors.New("workday roster source is unavailable")

// WorkdayRoster carries only the identity links needed to resolve Flow's
// legacy user-keyed days and HR-keyed days. HR and organization own the source
// records; Flow owns the work-day query and visibility policy.
type WorkdayRoster struct {
	ActiveMemberUserIDs   []int64
	BusinessMemberUserIDs []int64
	EmployeeUserIDs       map[int64]int64
}

// WorkdayRosterSource resolves directory identities for a Flow work-day read.
type WorkdayRosterSource interface {
	ResolveWorkdayRoster(ctx context.Context, businessID int64, candidateUserIDs []int64, includeAll bool) (WorkdayRoster, error)
}

func (r WorkdayRoster) EmployeeIDs() []int64 {
	ids := make([]int64, 0, len(r.EmployeeUserIDs))
	for id := range r.EmployeeUserIDs {
		if id > 0 {
			ids = append(ids, id)
		}
	}
	return ids
}

func (r WorkdayRoster) EmployeeIDForUser(userID int64) int64 {
	for employeeID, linkedUserID := range r.EmployeeUserIDs {
		if linkedUserID == userID {
			return employeeID
		}
	}
	return 0
}

func (r WorkdayRoster) HiddenEmployeeIDs() []int64 {
	activeMembers := make(map[int64]struct{}, len(r.BusinessMemberUserIDs))
	for _, userID := range r.BusinessMemberUserIDs {
		activeMembers[userID] = struct{}{}
	}
	visibleMembers := make(map[int64]struct{}, len(r.ActiveMemberUserIDs))
	for _, userID := range r.ActiveMemberUserIDs {
		visibleMembers[userID] = struct{}{}
	}
	var ids []int64
	for employeeID := range r.EmployeeUserIDs {
		if _, isActiveMember := activeMembers[employeeID]; !isActiveMember {
			continue
		}
		if _, isVisibleMember := visibleMembers[employeeID]; !isVisibleMember {
			ids = append(ids, employeeID)
		}
	}
	return ids
}

// UserIDForDay distinguishes a self-recorded user day from an HR employee day
// when their numeric identifiers happen to overlap.
func (r WorkdayRoster) UserIDForDay(employeeID, loggedBy int64) (int64, bool) {
	if employeeID == loggedBy {
		for _, userID := range r.ActiveMemberUserIDs {
			if userID == employeeID {
				return userID, true
			}
		}
		for _, userID := range r.BusinessMemberUserIDs {
			if userID == employeeID {
				return 0, false
			}
		}
	}
	userID, ok := r.EmployeeUserIDs[employeeID]
	return userID, ok
}
