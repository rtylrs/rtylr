package workspaces

import (
	"context"
	"errors"
)

// Presence status values. A day carries one of these, resolved server-side -
// mode and status only, deliberately no minutes: a company-wide view watched
// by every employee, including people with no management relationship to
// each other, is a different exposure than the workspace-admin-gated team
// view, which does show clocked/logged/gap minutes.
const (
	PresenceOnLeave    = "on_leave"
	PresenceNotStarted = "not_started"
)

var errPresenceAccessDenied = errors.New("presence access denied")

type PresenceEntry struct {
	EmployeeID int64  `json:"employee_id"`
	UserID     int64  `json:"user_id"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	// Status is a work mode ("onsite"/"remote"/"hybrid") or one of the
	// PresenceOnLeave/PresenceNotStarted constants above.
	Status string `json:"status"`
}

// ListPresenceForBusiness resolves every active employee's presence for one
// date, gated by view_directory - the same permission Directory uses,
// because this answers the same underlying question ("who works here") from
// a different angle.
//
// The three ids passed to different modules are deliberately not the same
// value: hr_core/staff resolve "ownerID" to the tenant's businessID (see
// canViewDirectory), while hr_leave's own handlers pass the caller's raw hub
// user id as what it calls "ownerID" - every leave request already on disk
// was written under that convention, so querying it any other way would
// silently return nothing rather than fail loudly.
func (m *Module) ListPresenceForBusiness(
	ctx context.Context,
	resolvedOwnerID, businessID, callerUserID int64,
	date string,
) ([]PresenceEntry, error) {
	employeeID, err := m.hrCore.GetEmployeeIDByUserID(ctx, businessID, callerUserID)
	if err != nil || employeeID <= 0 {
		return nil, errPresenceAccessDenied
	}
	allowed, err := m.hrCore.HasPermission(ctx, businessID, employeeID, "view_directory")
	if err != nil || !allowed {
		return nil, errPresenceAccessDenied
	}

	employees, err := m.hrCore.ListActiveEmployees(ctx, businessID)
	if err != nil {
		return nil, err
	}
	employeeIDs := make([]int64, 0, len(employees))
	for _, e := range employees {
		employeeIDs = append(employeeIDs, e.ID)
	}

	workDays, err := m.service.WorkDaysForDate(ctx, resolvedOwnerID, businessID, employeeIDs, date)
	if err != nil {
		return nil, err
	}
	onLeave, err := m.hrLeave.ApprovedLeaveEmployeeIDs(ctx, callerUserID, date)
	if err != nil {
		return nil, err
	}

	entries := make([]PresenceEntry, 0, len(employees))
	for _, e := range employees {
		status := PresenceNotStarted
		if day, ok := workDays[e.ID]; ok && day.WorkMode != "" {
			status = day.WorkMode
		}
		// On-leave overrides a work mode a day may already carry: someone who
		// clocked in earlier and then had time off approved for today is on
		// leave today, not "onsite."
		if onLeave[e.ID] {
			status = PresenceOnLeave
		}
		entries = append(entries, PresenceEntry{
			EmployeeID: e.ID, UserID: e.UserID, FirstName: e.FirstName, LastName: e.LastName, Status: status,
		})
	}
	return entries, nil
}
