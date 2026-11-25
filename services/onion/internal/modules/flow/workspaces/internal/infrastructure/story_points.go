package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

// AttachTaskStoryPoints sets StoryPoints on every task from flow_tasks. The
// generated task reads predate the column, so it is read here, once per page.
func (r *Repository) AttachTaskStoryPoints(ctx context.Context, tasks []domain.Task) error {
	if len(tasks) == 0 {
		return nil
	}
	index := make(map[int64]int, len(tasks))
	ids := make([]int64, 0, len(tasks))
	for i := range tasks {
		index[tasks[i].ID] = i
		ids = append(ids, tasks[i].ID)
		tasks[i].StoryPoints = nil
	}
	placeholders, args := int64Placeholders(ids)
	rows, err := r.pool.QueryContext(ctx, `SELECT id, story_points FROM flow_tasks
 WHERE id IN (`+placeholders+`) AND story_points IS NOT NULL`, args...)
	if err != nil {
		logger.Ctx(ctx).Error("AttachTaskStoryPoints failed", logger.Error(err))
		return err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var id int64
		var points int32
		if err := rows.Scan(&id, &points); err != nil {
			return err
		}
		tasks[index[id]].StoryPoints = &points
	}
	return rows.Err()
}

// SetTaskStoryPoints writes a task's story points; nil clears them.
func (r *Repository) SetTaskStoryPoints(ctx context.Context, businessID, taskID int64, points *int32) error {
	value := sql.NullInt32{}
	if points != nil {
		value = sql.NullInt32{Int32: *points, Valid: true}
	}
	_, err := r.pool.ExecContext(ctx, "UPDATE flow_tasks SET story_points = ? WHERE id = ? AND business_id = ?",
		value, taskID, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("SetTaskStoryPoints failed", logger.Error(err))
	}
	return err
}
