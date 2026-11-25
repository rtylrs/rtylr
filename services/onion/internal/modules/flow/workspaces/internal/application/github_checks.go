package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
)

// ListTaskPullRequests returns the PRs linked to a task with their checks, for
// anyone who can read the task (T-376).
func (s *Service) ListTaskPullRequests(ctx context.Context, callerID, businessID, taskID int64) ([]domain.TaskPullRequest, error) {
	if !s.CanReadTask(ctx, callerID, businessID, taskID) {
		return nil, domain.ErrUnauthorized
	}
	return s.repo.ListTaskPullRequests(ctx, businessID, taskID)
}

// attachTaskChecksStates fills ChecksState on a page of tasks of one business.
func (s *Service) attachTaskChecksStates(ctx context.Context, tasks []domain.Task) error {
	if len(tasks) == 0 {
		return nil
	}
	ids := make([]int64, len(tasks))
	for i := range tasks {
		ids[i] = tasks[i].ID
	}
	states, err := s.repo.TaskChecksStates(ctx, tasks[0].BusinessID, ids)
	if err != nil {
		return err
	}
	for i := range tasks {
		tasks[i].ChecksState = states[tasks[i].ID]
	}
	return nil
}
