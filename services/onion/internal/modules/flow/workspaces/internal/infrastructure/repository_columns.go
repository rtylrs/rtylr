package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/searchutil"
)

// Columns are the project's statuses (see status_columns.go). The column API
// keeps its shape so existing clients work unchanged; a column id is a status
// id.

// projectWriteGuard is the column write guard the column endpoints always had:
// project member, owner or creator. %s is the project_id expression.
const projectWriteGuard = `(
    EXISTS (SELECT 1 FROM flow_project_members pm WHERE pm.project_id = %[1]s AND pm.user_id = ?)
    OR EXISTS (SELECT 1 FROM flow_projects p WHERE p.id = %[1]s AND (p.owner_id = ? OR p.created_by = ?))
  )`

func (r *Repository) GetColumnsByProject(ctx context.Context, projectID, limit, offset int64) ([]domain.ProjectColumn, int64, error) {
	var ownerID int64
	if err := r.pool.QueryRowContext(ctx, "SELECT owner_id FROM flow_projects WHERE id = ?", projectID).Scan(&ownerID); err == nil {
		if seedErr := r.seedProjectStatuses(ctx, projectID, ownerID); seedErr != nil {
			logger.Ctx(ctx).Error("GetColumnsByProject: seeding statuses failed", logger.Error(seedErr))
		}
	}

	var total int64
	if err := r.pool.QueryRowContext(ctx, "SELECT COUNT(*) FROM flow_statuses WHERE project_id = ?", projectID).Scan(&total); err != nil {
		logger.Ctx(ctx).Error("GetColumnsByProject count failed", logger.Error(err))
		return nil, 0, err
	}
	rows, err := r.pool.QueryContext(ctx, `SELECT id, project_id, name, color, sort_order, created_by, created_at, updated_at
FROM flow_statuses WHERE project_id = ? ORDER BY sort_order, id LIMIT ? OFFSET ?`, projectID, limit, offset)
	if err != nil {
		logger.Ctx(ctx).Error("GetColumnsByProject failed", logger.Error(err))
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	cols := make([]domain.ProjectColumn, 0)
	for rows.Next() {
		var c domain.ProjectColumn
		var createdAt, updatedAt sql.NullTime
		if err := rows.Scan(&c.ID, &c.ProjectID, &c.Name, &c.Color, &c.SortOrder, &c.CreatedBy, &createdAt, &updatedAt); err != nil {
			return nil, 0, err
		}
		c.CreatedAt, c.UpdatedAt = createdAt.Time, updatedAt.Time
		cols = append(cols, c)
	}
	return cols, total, rows.Err()
}

// CreateColumn adds an in-progress-type status, which is what a new board
// column means. The bucket can be changed afterwards in the status settings.
func (r *Repository) CreateColumn(ctx context.Context, input domain.ProjectColumnInput) (int64, error) {
	businessID, err := projectBusinessID(ctx, r.pool, input.ProjectID)
	if err != nil {
		return 0, err
	}
	color := input.Color
	if color == "" {
		color = domain.DefaultStatusColor
	}
	res, err := r.pool.ExecContext(ctx, `INSERT INTO flow_statuses
(business_id, project_id, name, bucket, legacy_status, color, sort_order, is_default, created_by)
VALUES (?, ?, ?, 'active', 'in_progress', ?, ?, FALSE, ?)`,
		businessID, input.ProjectID, strings.TrimSpace(input.Name), color, input.SortOrder, input.CreatedBy)
	if err != nil {
		if isDuplicateKey(err) {
			return 0, domain.ErrStatusNameTaken
		}
		logger.Ctx(ctx).Error("CreateColumn failed", logger.Error(err))
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) UpdateColumn(ctx context.Context, input domain.UpdateProjectColumnInput) error {
	res, err := r.pool.ExecContext(ctx, `UPDATE flow_statuses
SET name = ?, color = COALESCE(NULLIF(?, ''), color), sort_order = ?
WHERE flow_statuses.id = ? AND `+fmt.Sprintf(projectWriteGuard, "flow_statuses.project_id"),
		strings.TrimSpace(input.Name), input.Color, input.SortOrder, input.ID, input.ActorID, input.ActorID, input.ActorID)
	if err != nil {
		if isDuplicateKey(err) {
			return domain.ErrStatusNameTaken
		}
		logger.Ctx(ctx).Error("UpdateColumn failed", logger.Error(err))
		return err
	}
	if affected, _ := res.RowsAffected(); affected == 0 {
		// Zero rows is either "no access" or "nothing changed"; only the first
		// is an error.
		var one int
		err := r.pool.QueryRowContext(ctx, "SELECT 1 FROM flow_statuses WHERE flow_statuses.id = ? AND "+
			fmt.Sprintf(projectWriteGuard, "flow_statuses.project_id"), input.ID, input.ActorID, input.ActorID, input.ActorID).Scan(&one)
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrUnauthorized
		}
		return err
	}
	return nil
}

// DeleteColumn deletes an empty status. The column endpoint has no way to say
// where tasks should go, so a status with tasks is refused here; the status
// endpoint takes a move target.
func (r *Repository) DeleteColumn(ctx context.Context, id, actorID int64) error {
	var businessID int64
	err := r.pool.QueryRowContext(ctx, "SELECT business_id FROM flow_statuses WHERE flow_statuses.id = ? AND "+
		fmt.Sprintf(projectWriteGuard, "flow_statuses.project_id"), id, actorID, actorID, actorID).Scan(&businessID)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrUnauthorized
	}
	if err != nil {
		return err
	}
	_, err = r.DeleteStatus(ctx, id, businessID, 0)
	return err
}

const moveTaskToColumnSQL = `UPDATE flow_tasks
SET section_id = ?, updated_at = NOW()
WHERE flow_tasks.id = ?
  AND (
    EXISTS (SELECT 1 FROM flow_project_members pm
            WHERE pm.project_id = flow_tasks.project_id AND pm.user_id = ?)
    OR EXISTS (SELECT 1 FROM flow_projects p
               WHERE p.id = flow_tasks.project_id
                 AND (p.owner_id = ? OR p.created_by = ?))
    OR EXISTS (SELECT 1 FROM flow_workspaces w WHERE w.id = flow_tasks.workspace_id AND w.owner_id = ?)
    OR EXISTS (SELECT 1 FROM flow_workspace_members wm WHERE wm.workspace_id = flow_tasks.workspace_id AND wm.user_id = ?)
  )
  AND (
    ? = 0
    OR EXISTS (
      SELECT 1 FROM flow_statuses c
      WHERE c.id = ? AND c.project_id = flow_tasks.project_id
    )
  )`

func (r *Repository) MoveTaskToColumn(ctx context.Context, taskID, columnID, actorID int64) error {
	res, err := r.pool.ExecContext(ctx, moveTaskToColumnSQL, moveTaskToColumnArgs(taskID, columnID, actorID)...)
	if err != nil {
		logger.Ctx(ctx).Error("MoveTaskToColumn failed", logger.Error(err))
		return err
	}
	if err := requireAffected(res); err != nil {
		return err
	}
	return syncTaskFromSection(ctx, r.pool, taskID)
}

// moveTaskAccessSQL is moveTaskToColumnSQL's WHERE clause as a SELECT, taking
// the same arguments minus the leading section_id. The ordered path uses it to
// tell "not permitted" from "nothing to change": a same-column reorder inside
// the second the row was last touched updates zero rows, which the plain
// affected-row check would misreport as unauthorized.
var moveTaskAccessSQL = "SELECT 1 FROM flow_tasks " + moveTaskToColumnSQL[strings.Index(moveTaskToColumnSQL, "WHERE"):] + " FOR UPDATE"

func moveTaskToColumnArgs(taskID, columnID, actorID int64) []any {
	var sectionID sql.NullInt64
	if columnID != 0 {
		sectionID = sql.NullInt64{Int64: columnID, Valid: true}
	}
	return []any{sectionID, taskID, actorID, actorID, actorID, actorID, actorID, columnID, columnID}
}

// MoveTaskToColumnOrdered moves the task exactly like MoveTaskToColumn and, in
// the same transaction, stores the card order: sort_order = index+1 for every
// id in orderedTaskIDs. An empty list is the plain move. Every listed id must
// belong to the moved task's project; if any does not, the whole call (move
// included) is rolled back with ErrTaskOrderOutsideProject.
//
// Access control: the move statement keeps its own access check, and the
// caller has already passed RequireTaskRole on taskID. Because every reordered
// id is constrained to taskID's project, that check covers all of them.
//
// Only the moved task gets updated_at = NOW() (from the move statement); the
// other cards keep their updated_at, since the column has ON UPDATE
// CURRENT_TIMESTAMP and a bare sort_order change would otherwise bump it.
func (r *Repository) MoveTaskToColumnOrdered(ctx context.Context, taskID, columnID, actorID int64, orderedTaskIDs []int64) error {
	if len(orderedTaskIDs) == 0 {
		return r.MoveTaskToColumn(ctx, taskID, columnID, actorID)
	}
	if err := domain.ValidateTaskOrder(taskID, orderedTaskIDs); err != nil {
		return err
	}

	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.ExecContext(ctx, moveTaskToColumnSQL, moveTaskToColumnArgs(taskID, columnID, actorID)...)
	if err != nil {
		logger.Ctx(ctx).Error("MoveTaskToColumnOrdered: move failed", logger.Error(err))
		return err
	}
	if err := requireAffected(res); err != nil {
		if !errors.Is(err, domain.ErrUnauthorized) {
			return err
		}
		var one int
		scanErr := tx.QueryRowContext(ctx, moveTaskAccessSQL, moveTaskToColumnArgs(taskID, columnID, actorID)[1:]...).Scan(&one)
		if errors.Is(scanErr, sql.ErrNoRows) {
			return domain.ErrUnauthorized
		}
		if scanErr != nil {
			return scanErr
		}
	}

	var projectID int64
	if err := tx.QueryRowContext(ctx, "SELECT project_id FROM flow_tasks WHERE id = ?", taskID).Scan(&projectID); err != nil {
		return err
	}

	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(orderedTaskIDs)), ",")
	inArgs := make([]any, 0, len(orderedTaskIDs)+1)
	for _, id := range orderedTaskIDs {
		inArgs = append(inArgs, id)
	}
	inArgs = append(inArgs, projectID)

	// RowsAffected only counts changed rows, so count the matches explicitly.
	var matched int
	countSQL := "SELECT COUNT(*) FROM flow_tasks WHERE id IN (" + placeholders + ") AND project_id = ? FOR UPDATE"
	if err := tx.QueryRowContext(ctx, countSQL, inArgs...).Scan(&matched); err != nil {
		return err
	}
	if matched != len(orderedTaskIDs) {
		return domain.ErrTaskOrderOutsideProject
	}

	// Placeholders only; ids and positions are bound, never concatenated.
	var cases strings.Builder
	args := make([]any, 0, 3*len(orderedTaskIDs)+1)
	for i, id := range orderedTaskIDs {
		cases.WriteString(" WHEN ? THEN ?")
		args = append(args, id, i+1)
	}
	args = append(args, inArgs...)
	updateSQL := "UPDATE flow_tasks SET sort_order = CASE id" + cases.String() +
		" END, updated_at = updated_at WHERE id IN (" + placeholders + ") AND project_id = ?"
	if _, err := tx.ExecContext(ctx, updateSQL, args...); err != nil {
		logger.Ctx(ctx).Error("MoveTaskToColumnOrdered: reorder failed", logger.Error(err))
		return err
	}
	if err := syncTaskFromSection(ctx, tx, taskID); err != nil {
		return err
	}
	return tx.Commit()
}

// ─────────────────────────────────────────────
// Task Search
// ─────────────────────────────────────────────

// dynamic-query: squirrel
func (r *Repository) SearchTasks(ctx context.Context, workspaceID int64, q string, limit, offset int64) ([]domain.Task, int64, error) {
	q = strings.TrimSpace(q)
	if q == "" {
		return []domain.Task{}, 0, nil
	}
	searchQuery := searchutil.BooleanPrefixQuery(q)
	if searchQuery == "" {
		return []domain.Task{}, 0, nil
	}

	countQuery := `SELECT COUNT(*)
	          FROM flow_tasks
	          WHERE workspace_id = ?
	            AND archived = FALSE
	            AND MATCH(title, description) AGAINST (? IN BOOLEAN MODE)`
	var total int64
	if err := r.pool.QueryRowContext(ctx, countQuery, workspaceID, searchQuery).Scan(&total); err != nil {
		logger.Ctx(ctx).Error("SearchTasks count failed", logger.Error(err))
		return nil, 0, err
	}

	query := `SELECT id, project_id, workspace_id, owner_id, parent_task_id, section_id, title, description,
	           status, priority, assignee_id, reporter_id, due_date, start_date, completed_at,
	           estimated_hours, actual_hours, sort_order, labels, COALESCE(archived, FALSE) AS archived,
	           created_at, updated_at
	          FROM flow_tasks
	          WHERE workspace_id = ?
	            AND archived = FALSE
	            AND MATCH(title, description) AGAINST (? IN BOOLEAN MODE)
	          ORDER BY updated_at DESC
	          LIMIT ? OFFSET ?`

	rows, err := r.pool.QueryContext(ctx, query, workspaceID, searchQuery, limit, offset)
	if err != nil {
		logger.Ctx(ctx).Error("SearchTasks failed", logger.Error(err))
		return nil, 0, err
	}
	defer rows.Close()

	tasks, err := scanTasksRaw(ctx, rows)
	if err != nil {
		return nil, 0, err
	}
	return tasks, total, nil
}
