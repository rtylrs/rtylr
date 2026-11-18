package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
)

// businessWideOpens reports whether callerID may open p because they hold
// business-wide Flow access in the business it belongs to: the business owner,
// an admin, or a Flow Admin holding flow:team:manage (hubProfileBypasses).
//
// This is access, not listing (ruling P8): the sidebar stays curated, and they
// reach a private project by link, the Insights project picker or the recycle
// bin. Opening is reading. It gives no write a Flow Admin did not already have;
// RequireProjectRole admits them at viewer level only.
//
// Two things stay closed. A My Tasks bucket belongs to one person and is exempt
// from every visibility policy, so nobody opens someone else's. And the project
// must belong to the business the caller is acting in, so business-wide access
// somewhere never opens a project somewhere else.
func (s *Service) businessWideOpens(ctx context.Context, p domain.Project, callerID, businessID int64) bool {
	if p.IsPersonal || p.BusinessID != businessID {
		return false
	}
	return s.hubProfileBypasses(ctx, callerID, businessID)
}

// opensProject is canSeeProject addressed by id, for reads scoped to one
// project: whoever may open the project may read everything on its board. A
// project that does not exist in businessID is false.
//
// It reads the four proxy facts rather than the full project row, because it
// runs on every project-scoped listing and the full row carries count
// subqueries none of these checks need.
func (s *Service) opensProject(ctx context.Context, projectID, callerID, businessID int64) bool {
	if projectID <= 0 || callerID <= 0 || businessID <= 0 || s.repo == nil {
		return false
	}
	p, err := s.repo.GetProjectProxyFacts(ctx, projectID)
	if err != nil || p.BusinessID != businessID {
		return false
	}
	return s.canSeeProject(ctx, p, callerID, businessID)
}

// CanReadTask is the read gate for a task's subresources: its subtasks,
// activity, attachments, time logs and links. It admits whoever IsTaskMember
// admits, plus anyone who may open the task's project (canSeeProject), such as
// a department manager of someone on it or a business owner, admin or Flow
// Admin, so the task drawer works on every board they can open.
//
// IsTaskMember itself is left alone because it also gates writes such as adding
// a comment; this only widens reads.
func (s *Service) CanReadTask(ctx context.Context, userID, businessID, taskID int64) bool {
	if s.IsTaskMember(ctx, userID, taskID) {
		return true
	}
	if taskID <= 0 || s.repo == nil {
		return false
	}
	task, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil || task.BusinessID != businessID {
		return false
	}
	return s.opensProject(ctx, task.ProjectID, userID, businessID)
}
