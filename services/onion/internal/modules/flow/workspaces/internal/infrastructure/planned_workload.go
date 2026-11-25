package infrastructure

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

// ListPlannedTasks returns the open, assigned, estimated tasks whose dates
// touch [from, to]. With projectIDs nil every project of the business counts;
// otherwise only those projects, plus the tasks assigned to ownerID anywhere.
func (r *Repository) ListPlannedTasks(ctx context.Context, businessID int64, projectIDs []int64, ownerID int64, from, to time.Time) ([]domain.PlannedTask, error) {
	query := `SELECT t.id, t.title, t.project_id, t.assignee_id, t.start_date, t.due_date, ` + taskEstimateMinutesSQL + `
  FROM flow_tasks t
  JOIN flow_projects p ON p.id = t.project_id AND p.business_id = t.business_id
 WHERE t.business_id = ? AND t.assignee_id IS NOT NULL AND COALESCE(t.archived, 0) = 0
   AND p.archived_at IS NULL
   AND ` + taskEstimateMinutesSQL + ` > 0
   AND NOT ` + taskFinishedSQL + `
   AND COALESCE(t.start_date, t.due_date) <= ? AND COALESCE(t.due_date, t.start_date) >= ?`
	args := []any{businessID, to.Format("2006-01-02"), from.Format("2006-01-02")}
	if projectIDs != nil {
		scope := "t.assignee_id = ?"
		args = append(args, ownerID)
		if len(projectIDs) > 0 {
			placeholders, ids := int64Placeholders(projectIDs)
			scope = "(t.project_id IN (" + placeholders + ") OR t.assignee_id = ?)"
			args = append(args[:len(args)-1], append(ids, ownerID)...)
		}
		query += " AND " + scope
	}
	query += " ORDER BY t.id LIMIT 5000"

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Ctx(ctx).Error("ListPlannedTasks failed", logger.Error(err))
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []domain.PlannedTask
	for rows.Next() {
		var task domain.PlannedTask
		var start, due sql.NullTime
		var minutes float64
		if err := rows.Scan(&task.ID, &task.Title, &task.ProjectID, &task.AssigneeID, &start, &due, &minutes); err != nil {
			return nil, err
		}
		if start.Valid {
			task.StartDate = &start.Time
		}
		if due.Valid {
			task.DueDate = &due.Time
		}
		task.EstimateMinutes = int64(minutes)
		out = append(out, task)
	}
	return out, rows.Err()
}
