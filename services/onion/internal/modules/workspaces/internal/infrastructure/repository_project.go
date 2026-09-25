package infrastructure

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/workspaces/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

func (r *Repository) UpdateProject(ctx context.Context, input domain.UpdateProjectInput) error {
	err := r.queries.UpdateProject(ctx, sqlc.UpdateProjectParams{
		Name:        input.Name,
		Description: sql.NullString{String: input.Description, Valid: input.Description != ""},
		Icon:        sql.NullString{String: input.Icon, Valid: input.Icon != ""},
		Color:       sql.NullString{String: input.Color, Valid: input.Color != ""},
		Status:      sqlc.NullFlowProjectsStatus{FlowProjectsStatus: sqlc.FlowProjectsStatus(input.Status), Valid: input.Status != ""},
		StartDate:   nullTimePtrToSql(input.StartDate),
		DueDate:     nullTimePtrToSql(input.DueDate),
		ID:          input.ID,
		BusinessID:  input.BusinessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("UpdateProject failed", logger.Error(err))
		return err
	}
	return nil
}

// ─────────────────────────────────────────────
// Tasks
// ─────────────────────────────────────────────

func (r *Repository) CreateTask(ctx context.Context, input domain.TaskInput) (int64, error) {
	labelsJSON, err := marshalLabels(input.Labels)
	if err != nil {
		return 0, err
	}

	priority := input.Priority
	if priority == "" {
		priority = "none"
	}

	var estHours sql.NullString
	if input.EstimatedHours != nil {
		estHours = sql.NullString{String: strconv.FormatFloat(*input.EstimatedHours, 'f', 2, 64), Valid: true}
	}

	res, err := r.queries.CreateTask(ctx, sqlc.CreateTaskParams{
		ProjectID:      input.ProjectID,
		WorkspaceID:    input.WorkspaceID,
		OwnerID:        input.OwnerID,
		BusinessID:     input.BusinessID,
		ParentTaskID:   nullInt64Ptr(input.ParentTaskID),
		Title:          input.Title,
		Description:    sql.NullString{String: input.Description, Valid: input.Description != ""},
		Priority:       sqlc.NullFlowTasksPriority{FlowTasksPriority: sqlc.FlowTasksPriority(priority), Valid: true},
		AssigneeID:     nullInt64Ptr(input.AssigneeID),
		ReporterID:     input.ReporterID,
		DueDate:        nullTimePtrToSql(input.DueDate),
		StartDate:      nullTimePtrToSql(input.StartDate),
		EstimatedHours: estHours,
		Labels:         labelsJSON,
	})
	if err != nil {
		logger.Ctx(ctx).Error("CreateTask failed", logger.Error(err))
		return 0, err
	}
	return res.LastInsertId()
}

// taskListConditions builds the shared WHERE conditions for a scoped task
// list (by project or by workspace) so the same filter logic backs both the
// COUNT and the paginated SELECT.
func taskListConditions(scopeCol string, scopeVal int64, filter domain.ListTasksFilter) sq.And {
	conds := sq.And{
		sq.Eq{scopeCol: scopeVal},
		sq.Expr("t.parent_task_id IS NULL"),
	}
	if filter.BusinessID != 0 {
		conds = append(conds, sq.Eq{"t.business_id": filter.BusinessID})
	}
	if filter.Status != "" {
		conds = append(conds, sq.Eq{"t.status": filter.Status})
	} else if filter.ActiveOnly {
		conds = append(conds, sq.Expr("t.status NOT IN ('done', 'cancelled')"))
	}
	if filter.AssigneeID != nil {
		conds = append(conds, sq.Eq{"t.assignee_id": *filter.AssigneeID})
	}
	if filter.Priority != "" {
		conds = append(conds, sq.Eq{"t.priority": filter.Priority})
	}
	return conds
}

// countTasks runs a COUNT(*) over flow_tasks using the given conditions. The
// conditions built by taskListConditions only ever reference "t."-prefixed
// columns, so the count never needs the list query's project/user JOINs.
// dynamic-query: squirrel
func (r *Repository) countTasks(ctx context.Context, conds sq.And) (int64, error) {
	q, args, err := sq.Select("COUNT(*)").
		From("flow_tasks t").
		Where(conds).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("countTasks build failed", logger.Error(err))
		return 0, err
	}
	var total int64
	if err := r.pool.QueryRowContext(ctx, q, args...).Scan(&total); err != nil {
		logger.Ctx(ctx).Error("countTasks failed", logger.Error(err))
		return 0, err
	}
	return total, nil
}

// dynamic-query: squirrel
func (r *Repository) GetTasksByProject(ctx context.Context, filter domain.ListTasksFilter) ([]domain.Task, int64, error) {
	conds := taskListConditions("t.project_id", filter.ProjectID, filter)

	total, err := r.countTasks(ctx, conds)
	if err != nil {
		return nil, 0, err
	}

	builder := enrichedTaskSelectForList(filter.Compact).
		Where(conds).
		OrderBy("t.sort_order ASC", "t.created_at ASC").
		Limit(uint64(filter.Limit)).  // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Offset(uint64(filter.Offset)) // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.

	tasks, err := r.listEnrichedTasks(ctx, builder)
	if err != nil {
		return nil, 0, err
	}
	return tasks, total, nil
}

// dynamic-query: squirrel
func (r *Repository) GetTasksByWorkspace(ctx context.Context, filter domain.ListTasksFilter) ([]domain.Task, int64, error) {
	conds := taskListConditions("t.workspace_id", filter.WorkspaceID, filter)

	total, err := r.countTasks(ctx, conds)
	if err != nil {
		return nil, 0, err
	}

	builder := enrichedTaskSelectForList(filter.Compact).
		Where(conds).
		OrderBy("t.created_at DESC").
		Limit(uint64(filter.Limit)).  // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Offset(uint64(filter.Offset)) // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.

	tasks, err := r.listEnrichedTasks(ctx, builder)
	if err != nil {
		return nil, 0, err
	}
	return tasks, total, nil
}

func (r *Repository) GetTaskByID(ctx context.Context, taskID int64) (domain.Task, error) {
	tasks, err := r.listEnrichedTasks(ctx, enrichedTaskSelect().Where(sq.Eq{"t.id": taskID}))
	if err != nil {
		return domain.Task{}, err
	}
	if len(tasks) == 0 {
		return domain.Task{}, sql.ErrNoRows
	}
	return tasks[0], nil
}

// dynamic-query: squirrel
func (r *Repository) GetSubtasks(ctx context.Context, parentTaskID, limit, offset int64) ([]domain.Task, int64, error) {
	conds := sq.And{sq.Eq{"t.parent_task_id": parentTaskID}}

	total, err := r.countTasks(ctx, conds)
	if err != nil {
		return nil, 0, err
	}

	tasks, err := r.listEnrichedTasks(ctx,
		enrichedTaskSelect().
			Where(conds).
			OrderBy("t.sort_order ASC", "t.created_at ASC").
			Limit(uint64(limit)).   // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
			Offset(uint64(offset)), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	)
	if err != nil {
		return nil, 0, err
	}
	return tasks, total, nil
}

// myTasksDoneWindowDays bounds how far back completed tasks surface in the global
// "My tasks" board, so the Done column shows recent work without unbounded history.
const myTasksDoneWindowDays = 14

// dynamic-query: squirrel
//
// workspaceID scopes the result to a single workspace when non-zero. Pass 0
// for the global "My tasks" board, which intentionally spans every workspace
// the user belongs to.
func (r *Repository) GetMyTasks(ctx context.Context, userID, workspaceID, limit, offset int64) ([]domain.Task, int64, error) {
	return r.getMyTasks(ctx, userID, 0, workspaceID, limit, offset)
}

// GetMyTasksForBusiness is the tenant-safe v2 read used by Today. The legacy
// GetMyTasks method above remains available to workspace-era callers that do
// not carry a business scope; v2 callers must use this method.
func (r *Repository) GetMyTasksForBusiness(ctx context.Context, userID, businessID, workspaceID, limit, offset int64) ([]domain.Task, int64, error) {
	if businessID <= 0 {
		return nil, 0, domain.ErrUnauthorized
	}
	return r.getMyTasks(ctx, userID, businessID, workspaceID, limit, offset)
}

func (r *Repository) getMyTasks(ctx context.Context, userID, businessID, workspaceID, limit, offset int64) ([]domain.Task, int64, error) {
	// All active work assigned to the user, plus tasks they completed within the
	// recent window so the board's Done column isn't always empty. Cancelled tasks
	// stay hidden. completed_at is stamped by both the status and partial-update
	// paths when a task moves to 'done'.
	doneCutoff := time.Now().UTC().AddDate(0, 0, -myTasksDoneWindowDays)
	conds := sq.And{
		sq.Eq{"t.assignee_id": userID},
		sq.Expr("t.parent_task_id IS NULL"),
		sq.Or{
			sq.Expr("t.status NOT IN ('done','cancelled')"),
			sq.And{
				sq.Eq{"t.status": "done"},
				sq.GtOrEq{"t.completed_at": doneCutoff},
			},
		},
	}
	if workspaceID != 0 {
		conds = append(conds, sq.Eq{"t.workspace_id": workspaceID})
	}
	if businessID != 0 {
		conds = append(conds, sq.Eq{"t.business_id": businessID})
	}

	total, err := r.countTasks(ctx, conds)
	if err != nil {
		return nil, 0, err
	}

	tasks, err := r.listEnrichedTasks(ctx,
		enrichedTaskSelect().
			Where(conds).
			OrderBy(
				"CASE WHEN t.due_date IS NULL THEN 1 ELSE 0 END",
				"t.due_date ASC",
				"t.created_at DESC",
			).
			Limit(uint64(limit)).   // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
			Offset(uint64(offset)), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	)
	if err != nil {
		return nil, 0, err
	}
	return tasks, total, nil
}
