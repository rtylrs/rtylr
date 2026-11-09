package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/platform/logger"
)

// A board column can stand for a task status (flow_project_columns.
// linked_status). The four default columns do; a column someone adds does not.
// The link holds both ways: a card dropped on a linked column takes its status,
// and a task that takes a status moves to that status's column. These helpers
// are the only place the rule lives; the services call them after each write.

type sqlExecQuerier interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// placeNewTask settles a new top-level task's column and status before it is
// inserted. With a column given, a linked column decides the status. Without
// one, the task goes to its status's column, or the project's first column
// when its status has none (backlog, cancelled), instead of no column at all.
func placeNewTask(ctx context.Context, q sqlExecQuerier, projectID int64, sectionID *int64, status string) (*int64, string, error) {
	if sectionID != nil {
		var linked sql.NullString
		err := q.QueryRowContext(ctx,
			"SELECT linked_status FROM flow_project_columns WHERE id = ? AND project_id = ?",
			*sectionID, projectID).Scan(&linked)
		if errors.Is(err, sql.ErrNoRows) {
			return sectionID, status, nil
		}
		if err != nil {
			return nil, "", err
		}
		if linked.Valid {
			return sectionID, linked.String, nil
		}
		return sectionID, status, nil
	}

	var column int64
	err := q.QueryRowContext(ctx,
		`SELECT id FROM flow_project_columns WHERE project_id = ?
		 ORDER BY (linked_status <=> ?) DESC, sort_order ASC, id ASC LIMIT 1`,
		projectID, status).Scan(&column)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, status, nil
	}
	if err != nil {
		return nil, "", err
	}
	return &column, status, nil
}

// AlignTaskStatusToColumn gives a top-level task the status of the column it
// sits in, when that column is linked and the status differs. The matching
// custom status (flow_statuses) comes along, so the two never drift, and
// completed_at follows done. It reports the status before and after; changed
// is false when nothing was written.
func (r *Repository) AlignTaskStatusToColumn(ctx context.Context, taskID int64) (from, to string, changed bool, err error) {
	var linked sql.NullString
	err = r.pool.QueryRowContext(ctx,
		`SELECT t.status, c.linked_status FROM flow_tasks t
		 JOIN flow_project_columns c ON c.id = t.section_id
		 WHERE t.id = ? AND t.parent_task_id IS NULL`, taskID).Scan(&from, &linked)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", false, nil
	}
	if err != nil {
		return "", "", false, err
	}
	if !linked.Valid || linked.String == from {
		return from, from, false, nil
	}
	to = linked.String
	if _, err = r.pool.ExecContext(ctx,
		`UPDATE flow_tasks t
		 SET t.status = ?,
		     t.status_id = (SELECT s.id FROM flow_statuses s
		                    WHERE s.project_id = t.project_id AND s.legacy_status = ?
		                    ORDER BY s.is_default DESC, s.sort_order ASC, s.id ASC LIMIT 1),
		     t.completed_at = CASE WHEN ? = 'done' THEN COALESCE(t.completed_at, NOW()) ELSE NULL END
		 WHERE t.id = ?`, to, to, to, taskID); err != nil {
		logger.Ctx(ctx).Error("AlignTaskStatusToColumn failed", logger.Error(err))
		return "", "", false, err
	}
	return from, to, true, nil
}

// AlignTaskColumnToStatus moves a top-level task to the column linked to its
// status, at the bottom (sort_order 0 sorts after placed cards), when it is not
// already in a column linked to that status. A status with no column, such as
// backlog, leaves the card where it is.
func (r *Repository) AlignTaskColumnToStatus(ctx context.Context, taskID int64) error {
	if _, err := r.pool.ExecContext(ctx,
		`UPDATE flow_tasks t
		 JOIN flow_project_columns target
		   ON target.project_id = t.project_id AND target.linked_status = t.status
		 LEFT JOIN flow_project_columns cur ON cur.id = t.section_id
		 SET t.section_id = target.id, t.sort_order = 0
		 WHERE t.id = ? AND t.parent_task_id IS NULL
		   AND (cur.id IS NULL OR cur.linked_status IS NULL OR cur.linked_status <> t.status)`,
		taskID); err != nil {
		logger.Ctx(ctx).Error("AlignTaskColumnToStatus failed", logger.Error(err))
		return err
	}
	return nil
}
