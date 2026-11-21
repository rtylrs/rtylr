package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

// F-A: per-project custom statuses.
//
// Authorization sits here rather than in the SQL. Every statement in
// statuses.sql is scoped by business_id and nothing else, deliberately: the
// workspace-membership EXISTS clauses that used to double as authorization are
// unsatisfiable for any v2-created resource, and when SetTaskLabels kept them
// the caller passed the role check and was then denied by the write.

// ListProjectStatuses returns a project's status set.
//
// Read access only, so the gate is viewer. Statuses are the project's columns,
// so a project that somehow has none is seeded from the template first.
func (s *Service) ListProjectStatuses(ctx context.Context, projectID, callerID, businessID int64) ([]domain.Status, error) {
	if err := s.RequireProjectRole(ctx, projectID, callerID, businessID, domain.FlowRoleViewer); err != nil {
		return nil, domain.ErrUnauthorized
	}
	statuses, err := s.repo.ListStatusesByProject(ctx, businessID, projectID)
	if err != nil || len(statuses) > 0 {
		return statuses, err
	}
	if seedErr := s.repo.SeedDefaultColumns(ctx, projectID, 0, callerID); seedErr != nil {
		return nil, seedErr
	}
	return s.repo.ListStatusesByProject(ctx, businessID, projectID)
}

// GetProjectStatus returns a business-scoped status for callers that need its
// project before performing a mutation (for example, to publish the realtime
// invalidation after deletion).
func (s *Service) GetProjectStatus(ctx context.Context, id, businessID int64) (domain.Status, error) {
	return s.repo.GetStatusByID(ctx, id, businessID)
}

// CreateProjectStatus adds a status to a project's set.
//
// Admin, not member. Changing which states exist reshapes the board for everyone
// on the project, which is a configuration act rather than day-to-day work. It
// matches SetProjectVisibility and RestoreProjectV2 in taking the admin tier.
func (s *Service) CreateProjectStatus(ctx context.Context, callerID, businessID int64, in domain.StatusInput) (domain.Status, error) {
	if err := s.RequireProjectRole(ctx, in.ProjectID, callerID, businessID, domain.FlowRoleAdmin); err != nil {
		return domain.Status{}, domain.ErrUnauthorized
	}
	normalized, err := domain.NormalizeStatusInput(in)
	if err != nil {
		return domain.Status{}, err
	}
	return s.repo.CreateStatus(ctx, businessID, callerID, normalized)
}

// UpdateProjectStatus replaces a status.
//
// The status is read first so its project can be role-checked. Trusting a
// project id from the request instead would let a caller with admin on one
// project edit a status belonging to another: that is exactly the hole T2 had,
// where the handler role-checked taskId and then trusted dependsOnTaskId.
func (s *Service) UpdateProjectStatus(ctx context.Context, id, callerID, businessID int64, in domain.StatusInput) (domain.Status, error) {
	existing, err := s.repo.GetStatusByID(ctx, id, businessID)
	if err != nil {
		return domain.Status{}, err
	}
	if roleErr := s.RequireProjectRole(ctx, existing.ProjectID, callerID, businessID, domain.FlowRoleAdmin); roleErr != nil {
		return domain.Status{}, domain.ErrUnauthorized
	}
	in.ProjectID = existing.ProjectID
	normalized, err := domain.NormalizeStatusInput(in)
	if err != nil {
		return domain.Status{}, err
	}
	return s.repo.UpdateStatus(ctx, id, businessID, existing.ProjectID, normalized)
}

// ReorderProjectStatuses sets the order of a project's whole status set at
// once. Admin, like create and update: the order is the board's column order.
// ids must be exactly the project's set; anything else is
// ErrStatusOrderMismatch and nothing is written.
func (s *Service) ReorderProjectStatuses(ctx context.Context, projectID, callerID, businessID int64, ids []int64) ([]domain.Status, error) {
	if err := s.RequireProjectRole(ctx, projectID, callerID, businessID, domain.FlowRoleAdmin); err != nil {
		return nil, domain.ErrUnauthorized
	}
	return s.repo.ReorderStatuses(ctx, businessID, projectID, ids)
}

// DeleteProjectStatus removes a status, moving its tasks to moveToID, and
// returns how many moved. A status with tasks needs a target; the project's
// last done status cannot be deleted.
func (s *Service) DeleteProjectStatus(ctx context.Context, id, moveToID, callerID, businessID int64) (int64, error) {
	existing, err := s.repo.GetStatusByID(ctx, id, businessID)
	if err != nil {
		return 0, err
	}
	if roleErr := s.RequireProjectRole(ctx, existing.ProjectID, callerID, businessID, domain.FlowRoleAdmin); roleErr != nil {
		return 0, domain.ErrUnauthorized
	}
	// Deleting a status moves every task on it in one statement, which can
	// close many blockers at once.
	var blockers []int64
	if moveToID != 0 {
		if blockers, err = s.repo.ListOpenBlockerIDsByStatus(ctx, id, businessID); err != nil {
			logger.Ctx(ctx).Error("DeleteProjectStatus: failed to read open blockers, no unblock notification will be sent",
				logger.Error(err))
		}
	}
	moved, err := s.repo.DeleteStatus(ctx, id, businessID, moveToID)
	if err != nil {
		return 0, err
	}
	if len(blockers) > 0 {
		s.announceUnblocked(ctx, businessID, callerID, blockers)
	}
	return moved, nil
}

// SetTaskStatus applies a custom status to a task, or clears it when statusID is
// zero, and returns the built-in string status the task now carries.
//
// Member, matching every other task mutation.
//
// Two checks, not one. RequireTaskRole covers the task, and the status is then
// read and confirmed to belong to the SAME project. Without the second check a
// member of project A could put a task in project A onto project B's status,
// which would render as a status name nobody on A can see and count into B's
// rollups.
func (s *Service) SetTaskStatus(ctx context.Context, taskID, statusID, callerID, businessID int64) (string, error) {
	if err := s.RequireTaskRole(ctx, taskID, callerID, businessID, domain.FlowRoleMember); err != nil {
		return "", domain.ErrUnauthorized
	}

	// Clearing needs no status lookup, but it does need a legacy value to write,
	// and the task's current one is the only correct answer: clearing a custom
	// status must not also move the task.
	if statusID == 0 {
		task, err := s.repo.GetTaskByID(ctx, taskID)
		if err != nil {
			return "", err
		}
		announceUnblocked := s.watchUnblocks(ctx, businessID, taskID)
		if err := s.repo.SetTaskStatusID(ctx, taskID, businessID, 0, task.Status); err != nil {
			return "", err
		}
		announceUnblocked(callerID)
		return task.Status, nil
	}

	status, err := s.repo.GetStatusByID(ctx, statusID, businessID)
	if err != nil {
		return "", err
	}
	task, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		return "", err
	}
	if status.ProjectID != task.ProjectID {
		return "", domain.ErrUnauthorized
	}

	announceUnblocked := s.watchUnblocks(ctx, businessID, taskID)
	if err := s.repo.SetTaskStatusID(ctx, taskID, businessID, statusID, status.LegacyStatus); err != nil {
		return "", err
	}
	announceUnblocked(callerID)
	if task.Status != status.LegacyStatus {
		task.Status = status.LegacyStatus
		s.announceStatusChange(ctx, task, callerID, status.LegacyStatus)
	}
	return status.LegacyStatus, nil
}

// GetStatusTemplate returns the set new projects in this business copy.
func (s *Service) GetStatusTemplate(ctx context.Context, businessID int64) ([]domain.StatusTemplateItem, error) {
	return s.repo.GetStatusTemplate(ctx, businessID)
}

// SaveStatusTemplate validates and replaces the template. Existing projects
// keep their own statuses. The caller's permission is checked by the API.
func (s *Service) SaveStatusTemplate(ctx context.Context, businessID int64, items []domain.StatusTemplateItem) ([]domain.StatusTemplateItem, error) {
	normalized, err := domain.NormalizeStatusTemplate(items)
	if err != nil {
		return nil, err
	}
	if err := s.repo.SaveStatusTemplate(ctx, businessID, normalized); err != nil {
		return nil, err
	}
	return normalized, nil
}
