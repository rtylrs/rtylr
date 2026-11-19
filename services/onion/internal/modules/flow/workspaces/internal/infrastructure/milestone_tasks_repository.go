package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/voxire/onion/internal/platform/logger"
)

// MilestoneProjectID returns the project a milestone belongs to, or 0 when no
// such milestone exists in the business, so another tenant's id never matches.
func (r *Repository) MilestoneProjectID(ctx context.Context, milestoneID, businessID int64) (int64, error) {
	var projectID int64
	err := r.pool.QueryRowContext(ctx,
		`SELECT project_id FROM flow_milestones WHERE id = ? AND business_id = ?`,
		milestoneID, businessID,
	).Scan(&projectID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	return projectID, err
}

// SetTaskMilestone files a task under one milestone, or none when milestoneID
// is nil. The join table allows several rows per task, so the old link is
// removed in the same transaction to keep a task in at most one milestone.
func (r *Repository) SetTaskMilestone(ctx context.Context, taskID int64, milestoneID *int64) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `DELETE FROM flow_milestone_tasks WHERE task_id = ?`, taskID); err != nil {
		logger.Ctx(ctx).Error("SetTaskMilestone delete failed", logger.Error(err))
		return err
	}
	if milestoneID != nil {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO flow_milestone_tasks (milestone_id, task_id) VALUES (?, ?)`,
			*milestoneID, taskID,
		); err != nil {
			logger.Ctx(ctx).Error("SetTaskMilestone insert failed", logger.Error(err))
			return err
		}
	}
	return tx.Commit()
}

// TaskMilestoneIDs maps each task id to its milestone. Tasks without one are
// absent from the map.
func (r *Repository) TaskMilestoneIDs(ctx context.Context, taskIDs []int64) (map[int64]int64, error) {
	out := make(map[int64]int64)
	if len(taskIDs) == 0 {
		return out, nil
	}
	args := make([]interface{}, len(taskIDs))
	for i, id := range taskIDs {
		args[i] = id
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(taskIDs)), ",")
	rows, err := r.pool.QueryContext(ctx,
		fmt.Sprintf(`SELECT task_id, MIN(milestone_id) FROM flow_milestone_tasks WHERE task_id IN (%s) GROUP BY task_id`, placeholders),
		args...,
	)
	if err != nil {
		logger.Ctx(ctx).Error("TaskMilestoneIDs failed", logger.Error(err))
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var taskID, milestoneID int64
		if err := rows.Scan(&taskID, &milestoneID); err != nil {
			return nil, err
		}
		out[taskID] = milestoneID
	}
	return out, rows.Err()
}

// DeleteMilestoneV2 removes a milestone the actor can reach through its
// project, and unlinks its tasks. The tasks themselves are kept.
func (r *Repository) DeleteMilestoneV2(ctx context.Context, id, actorID int64) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.ExecContext(ctx,
		fmt.Sprintf(`DELETE FROM flow_milestones WHERE id = ? AND %s`, memberPredicate("flow_milestones")),
		id, actorID, actorID, actorID, actorID, actorID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("DeleteMilestoneV2 failed", logger.Error(err))
		return err
	}
	if err := requireAffected(res); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM flow_milestone_tasks WHERE milestone_id = ?`, id); err != nil {
		logger.Ctx(ctx).Error("DeleteMilestoneV2 unlink failed", logger.Error(err))
		return err
	}
	return tx.Commit()
}
