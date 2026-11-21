package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"strconv"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

// epicRollupCap bounds one project's rollup. A project holds a handful of
// epics; the cap only stops a pathological one from returning without limit.
const epicRollupCap = 500

func taskTypeOrDefault(taskType string) string {
	if taskType == "" {
		return domain.TaskTypeTask
	}
	return taskType
}

// taskHierarchyConditions adds the type and epic filters to a task listing.
// They only reference "t." columns, so countTasks can share them.
func taskHierarchyConditions(filter domain.ListTasksFilter) sq.And {
	var conds sq.And
	if len(filter.TaskTypes) > 0 {
		conds = append(conds, sq.Eq{"t.task_type": filter.TaskTypes})
	}
	switch {
	case filter.EpicID != nil:
		conds = append(conds, sq.Eq{"t.epic_id": *filter.EpicID})
	case filter.NoEpic:
		conds = append(conds, sq.Expr("t.epic_id IS NULL AND t.task_type <> 'epic'"))
	}
	return conds
}

type hierarchyQuerier interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// loadHierarchyNodes reads the hierarchy columns of the given tasks inside one
// business. forUpdate locks the rows in id order, so two moves that touch the
// same pair of tasks queue instead of deadlocking.
func loadHierarchyNodes(ctx context.Context, q hierarchyQuerier, businessID int64, forUpdate bool, taskIDs ...int64) (map[int64]domain.HierarchyNode, error) {
	nodes := make(map[int64]domain.HierarchyNode, len(taskIDs))
	if len(taskIDs) == 0 {
		return nodes, nil
	}
	builder := sq.
		Select("id", "project_id", "business_id", "task_type", "parent_task_id", "epic_id").
		From("flow_tasks").
		Where(sq.Eq{"id": taskIDs, "business_id": businessID}).
		OrderBy("id").
		PlaceholderFormat(sq.Question)
	if forUpdate {
		builder = builder.Suffix("FOR UPDATE")
	}
	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var node domain.HierarchyNode
		var parentID, epicID sql.NullInt64
		if err := rows.Scan(&node.ID, &node.ProjectID, &node.BusinessID, &node.TaskType, &parentID, &epicID); err != nil {
			return nil, err
		}
		if parentID.Valid {
			node.ParentTaskID = &parentID.Int64
		}
		if epicID.Valid {
			node.EpicID = &epicID.Int64
		}
		nodes[node.ID] = node
	}
	return nodes, rows.Err()
}

func countHierarchyChildren(ctx context.Context, q hierarchyQuerier, node *domain.HierarchyNode) error {
	return q.QueryRowContext(ctx,
		`SELECT
		   (SELECT COUNT(*) FROM flow_tasks WHERE parent_task_id = ? AND business_id = ?),
		   (SELECT COUNT(*) FROM flow_tasks WHERE epic_id = ? AND business_id = ?)`,
		node.ID, node.BusinessID, node.ID, node.BusinessID,
	).Scan(&node.Subtasks, &node.EpicChildren)
}

// GetTaskHierarchyNode returns a task's place in the hierarchy, with its child
// counts. The bool is false when the business has no such task.
func (r *Repository) GetTaskHierarchyNode(ctx context.Context, taskID, businessID int64) (domain.HierarchyNode, bool, error) {
	nodes, err := loadHierarchyNodes(ctx, r.pool, businessID, false, taskID)
	if err != nil {
		logger.Ctx(ctx).Error("GetTaskHierarchyNode failed", logger.Error(err))
		return domain.HierarchyNode{}, false, err
	}
	node, ok := nodes[taskID]
	if !ok {
		return domain.HierarchyNode{}, false, nil
	}
	if err := countHierarchyChildren(ctx, r.pool, &node); err != nil {
		logger.Ctx(ctx).Error("GetTaskHierarchyNode child count failed", logger.Error(err))
		return domain.HierarchyNode{}, false, err
	}
	return node, true, nil
}

// SetTaskParent moves a task under a new parent, or to the top level when
// parentID is nil. The rules live in decide; this method only makes the read
// they are judged on and the write one unit, by locking both rows first. Two
// concurrent moves that would form a loop therefore see each other.
//
// decide receives a nil parent when parentID names no task in this business,
// and the task itself when it is named as its own parent. A task missing from
// the business reports sql.ErrNoRows.
func (r *Repository) SetTaskParent(
	ctx context.Context,
	taskID, businessID int64,
	parentID *int64,
	decide func(child domain.HierarchyNode, parent *domain.HierarchyNode) (domain.TaskPlacement, error),
) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	ids := []int64{taskID}
	if parentID != nil && *parentID != taskID {
		ids = append(ids, *parentID)
	}
	nodes, err := loadHierarchyNodes(ctx, tx, businessID, true, ids...)
	if err != nil {
		logger.Ctx(ctx).Error("SetTaskParent lock failed", logger.Error(err))
		return err
	}
	child, ok := nodes[taskID]
	if !ok {
		return sql.ErrNoRows
	}
	if err := countHierarchyChildren(ctx, tx, &child); err != nil {
		return err
	}

	var parent *domain.HierarchyNode
	if parentID != nil {
		if node, found := nodes[*parentID]; found {
			parent = &node
		}
	}
	placement, err := decide(child, parent)
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx,
		"UPDATE flow_tasks SET parent_task_id = ?, epic_id = ? WHERE id = ? AND business_id = ?",
		nullInt64Ptr(placement.ParentTaskID), nullInt64Ptr(placement.EpicID), taskID, businessID,
	); err != nil {
		logger.Ctx(ctx).Error("SetTaskParent update failed", logger.Error(err))
		return err
	}
	return tx.Commit()
}

// DeleteTaskDetachingEpicTasks deletes a task and, when it is an epic, keeps
// its tasks by clearing their epic in the same transaction. It returns how
// many were detached. Subtasks of a deleted task are left as they always were.
func (r *Repository) DeleteTaskDetachingEpicTasks(ctx context.Context, taskID, businessID int64) (int64, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.ExecContext(ctx, "DELETE FROM flow_tasks WHERE id = ? AND business_id = ?", taskID, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("DeleteTask failed", logger.Error(err))
		return 0, err
	}
	deleted, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	if deleted == 0 {
		// Nothing in this business was deleted, so nothing may be detached.
		return 0, nil
	}
	res, err = tx.ExecContext(ctx, "UPDATE flow_tasks SET epic_id = NULL WHERE epic_id = ? AND business_id = ?", taskID, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("DeleteTask detach failed", logger.Error(err))
		return 0, err
	}
	detached, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return detached, nil
}

// epicChildBucketExpr resolves a task's bucket from its custom status, falling
// back to the built-in string status for a task that has none. The fallback
// mirrors domain.BucketForTaskStatus.
const epicChildBucketExpr = `COALESCE(fs.bucket, CASE COALESCE(c.status, 'todo')
	WHEN 'backlog' THEN 'backlog' WHEN 'todo' THEN 'backlog'
	WHEN 'in_progress' THEN 'active' WHEN 'in_review' THEN 'active'
	WHEN 'done' THEN 'done' WHEN 'cancelled' THEN 'cancelled' END)`

const epicChildEstimateExpr = `COALESCE(c.effort_minutes, ROUND(c.estimated_hours * 60), 0)`

const epicRollupQuery = `SELECT e.id, e.title, COALESCE(e.epic_color, ''), COALESCE(e.status, 'todo'),
	COUNT(c.id),
	CAST(COALESCE(SUM(` + epicChildBucketExpr + ` = 'backlog'), 0) AS SIGNED),
	CAST(COALESCE(SUM(` + epicChildBucketExpr + ` = 'active'), 0) AS SIGNED),
	CAST(COALESCE(SUM(` + epicChildBucketExpr + ` = 'done'), 0) AS SIGNED),
	CAST(COALESCE(SUM(` + epicChildBucketExpr + ` = 'cancelled'), 0) AS SIGNED),
	CAST(COALESCE(SUM(` + epicChildEstimateExpr + `), 0) AS SIGNED),
	CAST(COALESCE(SUM(CASE WHEN ` + epicChildBucketExpr + ` = 'done' THEN ` + epicChildEstimateExpr + ` ELSE 0 END), 0) AS SIGNED)
FROM flow_tasks e
LEFT JOIN flow_tasks c ON c.epic_id = e.id AND c.business_id = e.business_id
	AND c.parent_task_id IS NULL AND COALESCE(c.archived, FALSE) = FALSE
LEFT JOIN flow_statuses fs ON fs.id = c.status_id AND fs.business_id = c.business_id
WHERE e.project_id = ? AND e.business_id = ? AND e.task_type = 'epic' AND COALESCE(e.archived, FALSE) = FALSE
GROUP BY e.id, e.title, e.epic_color, e.status, e.sort_order, e.created_at
ORDER BY e.sort_order ASC, e.created_at ASC, e.id ASC
LIMIT ?`

// epicLoggedMinutesQuery sums flow_time_logs over each epic's tasks and those
// tasks' subtasks. It only reads the time-log table.
const epicLoggedMinutesQuery = `SELECT x.epic_id, CAST(COALESCE(SUM(tl.duration_minutes), 0) AS SIGNED)
FROM (
	SELECT c.epic_id AS epic_id, c.id AS task_id
	FROM flow_tasks c
	JOIN flow_tasks e ON e.id = c.epic_id AND e.business_id = c.business_id
	WHERE e.project_id = ? AND e.business_id = ? AND e.task_type = 'epic'
	  AND c.parent_task_id IS NULL AND COALESCE(c.archived, FALSE) = FALSE
	UNION ALL
	SELECT c.epic_id AS epic_id, s.id AS task_id
	FROM flow_tasks s
	JOIN flow_tasks c ON c.id = s.parent_task_id AND c.business_id = s.business_id
	JOIN flow_tasks e ON e.id = c.epic_id AND e.business_id = c.business_id
	WHERE e.project_id = ? AND e.business_id = ? AND e.task_type = 'epic'
	  AND c.parent_task_id IS NULL AND COALESCE(c.archived, FALSE) = FALSE
	  AND COALESCE(s.archived, FALSE) = FALSE
) x
JOIN flow_time_logs tl ON tl.task_id = x.task_id
GROUP BY x.epic_id`

// ListEpicRollups summarises every unarchived epic of one project in two
// queries, whatever the number of epics.
func (r *Repository) ListEpicRollups(ctx context.Context, projectID, businessID int64) ([]domain.EpicRollup, error) {
	rows, err := r.pool.QueryContext(ctx, epicRollupQuery, projectID, businessID, epicRollupCap)
	if err != nil {
		logger.Ctx(ctx).Error("ListEpicRollups failed", logger.Error(err))
		return nil, err
	}
	rollups := []domain.EpicRollup{}
	index := map[int64]int{}
	for rows.Next() {
		var e domain.EpicRollup
		if err := rows.Scan(
			&e.EpicID, &e.Title, &e.Color, &e.Status,
			&e.TotalTasks, &e.BacklogTasks, &e.ActiveTasks, &e.DoneTasks, &e.CancelledTasks,
			&e.EstimateMinutes, &e.DoneEstimateMinutes,
		); err != nil {
			return nil, errors.Join(err, rows.Close())
		}
		e.Identifier = "T-" + strconv.FormatInt(e.EpicID, 10)
		index[e.EpicID] = len(rollups)
		rollups = append(rollups, e)
	}
	if err := errors.Join(rows.Err(), rows.Close()); err != nil {
		return nil, err
	}
	if len(rollups) == 0 {
		return rollups, nil
	}

	logged, err := r.pool.QueryContext(ctx, epicLoggedMinutesQuery, projectID, businessID, projectID, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("ListEpicRollups logged minutes failed", logger.Error(err))
		return nil, err
	}
	defer logged.Close()
	for logged.Next() {
		var epicID, minutes int64
		if err := logged.Scan(&epicID, &minutes); err != nil {
			return nil, err
		}
		if i, ok := index[epicID]; ok {
			rollups[i].LoggedMinutes = minutes
		}
	}
	return rollups, logged.Err()
}
