package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/flow/insights/internal/domain"
)

// AssigneeNames reads the display name Flow's task list shows for each of
// userIDs in businessID. The workspaces module serves it from Flow's identity
// reader; users with no name are left out of the result.
type AssigneeNames func(ctx context.Context, businessID int64, userIDs []int64) (map[int64]string, error)

// NameOverdueAssignees sets AssigneeName on every assigned task in tasks with
// one read through names. A nil names leaves every name empty. A read failure
// is returned, like Flow's own task list does.
func NameOverdueAssignees(ctx context.Context, names AssigneeNames, businessID int64, tasks []domain.OverdueTask) error {
	if names == nil {
		return nil
	}
	ids := make([]int64, 0, len(tasks))
	seen := make(map[int64]bool, len(tasks))
	for _, task := range tasks {
		if task.AssigneeID > 0 && !seen[task.AssigneeID] {
			seen[task.AssigneeID] = true
			ids = append(ids, task.AssigneeID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	resolved, err := names(ctx, businessID, ids)
	if err != nil {
		return err
	}
	for i := range tasks {
		if tasks[i].AssigneeID > 0 {
			tasks[i].AssigneeName = resolved[tasks[i].AssigneeID]
		}
	}
	return nil
}
