package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
// rule: creator, member, or hub bypass. Public visibility never widens read
// access here - public projects surface as metadata on profiles, not as
// openable boards.
func (s *Service) canSeeProject(ctx context.Context, p domain.Project, callerID, businessID int64) bool {
	if s.hubProfileBypasses(ctx, callerID, businessID) {
		return true
	}
	if p.CreatedBy == callerID || p.OwnerID == callerID {
		return true
	}
	return s.repo.IsProjectMember(ctx, callerID, p.ID)
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
	includeAll := s.hubProfileBypasses(ctx, callerID, businessID)
	return s.repo.ListProjectsV2(ctx, businessID, callerID, includeAll, status, visibility, q, limit, offset)
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
	if s.hubProfileBypasses(ctx, userID, businessID) {
		return true
	}
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
	includeAll := s.hubProfileBypasses(ctx, callerID, businessID)
	filter.BusinessID = businessID
	return s.repo.GetTasksByBusinessV2(ctx, filter, callerID, includeAll)
}

// SearchTasksV2 is business-wide fulltext search under the same visibility
// clause as listing.
func (s *Service) SearchTasksV2(ctx context.Context, callerID, businessID int64, q string, limit, offset int64) ([]domain.Task, int64, error) {
	includeAll := s.hubProfileBypasses(ctx, callerID, businessID)
	return s.repo.SearchTasksV2(ctx, businessID, callerID, includeAll, q, limit, offset)
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
	if !s.repo.IsWorkspaceMember(ctx, assigneeID, p.WorkspaceID) {
		return fmt.Errorf("assignee is not a workspace member")
	}
	if s.hubProfileBypasses(ctx, assigneeID, businessID) {
		return nil
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
func (s *Service) AddTaskDependency(ctx context.Context, taskID, dependsOnTaskID int64) error {
	return s.repo.AddTaskDependency(ctx, domain.AddTaskDependencyInput{
		TaskID:          taskID,
		DependsOnTaskID: dependsOnTaskID,
	})
}

// RemoveTaskDependency drops the blocking edge between two tasks.
func (s *Service) RemoveTaskDependency(ctx context.Context, taskID, dependsOnTaskID int64) error {
	return s.repo.RemoveTaskDependency(ctx, domain.RemoveTaskDependencyInput{
		TaskID:          taskID,
		DependsOnTaskID: dependsOnTaskID,
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

func (s *Service) ListCyclesV2(ctx context.Context, projectID, businessID, limit, offset int64) ([]domain.Cycle, int64, error) {
	return s.repo.ListCyclesV2(ctx, projectID, businessID, limit, offset)
}

func (s *Service) CreateCycleV2(ctx context.Context, businessID, projectID int64, name, startDate, endDate string) (domain.Cycle, error) {
	sd, _ := time.Parse("2006-01-02", startDate)
	ed, _ := time.Parse("2006-01-02", endDate)
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

func (s *Service) ListMilestonesV2(ctx context.Context, businessID int64, projectID *int64, limit, offset int64) ([]domain.Milestone, int64, error) {
	return s.repo.ListMilestonesV2(ctx, businessID, projectID, limit, offset)
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

func (s *Service) ListViewsV2(ctx context.Context, businessID, limit, offset int64) ([]domain.View, int64, error) {
	return s.repo.ListViewsV2(ctx, businessID, limit, offset)
}

func (s *Service) CreateViewV2(ctx context.Context, businessID, createdBy int64, name, icon, filtersJSON, groupBy, sortBy, sortDir string) (domain.View, error) {
	var filters map[string]interface{}
	_ = json.Unmarshal([]byte(filtersJSON), &filters)
	id, err := s.repo.CreateViewV2(ctx, domain.ViewInput{
		BusinessID: businessID, Name: name, Icon: icon, Filters: filters,
		GroupBy: groupBy, SortBy: sortBy, SortDir: sortDir, CreatedBy: createdBy,
	})
	if err != nil {
		return domain.View{}, err
	}
	return domain.View{ID: id, Name: name, Icon: icon, Filters: filters, GroupBy: groupBy, SortBy: sortBy, SortDir: sortDir}, nil
}

func (s *Service) UpdateViewV2(ctx context.Context, actorID, viewID int64) error {
	return s.repo.UpdateView(ctx, domain.UpdateViewInput{ID: viewID, ActorID: actorID})
}
