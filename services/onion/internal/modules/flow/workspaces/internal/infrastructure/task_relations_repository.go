package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/application/identity"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

// blockingPathNodeLimit bounds the cycle walk. A blocking chain this long is
// not something a board produces; past it the add is refused rather than left
// unchecked.
const blockingPathNodeLimit = 5000

// taskOpenSQL is true while a task is neither done nor cancelled. The project
// status's bucket decides when the task has one; the built-in string status,
// whose done and cancelled values carry the same names, decides otherwise.
func taskOpenSQL(taskAlias, statusAlias string) string {
	return fmt.Sprintf("COALESCE(%s.bucket, %s.status, 'todo') NOT IN ('done', 'cancelled')", statusAlias, taskAlias)
}

// openBlockerExistsSQL is true for a task row aliased t that has at least one
// open blocker in its own business.
var openBlockerExistsSQL = `EXISTS (SELECT 1 FROM flow_task_dependencies bd
	JOIN flow_tasks bt ON bt.id = bd.depends_on_task_id AND bt.business_id = bd.business_id
	LEFT JOIN flow_statuses bts ON bts.id = bt.status_id
	WHERE bd.task_id = t.id AND bd.business_id = t.business_id AND bd.link_type = 'blocks'
	  AND ` + taskOpenSQL("bt", "bts") + `)`

// AddTaskRelation stores one relation. The caller has already confirmed both
// tasks belong to businessID. A repeat of the same row is ErrRelationExists.
func (r *Repository) AddTaskRelation(ctx context.Context, edge domain.TaskRelationEdge, businessID int64) (int64, error) {
	res, err := r.pool.ExecContext(ctx,
		"INSERT INTO flow_task_dependencies (task_id, depends_on_task_id, link_type, business_id) VALUES (?, ?, ?, ?)",
		edge.TaskID, edge.DependsOnTaskID, edge.LinkType, businessID,
	)
	if err != nil {
		if isDuplicateKey(err) {
			return 0, domain.ErrRelationExists
		}
		logger.Ctx(ctx).Error("AddTaskRelation failed", logger.Error(err))
		return 0, err
	}
	return res.LastInsertId()
}

// TaskRelationExists reports whether the two tasks already hold a relation of
// this type in either direction.
func (r *Repository) TaskRelationExists(ctx context.Context, taskID, otherTaskID int64, linkType string, businessID int64) (bool, error) {
	var one int
	err := r.pool.QueryRowContext(ctx, `SELECT 1 FROM flow_task_dependencies
WHERE business_id = ? AND link_type = ?
  AND ((task_id = ? AND depends_on_task_id = ?) OR (task_id = ? AND depends_on_task_id = ?))
LIMIT 1`, businessID, linkType, taskID, otherTaskID, otherTaskID, taskID).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		logger.Ctx(ctx).Error("TaskRelationExists failed", logger.Error(err))
		return false, err
	}
	return true, nil
}

func (r *Repository) GetTaskRelation(ctx context.Context, id, businessID int64) (domain.TaskRelationEdge, error) {
	edge := domain.TaskRelationEdge{ID: id}
	err := r.pool.QueryRowContext(ctx,
		"SELECT task_id, depends_on_task_id, link_type FROM flow_task_dependencies WHERE id = ? AND business_id = ?",
		id, businessID,
	).Scan(&edge.TaskID, &edge.DependsOnTaskID, &edge.LinkType)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.TaskRelationEdge{}, domain.ErrRelationNotFound
	}
	if err != nil {
		logger.Ctx(ctx).Error("GetTaskRelation failed", logger.Error(err))
		return domain.TaskRelationEdge{}, err
	}
	return edge, nil
}

func (r *Repository) DeleteTaskRelation(ctx context.Context, id, businessID int64) error {
	res, err := r.pool.ExecContext(ctx,
		"DELETE FROM flow_task_dependencies WHERE id = ? AND business_id = ?", id, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("DeleteTaskRelation failed", logger.Error(err))
		return err
	}
	if affected, _ := res.RowsAffected(); affected == 0 {
		return domain.ErrRelationNotFound
	}
	return nil
}

// ListTaskRelations returns every relation the task is either end of, read
// from its side, oldest first. The other task must sit in the same business;
// project visibility is the caller's to apply.
func (r *Repository) ListTaskRelations(ctx context.Context, taskID, businessID int64) ([]domain.TaskRelation, error) {
	rows, err := r.pool.QueryContext(ctx, `SELECT d.id, d.task_id, d.depends_on_task_id, d.link_type,
       o.id, o.project_id, o.title, COALESCE(o.status, 'todo'), o.assignee_id,
       COALESCE(p.name, ''), COALESCE(fs.name, ''), COALESCE(fs.color, ''), COALESCE(fs.bucket, '')
FROM flow_task_dependencies d
JOIN flow_tasks o ON o.id = IF(d.task_id = ?, d.depends_on_task_id, d.task_id) AND o.business_id = d.business_id
LEFT JOIN flow_projects p ON p.id = o.project_id
LEFT JOIN flow_statuses fs ON fs.id = o.status_id
WHERE d.business_id = ? AND (d.task_id = ? OR d.depends_on_task_id = ?)
ORDER BY d.id`, taskID, businessID, taskID, taskID)
	if err != nil {
		logger.Ctx(ctx).Error("ListTaskRelations failed", logger.Error(err))
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var relations []domain.TaskRelation
	for rows.Next() {
		var edge domain.TaskRelationEdge
		var other domain.RelatedTask
		var assigneeID sql.NullInt64
		if err := rows.Scan(&edge.ID, &edge.TaskID, &edge.DependsOnTaskID, &edge.LinkType,
			&other.ID, &other.ProjectID, &other.Title, &other.Status, &assigneeID,
			&other.ProjectName, &other.StatusName, &other.StatusColor, &other.StatusBucket); err != nil {
			return nil, err
		}
		if assigneeID.Valid {
			other.AssigneeID = &assigneeID.Int64
		}
		if other.StatusBucket == "" {
			other.StatusBucket = domain.BucketForTaskStatus(other.Status)
		}
		other.Identifier = "T-" + strconv.FormatInt(other.ID, 10)
		relations = append(relations, domain.TaskRelation{
			ID:        edge.ID,
			Type:      edge.LinkType,
			Direction: edge.DirectionFrom(taskID),
			Task:      other,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}

	keys := make([]identity.IdentityKey, 0, len(relations))
	for _, relation := range relations {
		if relation.Task.AssigneeID != nil {
			keys = append(keys, identity.IdentityKey{BusinessID: businessID, UserID: *relation.Task.AssigneeID})
		}
	}
	profiles, err := r.readIdentities(ctx, keys)
	if err != nil {
		return nil, err
	}
	for i := range relations {
		if assignee := relations[i].Task.AssigneeID; assignee != nil {
			name := profiles[identity.IdentityKey{BusinessID: businessID, UserID: *assignee}].Name
			relations[i].Task.AssigneeName = name
			relations[i].Task.AssigneeInitials = initialsOf(name)
		}
	}
	return relations, nil
}

// HasBlockingPath reports whether fromTaskID is blocked, directly or through
// other tasks, by toTaskID. Adding "fromTaskID blocks toTaskID" on top of such
// a path would close a cycle.
func (r *Repository) HasBlockingPath(ctx context.Context, fromTaskID, toTaskID, businessID int64) (bool, error) {
	visited := map[int64]struct{}{fromTaskID: {}}
	frontier := []int64{fromTaskID}
	for len(frontier) > 0 {
		q, args, err := sq.Select("depends_on_task_id").
			From("flow_task_dependencies").
			Where(sq.Eq{"task_id": frontier, "business_id": businessID, "link_type": domain.RelationBlocks}).
			PlaceholderFormat(sq.Question).
			ToSql()
		if err != nil {
			return false, err
		}
		next, err := r.queryIDs(ctx, q, args...)
		if err != nil {
			logger.Ctx(ctx).Error("HasBlockingPath failed", logger.Error(err))
			return false, err
		}
		frontier = frontier[:0]
		for _, id := range next {
			if id == toTaskID {
				return true, nil
			}
			if _, seen := visited[id]; seen {
				continue
			}
			visited[id] = struct{}{}
			frontier = append(frontier, id)
		}
		if len(visited) > blockingPathNodeLimit {
			return true, nil
		}
	}
	return false, nil
}

// CountOpenBlockers returns how many open tasks block each of taskIDs, in one
// query. Tasks with none are absent from the map.
func (r *Repository) CountOpenBlockers(ctx context.Context, taskIDs []int64, businessID int64) (map[int64]int32, error) {
	out := make(map[int64]int32, len(taskIDs))
	if len(taskIDs) == 0 {
		return out, nil
	}
	q, args, err := sq.Select("d.task_id", "COUNT(*)").
		From("flow_task_dependencies d").
		Join("flow_tasks b ON b.id = d.depends_on_task_id AND b.business_id = d.business_id").
		LeftJoin("flow_statuses bs ON bs.id = b.status_id").
		Where(sq.Eq{"d.task_id": taskIDs, "d.business_id": businessID, "d.link_type": domain.RelationBlocks}).
		Where(taskOpenSQL("b", "bs")).
		GroupBy("d.task_id").
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.pool.QueryContext(ctx, q, args...)
	if err != nil {
		logger.Ctx(ctx).Error("CountOpenBlockers failed", logger.Error(err))
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var taskID int64
		var count int32
		if err := rows.Scan(&taskID, &count); err != nil {
			return nil, err
		}
		out[taskID] = count
	}
	return out, rows.Err()
}

// ListOpenBlockerIDs returns which of taskIDs are open and block at least one
// task. Read before a status write, it is the set whose closing can unblock
// something.
func (r *Repository) ListOpenBlockerIDs(ctx context.Context, taskIDs []int64, businessID int64) ([]int64, error) {
	if len(taskIDs) == 0 {
		return nil, nil
	}
	return r.listOpenBlockerIDs(ctx, sq.Eq{"b.id": taskIDs}, businessID)
}

// ListOpenBlockerIDsByStatus is ListOpenBlockerIDs for every task on one
// project status, for the write that moves them all at once.
func (r *Repository) ListOpenBlockerIDsByStatus(ctx context.Context, statusID, businessID int64) ([]int64, error) {
	return r.listOpenBlockerIDs(ctx, sq.Eq{"b.status_id": statusID}, businessID)
}

func (r *Repository) listOpenBlockerIDs(ctx context.Context, scope sq.Eq, businessID int64) ([]int64, error) {
	q, args, err := sq.Select("b.id").
		From("flow_tasks b").
		LeftJoin("flow_statuses bs ON bs.id = b.status_id").
		Where(scope).
		Where(sq.Eq{"b.business_id": businessID}).
		Where(taskOpenSQL("b", "bs")).
		Where(`EXISTS (SELECT 1 FROM flow_task_dependencies d
			WHERE d.depends_on_task_id = b.id AND d.business_id = b.business_id AND d.link_type = 'blocks')`).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}
	ids, err := r.queryIDs(ctx, q, args...)
	if err != nil {
		logger.Ctx(ctx).Error("listOpenBlockerIDs failed", logger.Error(err))
	}
	return ids, err
}

// ListUnblockedTasks returns the open tasks that were blocked by one of
// blockerIDs, where that blocker is now closed and no open blocker remains.
// A task several of the blockers were holding comes back once.
func (r *Repository) ListUnblockedTasks(ctx context.Context, blockerIDs []int64, businessID int64) ([]domain.UnblockedTask, error) {
	if len(blockerIDs) == 0 {
		return nil, nil
	}
	q, args, err := sq.Select("DISTINCT t.id", "t.project_id", "t.workspace_id", "t.owner_id", "t.title",
		"COALESCE(t.assignee_id, 0)", "COALESCE(p.name, '')", "COALESCE(p.color, '')").
		From("flow_task_dependencies d").
		Join("flow_tasks c ON c.id = d.depends_on_task_id AND c.business_id = d.business_id").
		LeftJoin("flow_statuses cs ON cs.id = c.status_id").
		Join("flow_tasks t ON t.id = d.task_id AND t.business_id = d.business_id").
		LeftJoin("flow_statuses ts ON ts.id = t.status_id").
		LeftJoin("flow_projects p ON p.id = t.project_id").
		Where(sq.Eq{"d.depends_on_task_id": blockerIDs, "d.business_id": businessID, "d.link_type": domain.RelationBlocks}).
		Where("NOT (" + taskOpenSQL("c", "cs") + ")").
		Where(taskOpenSQL("t", "ts")).
		Where("NOT " + openBlockerExistsSQL).
		OrderBy("t.id").
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.pool.QueryContext(ctx, q, args...)
	if err != nil {
		logger.Ctx(ctx).Error("ListUnblockedTasks failed", logger.Error(err))
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var tasks []domain.UnblockedTask
	for rows.Next() {
		var task domain.UnblockedTask
		if err := rows.Scan(&task.ID, &task.ProjectID, &task.WorkspaceID, &task.OwnerID, &task.Title,
			&task.AssigneeID, &task.ProjectName, &task.ProjectColor); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

func (r *Repository) queryIDs(ctx context.Context, query string, args ...any) ([]int64, error) {
	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
