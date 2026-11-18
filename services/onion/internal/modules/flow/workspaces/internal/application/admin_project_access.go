package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
)

// administratorOpens reports whether callerID may open p because they own or
// administer the business it belongs to.
//
// This is access, not listing (ruling P8): the sidebar stays curated, and an
// administrator reaches a private project by link, the Insights project picker
// or the recycle bin. It mirrors the business owner/admin pass that
// RequireProjectRole already gives writes, so an administrator who may edit a
// project may also open it.
//
// Two things stay closed. A My Tasks bucket belongs to one person and is exempt
// from every visibility policy, so no administrator opens someone else's. And
// the project must belong to the business the caller is acting in, so being an
// administrator somewhere never opens a project somewhere else.
func (s *Service) administratorOpens(ctx context.Context, p domain.Project, callerID, businessID int64) bool {
	if p.IsPersonal || p.BusinessID != businessID {
		return false
	}
	return s.isBusinessAdministrator(ctx, callerID, businessID)
}

// administratorOpensProject is administratorOpens addressed by id, for reads
// scoped to one project. A project that does not exist in businessID is false.
//
// The administrator check runs first because it is the cheaper answer for the
// common caller, who is not one, and saves the project lookup.
func (s *Service) administratorOpensProject(ctx context.Context, projectID, callerID, businessID int64) bool {
	if projectID <= 0 || callerID <= 0 || businessID <= 0 {
		return false
	}
	if !s.isBusinessAdministrator(ctx, callerID, businessID) {
		return false
	}
	p, err := s.repo.GetProjectV2ByID(ctx, projectID, businessID)
	if err != nil {
		return false
	}
	return s.administratorOpens(ctx, p, callerID, businessID)
}

// CanReadTask is the read gate for a task's subresources: its subtasks,
// activity, attachments, time logs and links. It admits whoever IsTaskMember
// admits, plus a business owner or admin reading a task in a project they may
// open, so the task drawer works on a board they opened by link.
//
// IsTaskMember itself is left alone because it also gates writes such as adding
// a comment; this only widens reads.
func (s *Service) CanReadTask(ctx context.Context, userID, businessID, taskID int64) bool {
	if s.IsTaskMember(ctx, userID, taskID) {
		return true
	}
	if taskID <= 0 || !s.isBusinessAdministrator(ctx, userID, businessID) {
		return false
	}
	task, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil || task.BusinessID != businessID {
		return false
	}
	return s.administratorOpensProject(ctx, task.ProjectID, userID, businessID)
}
