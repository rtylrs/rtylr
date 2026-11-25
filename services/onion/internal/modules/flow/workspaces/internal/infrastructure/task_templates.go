package infrastructure

import (
	"context"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

// ListTaskTemplates returns a project's templates, by task type.
func (r *Repository) ListTaskTemplates(ctx context.Context, businessID, projectID int64) ([]domain.TaskTemplate, error) {
	rows, err := r.pool.QueryContext(ctx, `SELECT id, project_id, task_type, body, updated_at
  FROM flow_task_templates WHERE business_id = ? AND project_id = ? ORDER BY task_type`, businessID, projectID)
	if err != nil {
		logger.Ctx(ctx).Error("ListTaskTemplates failed", logger.Error(err))
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []domain.TaskTemplate
	for rows.Next() {
		var t domain.TaskTemplate
		if err := rows.Scan(&t.ID, &t.ProjectID, &t.TaskType, &t.Body, &t.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// UpsertTaskTemplate saves the template for the project's task type.
func (r *Repository) UpsertTaskTemplate(ctx context.Context, businessID, projectID int64, taskType, body string, actorID int64) error {
	_, err := r.pool.ExecContext(ctx, `INSERT INTO flow_task_templates (business_id, project_id, task_type, body, updated_by)
VALUES (?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE body = VALUES(body), updated_by = VALUES(updated_by)`,
		businessID, projectID, taskType, body, actorID)
	if err != nil {
		logger.Ctx(ctx).Error("UpsertTaskTemplate failed", logger.Error(err))
	}
	return err
}

// DeleteTaskTemplate removes the template for the project's task type.
func (r *Repository) DeleteTaskTemplate(ctx context.Context, businessID, projectID int64, taskType string) error {
	_, err := r.pool.ExecContext(ctx,
		"DELETE FROM flow_task_templates WHERE business_id = ? AND project_id = ? AND task_type = ?",
		businessID, projectID, taskType)
	if err != nil {
		logger.Ctx(ctx).Error("DeleteTaskTemplate failed", logger.Error(err))
	}
	return err
}
