package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
)

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

// ErrInvalidMilestoneDate reports a target date a milestone cannot be created
// with.
var ErrInvalidMilestoneDate = errors.New("milestone target date is invalid")

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
