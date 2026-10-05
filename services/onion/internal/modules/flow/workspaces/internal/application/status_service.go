package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
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
// Read access only, so the gate is viewer. An empty set is the normal answer for
// a project that has not configured any, and means "use the built-ins".
func (s *Service) ListProjectStatuses(ctx context.Context, projectID, callerID, businessID int64) ([]domain.Status, error) {
	if err := s.RequireProjectRole(ctx, projectID, callerID, businessID, domain.FlowRoleViewer); err != nil {
		return nil, domain.ErrUnauthorized
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

// DeleteProjectStatus removes a status, returning how many tasks fell back to
// their built-in string status.
func (s *Service) DeleteProjectStatus(ctx context.Context, id, callerID, businessID int64) (int64, error) {
	existing, err := s.repo.GetStatusByID(ctx, id, businessID)
	if err != nil {
		return 0, err
	}
	if roleErr := s.RequireProjectRole(ctx, existing.ProjectID, callerID, businessID, domain.FlowRoleAdmin); roleErr != nil {
		return 0, domain.ErrUnauthorized
	}
	return s.repo.DeleteStatus(ctx, id, businessID)
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
		if err := s.repo.SetTaskStatusID(ctx, taskID, businessID, 0, task.Status); err != nil {
			return "", err
		}
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

	if err := s.repo.SetTaskStatusID(ctx, taskID, businessID, statusID, status.LegacyStatus); err != nil {
		return "", err
	}
	if task.Status != status.LegacyStatus {
		task.Status = status.LegacyStatus
		s.announceStatusChange(ctx, task, callerID, status.LegacyStatus)
	}
	return status.LegacyStatus, nil
}
