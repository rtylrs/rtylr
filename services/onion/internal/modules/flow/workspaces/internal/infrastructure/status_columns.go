package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

// A project's flow_statuses rows are its board columns. A task's status_id is
// its column, section_id mirrors status_id for clients that still read it, and
// the built-in string status follows the status's legacy_status so every
// existing filter, index and report keeps working. The helpers below are the
// only places that keep those three columns in step.

type sqlRunner interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// syncTaskFromSectionSQL makes the task's column (section_id) authoritative:
// the board and the old Section field both write section_id.
const syncTaskFromSectionSQL = `UPDATE flow_tasks t
JOIN flow_statuses s ON s.id = t.section_id AND s.project_id = t.project_id
SET t.status_id = s.id,
    t.completed_at = CASE WHEN s.legacy_status = 'done' AND t.completed_at IS NULL THEN NOW()
                          WHEN s.legacy_status <> 'done' THEN NULL ELSE t.completed_at END,
    t.status = s.legacy_status
WHERE t.id = ?`

// syncTaskFromLegacySQL makes the built-in string status authoritative. The
// current status_id is kept when it belongs to the task's project and agrees
// with the string status; otherwise the project's first status with that
// string, then the first in the same bucket, is picked.
const syncTaskFromLegacySQL = `UPDATE flow_tasks t
SET t.status_id = COALESCE(
      (SELECT s.id FROM flow_statuses s
        WHERE s.id = t.status_id AND s.project_id = t.project_id AND s.legacy_status = t.status),
      (SELECT s.id FROM flow_statuses s
        WHERE s.project_id = t.project_id AND s.legacy_status = t.status
        ORDER BY s.sort_order, s.id LIMIT 1),
      (SELECT s.id FROM flow_statuses s
        WHERE s.project_id = t.project_id
          AND s.bucket = CASE t.status WHEN 'todo' THEN 'backlog' WHEN 'backlog' THEN 'backlog'
                                       WHEN 'in_progress' THEN 'active' WHEN 'in_review' THEN 'active'
                                       ELSE t.status END
        ORDER BY s.sort_order, s.id LIMIT 1)
    ),
    t.section_id = t.status_id
WHERE t.id = ?`

func syncTaskFromSection(ctx context.Context, q sqlRunner, taskID int64) error {
	if _, err := q.ExecContext(ctx, syncTaskFromSectionSQL, taskID); err != nil {
		logger.Ctx(ctx).Error("syncTaskFromSection failed", logger.Error(err))
		return err
	}
	// A section that is not a status of the task's project (or none at all)
	// falls back to the string status.
	return syncTaskFromLegacy(ctx, q, taskID)
}

func syncTaskFromLegacy(ctx context.Context, q sqlRunner, taskID int64) error {
	if _, err := q.ExecContext(ctx, syncTaskFromLegacySQL, taskID); err != nil {
		logger.Ctx(ctx).Error("syncTaskFromLegacy failed", logger.Error(err))
		return err
	}
	return nil
}

func projectBusinessID(ctx context.Context, q sqlRunner, projectID int64) (int64, error) {
	var businessID int64
	err := q.QueryRowContext(ctx, "SELECT business_id FROM flow_projects WHERE id = ?", projectID).Scan(&businessID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, domain.ErrProjectNotFound
	}
	return businessID, err
}

// statusTemplate returns the business's template, or the built-in set when it
// has never saved one.
func (r *Repository) statusTemplate(ctx context.Context, q sqlRunner, businessID int64) ([]domain.StatusTemplateItem, error) {
	rows, err := q.QueryContext(ctx, `SELECT name, bucket, legacy_status, color, sort_order
FROM flow_status_templates WHERE business_id = ? ORDER BY sort_order, id`, businessID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var items []domain.StatusTemplateItem
	for rows.Next() {
		var item domain.StatusTemplateItem
		if err := rows.Scan(&item.Name, &item.Bucket, &item.LegacyStatus, &item.Color, &item.SortOrder); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return domain.DefaultStatusSet(), nil
	}
	return items, nil
}

// GetStatusTemplate is statusTemplate for the application layer.
func (r *Repository) GetStatusTemplate(ctx context.Context, businessID int64) ([]domain.StatusTemplateItem, error) {
	return r.statusTemplate(ctx, r.pool, businessID)
}

// SaveStatusTemplate replaces the business's template. Existing projects keep
// their own statuses; only projects created afterwards copy the new set.
func (r *Repository) SaveStatusTemplate(ctx context.Context, businessID int64, items []domain.StatusTemplateItem) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, "DELETE FROM flow_status_templates WHERE business_id = ?", businessID); err != nil {
		return err
	}
	for _, item := range items {
		if _, err := tx.ExecContext(ctx, `INSERT INTO flow_status_templates
(business_id, name, bucket, legacy_status, color, sort_order) VALUES (?, ?, ?, ?, ?, ?)`,
			businessID, item.Name, item.Bucket, item.LegacyStatus, item.Color, item.SortOrder); err != nil {
			if isDuplicateKey(err) {
				return domain.ErrStatusNameTaken
			}
			return err
		}
	}
	return tx.Commit()
}

// seedProjectStatuses gives a project without statuses a copy of its
// business's template. Safe to call repeatedly and concurrently: a project
// that already has statuses is left alone, and the unique name per project
// turns a racing duplicate insert into a no-op.
func (r *Repository) seedProjectStatuses(ctx context.Context, projectID, createdBy int64) error {
	var existing int64
	if err := r.pool.QueryRowContext(ctx, "SELECT COUNT(*) FROM flow_statuses WHERE project_id = ?", projectID).Scan(&existing); err != nil {
		return err
	}
	if existing > 0 {
		return nil
	}
	businessID, err := projectBusinessID(ctx, r.pool, projectID)
	if err != nil {
		return err
	}
	items, err := r.statusTemplate(ctx, r.pool, businessID)
	if err != nil {
		return err
	}
	var defaultID int64
	for _, item := range items {
		res, err := r.pool.ExecContext(ctx, `INSERT IGNORE INTO flow_statuses
(business_id, project_id, name, bucket, legacy_status, color, sort_order, is_default, created_by)
VALUES (?, ?, ?, ?, ?, ?, ?, FALSE, ?)`,
			businessID, projectID, item.Name, item.Bucket, item.LegacyStatus, item.Color, item.SortOrder, createdBy)
		if err != nil {
			return err
		}
		if defaultID == 0 && item.LegacyStatus == domain.TaskStatusTodo {
			if id, idErr := res.LastInsertId(); idErr == nil {
				defaultID = id
			}
		}
	}
	if defaultID != 0 {
		if _, err := r.pool.ExecContext(ctx, "UPDATE flow_statuses SET is_default = TRUE WHERE id = ?", defaultID); err != nil {
			return err
		}
	}
	return nil
}

// SeedDefaultColumns seeds the project's statuses, which are its columns.
// workspaceID is unused since columns stopped being stored per workspace.
func (r *Repository) SeedDefaultColumns(ctx context.Context, projectID, _, createdBy int64) error {
	return r.seedProjectStatuses(ctx, projectID, createdBy)
}

// moveTasksToStatus moves every task on one status to another status of the
// same project, keeping the three status columns and completed_at in step.
func moveTasksToStatus(ctx context.Context, q sqlRunner, fromID int64, to domain.Status) (int64, error) {
	res, err := q.ExecContext(ctx, `UPDATE flow_tasks t
SET t.status_id = ?, t.section_id = ?, t.status = ?,
    t.completed_at = CASE WHEN ? = 'done' AND t.completed_at IS NULL THEN NOW()
                          WHEN ? <> 'done' THEN NULL ELSE t.completed_at END
WHERE t.status_id = ? AND t.business_id = ?`,
		to.ID, to.ID, to.LegacyStatus, to.LegacyStatus, to.LegacyStatus, fromID, to.BusinessID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// deleteStatusTx deletes a status inside an open transaction. Tasks on it move
// to moveToID, which is required when there are any, and the project's last
// done status is refused.
func deleteStatusTx(ctx context.Context, tx *sql.Tx, existing domain.Status, moveToID int64, loadStatus func(id int64) (domain.Status, error)) (int64, error) {
	var taskCount int64
	if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM flow_tasks WHERE status_id = ? AND business_id = ? FOR UPDATE",
		existing.ID, existing.BusinessID).Scan(&taskCount); err != nil {
		return 0, err
	}
	if existing.Bucket == domain.StatusBucketDone {
		var otherDone int64
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM flow_statuses
WHERE project_id = ? AND business_id = ? AND bucket = 'done' AND id <> ?`,
			existing.ProjectID, existing.BusinessID, existing.ID).Scan(&otherDone); err != nil {
			return 0, err
		}
		if otherDone == 0 {
			return 0, domain.ErrLastDoneStatus
		}
	}

	var moved int64
	var target domain.Status
	if moveToID != 0 {
		var err error
		target, err = loadStatus(moveToID)
		if err != nil || target.ProjectID != existing.ProjectID || target.ID == existing.ID {
			return 0, domain.ErrStatusMoveTarget
		}
	}
	if taskCount > 0 {
		if moveToID == 0 {
			return 0, domain.ErrStatusInUse
		}
		var err error
		if moved, err = moveTasksToStatus(ctx, tx, existing.ID, target); err != nil {
			return 0, err
		}
	}

	res, err := tx.ExecContext(ctx, "DELETE FROM flow_statuses WHERE id = ? AND business_id = ?", existing.ID, existing.BusinessID)
	if err != nil {
		return 0, err
	}
	if affected, _ := res.RowsAffected(); affected == 0 {
		return 0, domain.ErrStatusNotFound
	}

	if existing.IsDefault {
		next := target.ID
		if next == 0 {
			_ = tx.QueryRowContext(ctx, `SELECT id FROM flow_statuses WHERE project_id = ? AND business_id = ?
ORDER BY (bucket = 'backlog') DESC, sort_order, id LIMIT 1`, existing.ProjectID, existing.BusinessID).Scan(&next)
		}
		if next != 0 {
			if _, err := tx.ExecContext(ctx, "UPDATE flow_statuses SET is_default = (id = ?) WHERE project_id = ? AND business_id = ?",
				next, existing.ProjectID, existing.BusinessID); err != nil {
				return 0, err
			}
		}
	}
	return moved, nil
}
