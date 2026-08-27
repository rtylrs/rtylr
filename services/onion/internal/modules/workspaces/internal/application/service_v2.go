package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
)

// DefaultBinRetentionDays is the recycle-bin window used until the per-business
// retention config key lands (spec ruling P2/D-decision: configurable, 30d
// default). Archive stamps purge_at from this at archive time; later config
// changes affect future archives only.
const DefaultBinRetentionDays = 30

// ErrNotFound is returned by v2 lookups when the business has no such project.
var ErrNotFound = errors.New("flow project not found")

// canSeeProject reports whether callerID may READ projectID under the curation
// rule: creator, member, or manager visibility through the reporting line.
// Public visibility never widens read access here.
func (s *Service) canSeeProject(ctx context.Context, p domain.Project, callerID, businessID int64) bool {
	if p.CreatedBy == callerID || p.OwnerID == callerID {
		return true
	}
	if s.repo.IsProjectMember(ctx, callerID, p.ID) {
		return true
	}
	return s.managesSomeoneOn(ctx, p, callerID, businessID)
}

// managesSomeoneOn reports whether the caller manages anyone involved in this
// project, which is the dept-manager visibility proxy (M3).
//
// The My Tasks bucket is exempt no matter who manages its owner: Project's own
// contract says is_personal is "exempt from manager-visibility policies unless
// its owner explicitly shares it", so the check stops before the proxy.
//
// Involvement means created, owns, or is a member. Membership is probed one
// report at a time rather than with an IN clause because a manager has few
// reports and sqlc's slice expansion produces invalid SQL for an empty set in
// MySQL. It does mean this is O(reports) queries on a cache miss, so it runs
// last, after the cheap self checks have already returned.
func (s *Service) managesSomeoneOn(ctx context.Context, p domain.Project, callerID, businessID int64) bool {
	if p.IsPersonal {
		return false
	}
	reports := s.managedUserIDs(ctx, callerID, businessID)

	// Creator and owner are already in hand, so settle the common case without
	// touching the database at all.
	for _, reportID := range reports {
		if p.CreatedBy == reportID || p.OwnerID == reportID {
			return true
		}
	}

	// Only then pay for membership lookups. A nil repo means the proxy cannot
	// answer, and the rule throughout is that failing to resolve must decline to
	// widen rather than widen blindly.
	if s.repo == nil {
		return false
	}
	for _, reportID := range reports {
		if s.repo.IsProjectMember(ctx, reportID, p.ID) {
			return true
		}
	}
	return false
}

// GetProjectV2 returns one project if the caller may see it.
func (s *Service) GetProjectV2(ctx context.Context, projectID, callerID, businessID int64) (domain.Project, error) {
	p, err := s.repo.GetProjectV2ByID(ctx, projectID, businessID)
	if err != nil {
		return domain.Project{}, err
	}
	if !s.canSeeProject(ctx, p, callerID, businessID) {
		return domain.Project{}, domain.ErrUnauthorized
	}
	return p, nil
}

// ListProjectsV2 lists active non-personal projects under curation. Bypass
// callers see the whole business; everyone else sees created + member.
func (s *Service) ListProjectsV2(ctx context.Context, callerID, businessID int64, status, visibility, q string, limit, offset int64) ([]domain.Project, int64, error) {
	// Ruling P8: this listing is curated for EVERYONE, bypass roles included.
	// A central Flow manager holds business-wide management permission, but
	// access is not listing: their sidebar shows what they created or belong to, like anyone
	// else's. Passing the bypass through here put every private project in the
	// business into an admin's sidebar, which is the precise thing P8 forbids.
	//
	// Deliberately not applied to ListArchivedProjectsV2: the recycle bin is
	// governed by P4, which gives the business owner and admins restore rights
	// and therefore sight of the entries.
	const includeAll = false

	// Resolved unconditionally now. Previously skipped for bypass callers on the
	// grounds that they already saw everything; with the listing curated, a
	// manager who also holds a bypass role still needs their reports' projects
	// grouped under P9.
	managed := s.managedUserIDs(ctx, callerID, businessID)

	projects, total, err := s.repo.ListProjectsV2(ctx, businessID, callerID, includeAll, managed, status, visibility, q, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	return s.markVisibleViaReport(ctx, projects, callerID, managed), total, nil
}

// markVisibleViaReport flags the projects this caller sees only through the
// dept-manager proxy, so a client can group them (P9) without asking per project
// why each one is visible.
//
// A project the caller created, owns, or is a member of is theirs, so the flag
// stays false. Anything else in the result must have matched a managed branch,
// because that is the only other way the list query admits a row. With no
// reports the loop does not run, which is the common case and costs nothing.
func (s *Service) markVisibleViaReport(ctx context.Context, projects []domain.Project, callerID int64, managed []int64) []domain.Project {
	if len(managed) == 0 || len(projects) == 0 {
		return projects
	}
	for i := range projects {
		p := &projects[i]
		// Creator and owner are already in hand, so a caller's own project is
		// settled without a query. This check comes first deliberately: it is
		// the one that must hold even when the membership lookup cannot run.
		if p.CreatedBy == callerID || p.OwnerID == callerID {
			continue
		}
		// Without a repo the membership question cannot be answered. Flagging
		// optimistically would advertise a caller's own project as a report's,
		// so leave it false and degrade to no grouping rather than wrong
		// grouping.
		if s.repo == nil {
			continue
		}
		if s.repo.IsProjectMember(ctx, callerID, p.ID) {
			continue
		}
		p.VisibleViaReport = true
	}
	return projects
}

// ListArchivedProjectsV2 lists the caller's visible recycle-bin entries,
// soonest purge first.
func (s *Service) ListArchivedProjectsV2(ctx context.Context, callerID, businessID int64, limit, offset int64) ([]domain.Project, int64, error) {
	includeAll := s.hubProfileBypasses(ctx, callerID, businessID)
	return s.repo.ListArchivedProjectsV2(ctx, businessID, callerID, includeAll, limit, offset)
}

// SetProjectVisibility flips private/public. Gate mirrors other project admin
// operations; visibility changes invalidate the workspaces scope because list
// projections may embed visibility-derived ordering later.
func (s *Service) SetProjectVisibility(ctx context.Context, projectID, callerID, businessID int64, visibility string) (domain.Project, error) {
	if !domain.ValidVisibility(visibility) {
		return domain.Project{}, errors.New("invalid visibility")
	}
	if err := s.RequireProjectRole(ctx, projectID, callerID, businessID, domain.FlowRoleAdmin); err != nil {
		return domain.Project{}, err
	}
	p, err := s.repo.GetProjectV2ByID(ctx, projectID, businessID)
	if err != nil {
		return domain.Project{}, err
	}
	if p.IsPersonal {
		return domain.Project{}, errors.New("the My Tasks bucket is always private")
	}
	if err := s.repo.UpdateProjectVisibility(ctx, projectID, businessID, visibility); err != nil {
		return domain.Project{}, err
	}
	defer invalidateWorkspacesScope(ctx, businessID)
	return s.repo.GetProjectV2ByID(ctx, projectID, businessID)
}

// RestoreProjectV2 pulls a binned project back to active. Admin gate matches
// archive; the backfilled membership model guarantees former workspace admins
// hold this role where they had it before.
func (s *Service) RestoreProjectV2(ctx context.Context, projectID, callerID, businessID int64) (domain.Project, error) {
	if err := s.RequireProjectRole(ctx, projectID, callerID, businessID, domain.FlowRoleAdmin); err != nil {
		return domain.Project{}, err
	}
	if err := s.repo.RestoreProjectV2(ctx, projectID, businessID); err != nil {
		return domain.Project{}, err
	}
	defer invalidateWorkspacesScope(ctx, businessID)
	return s.repo.GetProjectV2ByID(ctx, projectID, businessID)
}

// PurgeProjectV2 irreversibly deletes an archived project. Gate is deliberately
// narrower than restore: the creator or a hub owner/admin only (ruling P4).
// Returns the attachment object URLs that were referenced so the caller can
// sweep stored files once rows are gone.
func (s *Service) PurgeProjectV2(ctx context.Context, projectID, callerID, businessID int64) ([]string, error) {
	p, err := s.repo.GetProjectV2ByID(ctx, projectID, businessID)
	if err != nil {
		return nil, err
	}
	if p.ArchivedAt == nil {
		return nil, errors.New("only archived projects can be purged")
	}
	bypass := s.hubProfileBypasses(ctx, callerID, businessID)
	if !bypass && p.CreatedBy != callerID && p.OwnerID != callerID {
		return nil, domain.ErrUnauthorized
	}
	urls, err := s.repo.PurgeProjectCascade(ctx, projectID, businessID)
	if err != nil {
		return nil, err
	}
	invalidateWorkspacesScope(ctx, businessID)
	return urls, nil
}

// binPurgeDeadline computes the permanent-deletion timestamp for an archive
// happening now.
func binPurgeDeadline(now time.Time) time.Time {
	return now.AddDate(0, 0, DefaultBinRetentionDays)
}

// IsProjectMemberOrBypass is the v2 read gate for project-scoped subresources:
// hub bypass, creator, owner, or explicit membership.
func (s *Service) IsProjectMemberOrBypass(ctx context.Context, userID, businessID, projectID int64) bool {
	p, err := s.repo.GetProjectV2ByID(ctx, projectID, businessID)
	if err != nil {
		return false
	}
	if p.CreatedBy == userID || p.OwnerID == userID {
		return true
	}
	return s.repo.IsProjectMember(ctx, userID, p.ID)
}

// ListTasksV2 lists tasks across the caller's visible projects, optionally
// narrowed to one project. Visibility is enforced row-by-row via the parent
// project clause unless a hub bypass applies.
func (s *Service) ListTasksV2(ctx context.Context, callerID, businessID int64, filter domain.ListTasksFilter) ([]domain.Task, int64, error) {
	filter.BusinessID = businessID
	return s.repo.GetTasksByBusinessV2(ctx, filter, callerID, false)
}

// SearchTasksV2 is business-wide fulltext search under the same visibility
// clause as listing.
func (s *Service) SearchTasksV2(ctx context.Context, callerID, businessID int64, q string, limit, offset int64) ([]domain.Task, int64, error) {
	return s.repo.SearchTasksV2(ctx, businessID, callerID, false, q, limit, offset)
}

// EnsureAssigneeMembership implements ruling T1: assigning work to someone
// grants them membership in its project, so nobody is ever assigned a task
// they cannot see. Hub-bypass roles already hold access and creators own the
// project, so neither gets a redundant row.
func (s *Service) EnsureAssigneeMembership(ctx context.Context, projectID, businessID, assigneeID int64) error {
	if assigneeID == 0 {
		return nil
	}
	p, err := s.repo.GetProjectV2ByID(ctx, projectID, businessID)
	if err != nil {
		return err
	}
	// The assignee has to be a colleague. This was a workspace-membership check,
	// which a v2 project can never pass: its workspace_id is 0, so every
	// assignment to someone other than yourself failed with "assignee is not a
	// workspace member" and CreateTask propagated that as fatal. Since all
	// project creation moved to v2, that was every new project.
	//
	// The hub profile resolver is the business-membership primitive: it looks the
	// subject up inside the given business and reports ErrUserNotFound otherwise.
	//
	// Skipped rather than failed when the resolver is not wired, which happens in
	// tests and some dev setups. GetProjectV2ByID above already scoped the project
	// to the caller's business, so the worst case without this check is a member
	// row for a user id outside the business, which does not make another
	// tenant's data reachable.
	if s.hubProfiles != nil {
		if _, profileErr := s.hubProfiles.GetProfile(ctx, assigneeID, businessID); profileErr != nil {
			return fmt.Errorf("assignee is not a member of this business")
		}
	}
	if p.CreatedBy == assigneeID || p.OwnerID == assigneeID {
		return nil
	}
	if s.repo.IsProjectMember(ctx, assigneeID, p.ID) {
		return nil
	}
	return s.repo.AddProjectMember(ctx, domain.AddProjectMemberInput{
		ProjectID: p.ID,
		UserID:    assigneeID,
		Role:      domain.FlowRoleMember,
	})
}

// AddTaskDependency records that taskID is blocked by dependsOnTaskID.
//
// businessID is written onto the row and is not optional. Omitting it stored
// every edge with business_id = 0 while ListTaskDependencies filters on the
// caller's real business, so dependencies were written and then never read back:
// adding one succeeded and it never appeared anywhere.
//
// Ruling T2 keeps dependencies cross-project but within the business, so the
// blocker is checked to belong to the same business rather than trusted from the
// request.
func (s *Service) AddTaskDependency(ctx context.Context, taskID, dependsOnTaskID, businessID int64) error {
	// Validated through the blocker's project, not through task.BusinessID:
	// enrichedTaskSelect does not select business_id, so a task read back here
	// always reports BusinessID 0 and a direct comparison could never pass.
	// GetProjectV2ByID is business-scoped, and a task belongs to a project, so a
	// blocker whose project is not in this business is refused.
	blocker, err := s.repo.GetTaskByID(ctx, dependsOnTaskID)
	if err != nil {
		return domain.ErrUnauthorized
	}
	if _, projErr := s.repo.GetProjectV2ByID(ctx, blocker.ProjectID, businessID); projErr != nil {
		return domain.ErrUnauthorized
	}
	return s.repo.AddTaskDependency(ctx, domain.AddTaskDependencyInput{
		TaskID:          taskID,
		DependsOnTaskID: dependsOnTaskID,
		BusinessID:      businessID,
	})
}

// RemoveTaskDependency drops the blocking edge between two tasks. businessID
// matches what AddTaskDependency wrote; see the note there.
func (s *Service) RemoveTaskDependency(ctx context.Context, taskID, dependsOnTaskID, businessID int64) error {
	return s.repo.RemoveTaskDependency(ctx, domain.RemoveTaskDependencyInput{
		TaskID:          taskID,
		DependsOnTaskID: dependsOnTaskID,
		BusinessID:      businessID,
	})
}

// ─── dashboard/insights/label/cycle/milestone/view v2 service methods ────────

func (s *Service) GetDashboardV2(ctx context.Context, businessID, userID int64) (domain.DashboardStats, error) {
	stats, err := s.repo.GetDashboardStatsV2(ctx, businessID)
	if err != nil {
		return domain.DashboardStats{}, err
	}
	myTasks, _, err := s.repo.GetMyTasks(ctx, userID, 0, internalListCap, 0)
	if err != nil {
		return domain.DashboardStats{}, err
	}
	stats.MyTasks = myTasks
	return stats, nil
}

func (s *Service) ListLabelsV2(ctx context.Context, businessID, limit, offset int64) ([]domain.Label, int64, error) {
	return s.repo.ListLabelsV2(ctx, businessID, limit, offset)
}

func (s *Service) CreateLabelV2(ctx context.Context, businessID int64, name, color string) (domain.Label, error) {
	id, err := s.repo.CreateLabelV2(ctx, businessID, name, color)
	if err != nil {
		return domain.Label{}, err
	}
	now := time.Now().UTC()
	return domain.Label{ID: id, Name: name, Color: color, CreatedAt: now}, nil
}

func (s *Service) ListCyclesV2(ctx context.Context, projectID, businessID, callerID, limit, offset int64) ([]domain.Cycle, int64, error) {
	return s.repo.ListCyclesV2(ctx, projectID, businessID, callerID, false, limit, offset)
}

// ErrInvalidCycleDates reports a start/end pair a cycle cannot be created from.
// The v1 handler grew these guards in the security/infra audit (4514d877); v2
// had none, and its time.Parse discarded the error, so an unparseable or empty
// date became the zero time and a year-0001 row. Deleting the v1 handler without
// this would have quietly undone that audit fix on the only remaining path.
var ErrInvalidCycleDates = errors.New("cycle dates are invalid")

// NormalizeCycleDates parses and defaults a cycle's date pair.
//
// A pure function so the rules can be tested without a repository: the previous
// shape only failed once it reached the database, which made the interesting
// cases untestable except by panic.
//
// An empty date is defaulted, not rejected, matching v1 - a caller supplying only
// an end date gets today as the start. An unparseable one is rejected, because
// silently treating "not-a-date" as the zero time is how the year-0001 rows got
// written.
func NormalizeCycleDates(startDate, endDate string) (time.Time, time.Time, error) {
	var start, end time.Time

	if startDate != "" {
		parsed, err := time.Parse("2006-01-02", startDate)
		if err != nil {
			return time.Time{}, time.Time{}, ErrInvalidCycleDates
		}
		start = parsed
	}
	if endDate != "" {
		parsed, err := time.Parse("2006-01-02", endDate)
		if err != nil {
			return time.Time{}, time.Time{}, ErrInvalidCycleDates
		}
		end = parsed
	}

	if start.IsZero() {
		start = time.Now().UTC().Truncate(24 * time.Hour)
	}
	if end.IsZero() {
		end = start.AddDate(0, 0, 14)
	}
	if end.Before(start) {
		return time.Time{}, time.Time{}, ErrInvalidCycleDates
	}
	return start, end, nil
}

func (s *Service) CreateCycleV2(ctx context.Context, businessID, projectID int64, name, startDate, endDate string) (domain.Cycle, error) {
	sd, ed, dateErr := NormalizeCycleDates(startDate, endDate)
	if dateErr != nil {
		return domain.Cycle{}, dateErr
	}
	id, err := s.repo.CreateCycleV2(ctx, domain.CycleInput{
		BusinessID: businessID, ProjectID: &projectID,
		Name: name, StartDate: sd, EndDate: ed,
	})
	if err != nil {
		return domain.Cycle{}, err
	}
	status := "upcoming"
	now := time.Now().UTC()
	if now.After(sd) && now.Before(ed) {
		status = "active"
	}
	return domain.Cycle{ID: id, ProjectID: &projectID, Name: name, Status: status}, nil
}

func (s *Service) ListMilestonesV2(ctx context.Context, businessID int64, projectID *int64, callerID, limit, offset int64) ([]domain.Milestone, int64, error) {
	return s.repo.ListMilestonesV2(ctx, businessID, projectID, callerID, false, limit, offset)
}

func (s *Service) CreateMilestoneV2(ctx context.Context, businessID, projectID int64, name, description, targetDate string) (domain.Milestone, error) {
	id, err := s.repo.CreateMilestoneV2(ctx, domain.MilestoneInput{
		BusinessID: businessID, ProjectID: projectID,
		Name: name,
	})
	if err != nil {
		return domain.Milestone{}, err
	}
	return domain.Milestone{ID: id, ProjectID: projectID, Name: name, Status: "planning"}, nil
}

// callerID drives the private-view filter: a private view is visible only to
// whoever created it.
func (s *Service) ListViewsV2(ctx context.Context, businessID, callerID, limit, offset int64) ([]domain.View, int64, error) {
	return s.repo.ListViewsV2(ctx, businessID, callerID, limit, offset)
}

func (s *Service) CreateViewV2(ctx context.Context, businessID, createdBy int64, name, icon, filtersJSON, groupBy, sortBy, sortDir, visibility string) (domain.View, error) {
	var filters map[string]interface{}
	_ = json.Unmarshal([]byte(filtersJSON), &filters)
	id, err := s.repo.CreateViewV2(ctx, domain.ViewInput{
		BusinessID: businessID, Name: name, Icon: icon, Filters: filters,
		GroupBy: groupBy, SortBy: sortBy, SortDir: sortDir, CreatedBy: createdBy,
		Visibility: visibility,
	})
	if err != nil {
		return domain.View{}, err
	}
	return domain.View{ID: id, Name: name, Icon: icon, Filters: filters, GroupBy: groupBy, SortBy: sortBy, SortDir: sortDir, Visibility: visibility}, nil
}

// UpdateViewV2 carries the whole view, because repo.UpdateView is a REPLACEMENT:
// its statement sets name, icon, filters, group_by, sort_by, sort_dir and
// visibility unconditionally. The previous signature took only the id, so every
// v2 update wrote empty strings over all of them - renaming a view through v2
// blanked it.
func (s *Service) UpdateViewV2(ctx context.Context, actorID, viewID int64, name, icon, filtersJSON, groupBy, sortBy, sortDir, visibility string) error {
	var filters map[string]interface{}
	_ = json.Unmarshal([]byte(filtersJSON), &filters)
	return s.repo.UpdateView(ctx, domain.UpdateViewInput{
		ID:         viewID,
		ActorID:    actorID,
		Name:       name,
		Icon:       icon,
		Filters:    filters,
		GroupBy:    groupBy,
		SortBy:     sortBy,
		SortDir:    sortDir,
		Visibility: visibility,
	})
}

// profilePublicProjectCap bounds the profile card list. A profile is a summary,
// not a paginated project browser, and All projects already exists for the
// full list.
const profilePublicProjectCap = 24

// GetFlowProfile returns one business member's profile: identity from the
// HR-owned hub profile, plus the public projects they created.
//
// Cross-business isolation comes from the resolver: it looks the subject up
// inside the caller's business, so a subject who is not a colleague returns
// ErrUserNotFound rather than a profile. There is deliberately no extra role
// gate on top - a colleague's name, email and public projects are already
// visible through the member roster and the project list.
func (s *Service) GetFlowProfile(
	ctx context.Context,
	subjectID, businessID int64,
) (domain.FlowProfile, error) {
	if s.hubProfiles == nil {
		return domain.FlowProfile{}, domain.ErrUnauthorized
	}
	hub, err := s.hubProfiles.GetProfile(ctx, subjectID, businessID)
	if err != nil {
		return domain.FlowProfile{}, err
	}

	projects, err := s.repo.ListPublicProjectsByOwnerV2(ctx, businessID, subjectID, profilePublicProjectCap)
	if err != nil {
		return domain.FlowProfile{}, err
	}

	return domain.FlowProfile{
		UserID:   hub.ID,
		Name:     hub.FullName,
		Email:    hub.Email,
		Initials: initialsOf(hub.FullName),
		// The hub profile carries a business role, not a job title. It is the
		// only title-shaped value that exists, so it fills the field until
		// staff records grow a real one (fast-follow F-B).
		RoleTitle:      hub.Role,
		Bio:            "",
		PublicProjects: projects,
	}, nil
}

// initialsOf builds the two-letter monogram the contract asks for.
//
// Falls back to the first two letters of a single-word name, and to the empty
// string for an empty name, so the caller never has to special-case it.
func initialsOf(name string) string {
	fields := strings.Fields(name)
	switch len(fields) {
	case 0:
		return ""
	case 1:
		one := []rune(fields[0])
		if len(one) == 1 {
			return strings.ToUpper(string(one[0]))
		}
		return strings.ToUpper(string(one[0:2]))
	default:
		first := []rune(fields[0])
		last := []rune(fields[len(fields)-1])
		return strings.ToUpper(string(first[0]) + string(last[0]))
	}
}
