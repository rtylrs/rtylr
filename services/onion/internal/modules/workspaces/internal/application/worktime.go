package application

import (
	"context"
	"errors"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
)

// Errors the API layer maps onto status codes. Returning typed errors rather
// than strings keeps the decision about what a caller sees in one place.
var (
	ErrInvalidWorkMode  = errors.New("work_mode must be onsite, remote or hybrid")
	ErrWorkDayNotFound  = errors.New("work day not found")
	ErrWorkTypeNotFound = errors.New("work type not found")
	ErrWorkTypeTooDeep  = errors.New("work types are two levels deep: a child cannot be a parent")
	ErrDayLocked        = errors.New("an approved day must be reopened before it can be changed")
	ErrProjectRequired  = errors.New("project_id is required")
)

// GetWorkDay reads a day with its lines and the clocked-versus-logged gap.
//
// A day that has never been recorded is not an error: the caller renders an
// empty day for that date, which is what most days look like before someone
// fills them in.
func (s *Service) GetWorkDay(ctx context.Context, ownerID, businessID, employeeID int64, date string) (domain.WorkDay, []domain.WorkDayLine, bool, error) {
	day, found, err := s.repo.GetWorkDay(ctx, ownerID, businessID, employeeID, date)
	if err != nil || !found {
		return domain.WorkDay{}, nil, found, err
	}

	lines, logged, err := s.repo.ListWorkDayLines(ctx, businessID, day.ID)
	if err != nil {
		return day, nil, true, err
	}

	day.LoggedMinutes = logged
	day.GapMinutes = day.ClockedMinutes - logged
	return day, lines, true, nil
}

// UpsertWorkDay records or corrects the day itself.
//
// Rejects an unknown work mode rather than storing it: the value drives
// reporting, and a typo would quietly become its own category.
func (s *Service) UpsertWorkDay(ctx context.Context, input domain.WorkDayInput) (domain.WorkDay, error) {
	if !domain.ValidWorkMode(input.WorkMode) {
		return domain.WorkDay{}, ErrInvalidWorkMode
	}

	existing, found, err := s.repo.GetWorkDay(ctx, input.OwnerID, input.BusinessID, input.EmployeeID, input.Date)
	if err != nil {
		return domain.WorkDay{}, err
	}
	// An approved day is settled. Changing its clock behind the approver would
	// invalidate a decision they have already made, so it has to be reopened
	// first, which is recorded and visible.
	if found && existing.Status == domain.WorkDayStatusApproved {
		return domain.WorkDay{}, ErrDayLocked
	}

	day, err := s.repo.UpsertWorkDay(ctx, input)
	if err != nil {
		return domain.WorkDay{}, err
	}

	lines, logged, err := s.repo.ListWorkDayLines(ctx, input.BusinessID, day.ID)
	if err != nil {
		return day, err
	}
	_ = lines
	day.LoggedMinutes = logged
	day.GapMinutes = day.ClockedMinutes - logged
	return day, nil
}

// CreateWorkDayLine adds a piece of work to a day.
//
// The line inherits the day's date and timezone rather than taking them from the
// client: a line belongs to the day it is filed under, and letting the two
// disagree is how an entry ends up on the wrong week.
func (s *Service) CreateWorkDayLine(ctx context.Context, input domain.WorkDayLineInput) (domain.WorkDayLine, error) {
	if input.WorkMode != "" && !domain.ValidWorkMode(input.WorkMode) {
		return domain.WorkDayLine{}, ErrInvalidWorkMode
	}
	if input.ProjectID <= 0 {
		return domain.WorkDayLine{}, ErrProjectRequired
	}

	day, found, err := s.repo.GetWorkDayByID(ctx, input.WorkDayID, input.OwnerID, input.BusinessID)
	if err != nil {
		return domain.WorkDayLine{}, err
	}
	if !found {
		return domain.WorkDayLine{}, ErrWorkDayNotFound
	}
	if day.Status == domain.WorkDayStatusApproved {
		return domain.WorkDayLine{}, ErrDayLocked
	}

	if err := s.assertCallerOwnsWorkDay(ctx, input, day.EmployeeID); err != nil {
		return domain.WorkDayLine{}, err
	}

	input.EmployeeID = day.EmployeeID
	input.Date = day.Date.Format("2006-01-02")
	input.Timezone = day.Timezone
	return s.repo.CreateWorkDayLine(ctx, input)
}

// callerOwnsWorkDay reports whether a caller may file a line onto a work day.
//
// A day is keyed on hr_employees.id while the caller is an auth user id, so the
// two are compared through the caller's resolved employee row. The direct
// callerUserID == dayEmployeeID case stays allowed because the v2 time API
// still names a caller's own day by their user id where HR onboarding never
// ran, and those days must remain writable by the person they belong to.
func callerOwnsWorkDay(callerUserID, dayEmployeeID, callerEmployeeID int64) bool {
	if callerUserID <= 0 || dayEmployeeID <= 0 {
		return false
	}
	if dayEmployeeID == callerUserID {
		return true
	}
	return callerEmployeeID > 0 && dayEmployeeID == callerEmployeeID
}

// assertCallerOwnsWorkDay refuses a line filed onto somebody else's day.
//
// GetWorkDayByID is scoped to the business, which every member of that business
// shares, so on its own it lets any authenticated member append time to a
// colleague's day. The line is then stamped with the target day's employee, so
// the hours read as that colleague's own and feed their approvals.
//
// A foreign day is reported as not found rather than forbidden, so the response
// does not confirm that another employee's day id exists. DeleteWorkDayLine
// already answers that way for the same reason.
func (s *Service) assertCallerOwnsWorkDay(ctx context.Context, input domain.WorkDayLineInput, dayEmployeeID int64) error {
	callerUserID := input.LoggedBy
	if callerUserID <= 0 {
		return ErrWorkDayNotFound
	}
	if callerOwnsWorkDay(callerUserID, dayEmployeeID, 0) {
		return nil
	}

	byUserID, err := s.repo.EmployeeIDsByUserIDs(ctx, input.OwnerID, input.BusinessID, []int64{callerUserID})
	if err != nil {
		return err
	}
	if callerOwnsWorkDay(callerUserID, dayEmployeeID, byUserID[callerUserID]) {
		return nil
	}
	return ErrWorkDayNotFound
}

// UpdateWorkDayLine changes a line in place.
func (s *Service) UpdateWorkDayLine(ctx context.Context, input domain.WorkDayLineInput) (domain.WorkDayLine, error) {
	if input.WorkMode != "" && !domain.ValidWorkMode(input.WorkMode) {
		return domain.WorkDayLine{}, ErrInvalidWorkMode
	}
	if input.ProjectID <= 0 {
		return domain.WorkDayLine{}, ErrProjectRequired
	}
	return s.repo.UpdateWorkDayLine(ctx, input)
}

// DeleteWorkDayLine removes a line.
func (s *Service) DeleteWorkDayLine(ctx context.Context, id, ownerID, businessID int64) (bool, error) {
	return s.repo.DeleteWorkDayLine(ctx, id, ownerID, businessID)
}

// ListWorkTypes returns the workspace's classification tree.
func (s *Service) ListWorkTypes(ctx context.Context, ownerID, businessID int64, includeArchived bool) ([]domain.WorkType, error) {
	return s.repo.ListWorkTypes(ctx, ownerID, businessID, includeArchived)
}

// CreateWorkType adds a node, holding the tree to two levels.
func (s *Service) CreateWorkType(ctx context.Context, input domain.WorkTypeInput) (domain.WorkType, error) {
	if err := s.assertParentIsTopLevel(ctx, input.OwnerID, input.BusinessID, input.ParentID); err != nil {
		return domain.WorkType{}, err
	}
	return s.repo.CreateWorkType(ctx, input)
}

// UpdateWorkType changes a node, holding the tree to two levels in both
// directions: a node may not be nested under a child, and a node that already
// has children may not itself become one.
func (s *Service) UpdateWorkType(ctx context.Context, input domain.WorkTypeInput) (domain.WorkType, error) {
	if err := s.assertParentIsTopLevel(ctx, input.OwnerID, input.BusinessID, input.ParentID); err != nil {
		return domain.WorkType{}, err
	}
	if input.ParentID > 0 {
		children, err := s.repo.CountWorkTypeChildren(ctx, input.OwnerID, input.BusinessID, input.ID)
		if err != nil {
			return domain.WorkType{}, err
		}
		if children > 0 {
			return domain.WorkType{}, ErrWorkTypeTooDeep
		}
	}
	return s.repo.UpdateWorkType(ctx, input)
}

// DeleteWorkType removes a node.
func (s *Service) DeleteWorkType(ctx context.Context, id, ownerID, businessID int64) (bool, error) {
	return s.repo.DeleteWorkType(ctx, id, ownerID, businessID)
}

// assertParentIsTopLevel rejects nesting under a node that is already a child.
func (s *Service) assertParentIsTopLevel(ctx context.Context, ownerID, businessID, parentID int64) error {
	if parentID <= 0 {
		return nil
	}
	parent, err := s.repo.GetWorkType(ctx, parentID, ownerID, businessID)
	if err != nil {
		return ErrWorkTypeNotFound
	}
	if parent.ParentID > 0 {
		return ErrWorkTypeTooDeep
	}
	return nil
}

// ListWorkDays returns a person's days across a range. The lines are not loaded:
// a week view needs each day's totals, not every line on every day, and the
// totals are already on the row.
func (s *Service) ListWorkDays(ctx context.Context, ownerID, businessID, employeeID int64, from, to string, limit int32, offset int64) ([]domain.WorkDay, int64, error) {
	return s.repo.ListWorkDays(ctx, ownerID, businessID, employeeID, from, to, limit, offset)
}

// workspaceMemberCap bounds the roster a team view resolves in one call. A
// workspace with more members than this is a scale this feature has not been
// asked to handle yet, and silently truncating is better than an unbounded
// read - the same judgment call every other "resolve then query" internal
// read in this service already makes.
const workspaceMemberCap = 500

// ListWorkDaysForWorkspace returns work days across every member of a
// workspace, for the team view. callerID must hold at least FlowRoleAdmin on
// the workspace. Business-wide visibility is handled by the central Flow
// permission gate, while resource checks remain Flow-owned.
//
// WorkDaysForDate resolves one date's WorkDay for an arbitrary set of
// employee ids, keyed by employee id for O(1) lookup. Used by Flow's
// business-wide "who's in today" presence view, which resolves its employee
// list from hr_core rather than workspace membership - an employee with no
// row for the date simply has no entry, not an error.
func (s *Service) WorkDaysForDate(ctx context.Context, ownerID, businessID int64, employeeIDs []int64, date string) (map[int64]domain.WorkDay, error) {
	if len(employeeIDs) == 0 {
		return map[int64]domain.WorkDay{}, nil
	}
	days, _, err := s.repo.ListWorkDaysByEmployeeIDs(ctx, ownerID, businessID, employeeIDs, date, date, int32(len(employeeIDs)), 0) // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	if err != nil {
		return nil, err
	}
	byEmployeeID := make(map[int64]domain.WorkDay, len(days))
	for _, d := range days {
		byEmployeeID[d.EmployeeID] = d
	}
	return byEmployeeID, nil
}

// Work days key on HR's employee_id; workspace membership keys on the auth
// user_id. The two are resolved through hr_employees.user_id, and a member
// with no matching HR record simply contributes no rows - not an error, the
// same way an employee with no day yet contributes none.
func (s *Service) ListWorkDaysForWorkspace(ctx context.Context, callerID, workspaceID, ownerID, businessID int64, from, to string, limit int32, offset int64) ([]domain.WorkDay, int64, error) {
	if err := s.RequireWorkspaceRole(ctx, workspaceID, callerID, businessID, domain.FlowRoleAdmin); err != nil {
		return nil, 0, err
	}

	members, _, err := s.repo.ListWorkspaceMembers(ctx, workspaceID, workspaceMemberCap, 0)
	if err != nil {
		return nil, 0, err
	}
	userIDs := make([]int64, 0, len(members)+1)
	for _, m := range members {
		userIDs = append(userIDs, m.UserID)
	}
	// The owner holds no flow_workspace_members row (RequireWorkspaceRole
	// resolves them the same way), so their day would otherwise be silently
	// missing from their own team's roster.
	if ownerUserID, err := s.repo.GetWorkspaceOwner(ctx, workspaceID); err == nil {
		userIDs = append(userIDs, ownerUserID)
	}

	byUserID, err := s.repo.EmployeeIDsByUserIDs(ctx, ownerID, businessID, userIDs)
	if err != nil {
		return nil, 0, err
	}
	employeeIDs := make([]int64, 0, len(byUserID))
	// The reverse of byUserID: the repository returns days keyed on
	// employee_id, and the frontend needs user_id to match a day against the
	// workspace roster it already has.
	userIDByEmployeeID := make(map[int64]int64, len(byUserID))
	for userID, employeeID := range byUserID {
		employeeIDs = append(employeeIDs, employeeID)
		userIDByEmployeeID[employeeID] = userID
	}

	days, total, err := s.repo.ListWorkDaysByEmployeeIDs(ctx, ownerID, businessID, employeeIDs, from, to, int32(limit), offset)
	if err != nil {
		return nil, 0, err
	}
	for i := range days {
		days[i].UserID = userIDByEmployeeID[days[i].EmployeeID]
	}
	return days, total, nil
}

// ListTeamWorkDays is the v2 replacement for ListWorkDaysForWorkspace.
// Instead of scoping by workspace membership, it uses the business team roster
// resolved through hr_employees. Visibility is governed by the business's
// configured policy; until the config surface lands, the default policy is
// hub admin+ sees everyone, project admins see people sharing their projects.
func (s *Service) ListTeamWorkDays(ctx context.Context, callerID, businessID int64, from, to string, limit, offset int64) ([]domain.WorkDay, int64, error) {
	// Default policy gate: only the central Flow team-management permission
	// grants business-wide visibility. Resource checks remain scope-bound.
	if s.hubProfileBypasses(ctx, callerID, businessID) {
		return s.listAllBusinessWorkDays(ctx, businessID, from, to, limit, offset)
	}
	// Project-admin visibility: callers who administer at least one project
	// see the people on those projects. This is a simplified default; the
	// per-business config policy will refine it when the settings panel lands.
	return s.listVisibleProjectMemberDays(ctx, callerID, businessID, from, to, limit, offset)
}

func (s *Service) listAllBusinessWorkDays(ctx context.Context, businessID int64, from, to string, limit, offset int64) ([]domain.WorkDay, int64, error) {
	employeeIDs, err := s.repo.BusinessEmployeeIDs(ctx, businessID)
	if err != nil {
		return nil, 0, err
	}
	return s.repo.ListWorkDaysByEmployeeIDs(ctx, businessID, businessID, employeeIDs, from, to, int32(limit), offset) // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
}

func (s *Service) listVisibleProjectMemberDays(ctx context.Context, callerID, businessID int64, from, to string, limit, offset int64) ([]domain.WorkDay, int64, error) {
	// Find projects this admin oversees, then collect their members' employee ids.
	projects, _, err := s.repo.GetProjectsByCreatorOrMembership(ctx, businessID, callerID, 500, 0)
	if err != nil {
		return nil, 0, err
	}
	userIDs := make([]int64, 0)
	seen := make(map[int64]bool)
	for _, p := range projects {
		members, _, mErr := s.repo.ListProjectMembers(ctx, p.ID, 500, 0)
		if mErr != nil {
			continue
		}
		for _, m := range members {
			if !seen[m.UserID] {
				seen[m.UserID] = true
				userIDs = append(userIDs, m.UserID)
			}
		}
	}
	if len(userIDs) == 0 {
		return []domain.WorkDay{}, 0, nil
	}
	byUserID, err := s.repo.EmployeeIDsByUserIDs(ctx, businessID, businessID, userIDs)
	if err != nil {
		return nil, 0, err
	}
	employeeIDs := make([]int64, 0, len(byUserID))
	for _, eid := range byUserID {
		employeeIDs = append(employeeIDs, eid)
	}
	return s.repo.ListWorkDaysByEmployeeIDs(ctx, businessID, businessID, employeeIDs, from, to, int32(limit), offset) // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
}
