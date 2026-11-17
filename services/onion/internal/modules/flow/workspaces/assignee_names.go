package workspaces

import "context"

// AssigneeNames returns the display name Flow's task list shows for each of
// userIDs in businessID, so another module can name a task's assignee the way
// Flow does without reading identity tables itself. Users with no name are
// left out. The project dashboard's overdue list is the consumer.
func (m *Module) AssigneeNames(ctx context.Context, businessID int64, userIDs []int64) (map[int64]string, error) {
	return m.repository.AssigneeNames(ctx, businessID, userIDs)
}
