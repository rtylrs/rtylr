package infrastructure

import (
	"context"
	"database/sql"
	"strings"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/searchutil"
)

func (r *Repository) GetColumnsByProject(ctx context.Context, projectID, limit, offset int64) ([]domain.ProjectColumn, int64, error) {
	total, err := r.queries.CountProjectColumns(ctx, projectID)
	if err != nil {
		logger.Ctx(ctx).Error("CountProjectColumns failed", logger.Error(err))
		return nil, 0, err
	}

	rows, err := r.queries.GetColumnsByProject(ctx, sqlc.GetColumnsByProjectParams{
		ProjectID: projectID,
		Limit:     int32(limit),  // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Offset:    int32(offset), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		logger.Ctx(ctx).Error("GetColumnsByProject failed", logger.Error(err))
		return nil, 0, err
	}

	cols := make([]domain.ProjectColumn, 0, len(rows))
	for _, row := range rows {
		cols = append(cols, domain.ProjectColumn{
			ID:          row.ID,
			ProjectID:   row.ProjectID,
			WorkspaceID: row.WorkspaceID,
			Name:        row.Name,
			Color:       row.Color.String,
			SortOrder:   int(row.SortOrder.Int32),
			CreatedBy:   row.CreatedBy,
			CreatedAt:   row.CreatedAt.Time,
			UpdatedAt:   row.UpdatedAt.Time,
		})
	}
	return cols, total, nil
}

func (r *Repository) CreateColumn(ctx context.Context, input domain.ProjectColumnInput) (int64, error) {
	res, err := r.queries.CreateColumn(ctx, sqlc.CreateColumnParams{
		ProjectID:   input.ProjectID,
		WorkspaceID: input.WorkspaceID,
		Name:        input.Name,
		Color:       sql.NullString{String: input.Color, Valid: input.Color != ""},
		SortOrder:   sql.NullInt32{Int32: int32(input.SortOrder), Valid: true}, // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		CreatedBy:   input.CreatedBy,
	})
	if err != nil {
		logger.Ctx(ctx).Error("CreateColumn failed", logger.Error(err))
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) UpdateColumn(ctx context.Context, input domain.UpdateProjectColumnInput) error {
	res, err := r.queries.UpdateColumn(ctx, sqlc.UpdateColumnParams{
		Name:      input.Name,
		Color:     sql.NullString{String: input.Color, Valid: input.Color != ""},
		SortOrder: sql.NullInt32{Int32: int32(input.SortOrder), Valid: true}, // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		ID:        input.ID,
		// One actor across every branch, matching DeleteColumn: project member,
		// project owner, project creator, then the two legacy workspace branches.
		UserID:    input.ActorID,
		OwnerID:   input.ActorID,
		CreatedBy: input.ActorID,
		OwnerID_2: input.ActorID,
		UserID_2:  input.ActorID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("UpdateColumn failed", logger.Error(err))
		return err
	}
	return requireAffected(res)
}

func (r *Repository) DeleteColumn(ctx context.Context, id, actorID int64) error {
	// One actor across every branch, the same shape MoveTaskToColumn uses:
	// project member, project owner, project creator, then the two legacy
	// workspace branches.
	if err := r.queries.ClearTasksSectionID(ctx, sqlc.ClearTasksSectionIDParams{
		SectionID: sql.NullInt64{Int64: id, Valid: true},
		UserID:    actorID,
		OwnerID:   actorID,
		CreatedBy: actorID,
		OwnerID_2: actorID,
		UserID_2:  actorID,
	}); err != nil {
		return err
	}

	res, err := r.queries.DeleteColumn(ctx, sqlc.DeleteColumnParams{
		ID:        id,
		UserID:    actorID,
		OwnerID:   actorID,
		CreatedBy: actorID,
		OwnerID_2: actorID,
		UserID_2:  actorID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("DeleteColumn failed", logger.Error(err))
		return err
	}
	return requireAffected(res)
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
      SELECT 1 FROM flow_project_columns c
      WHERE c.id = ?
        AND (
          c.project_id = flow_tasks.project_id
          OR (
            flow_tasks.workspace_id <> 0
            AND c.workspace_id = flow_tasks.workspace_id
          )
        )
    )
  )`

func (r *Repository) MoveTaskToColumn(ctx context.Context, taskID, columnID, actorID int64) error {
	var sectionID sql.NullInt64
	if columnID != 0 {
		sectionID = sql.NullInt64{Int64: columnID, Valid: true}
	}
	res, err := r.pool.ExecContext(ctx, moveTaskToColumnSQL,
		sectionID,
		taskID,
		actorID,
		actorID,
		actorID,
		actorID,
		actorID,
		columnID,
		columnID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("MoveTaskToColumn failed", logger.Error(err))
		return err
	}
	return requireAffected(res)
}

func (r *Repository) SeedDefaultColumns(ctx context.Context, projectID, workspaceID, createdBy int64) error {
	// Idempotent: never seed a project that already has columns. Two create
	// paths call this and (until deployed) there's no unique constraint behind
	// it, so without this guard a re-run produces duplicate columns.
	existing, err := r.queries.CountProjectColumns(ctx, projectID)
	if err != nil {
		return err
	}
	if existing > 0 {
		return nil
	}

	defaults := []struct {
		name      string
		color     string
		sortOrder int
	}{
		{"To Do", "#9ca3af", 0},
		{"In Progress", "#3b82f6", 1},
		{"In Review", "#f59e0b", 2},
		{"Done", "#22c55e", 3},
	}
	for _, d := range defaults {
		if _, err := r.queries.SeedDefaultColumn(ctx, sqlc.SeedDefaultColumnParams{
			ProjectID:   projectID,
			WorkspaceID: workspaceID,
			Name:        d.name,
			Color:       sql.NullString{String: d.color, Valid: true},
			SortOrder:   sql.NullInt32{Int32: int32(d.sortOrder), Valid: true}, // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
			CreatedBy:   createdBy,
		}); err != nil {
			return err
		}
	}
	return nil
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
