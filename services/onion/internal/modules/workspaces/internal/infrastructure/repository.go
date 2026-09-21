package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/workspaces/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/searchutil"
	"github.com/voxire/onion/internal/platform/utils/times"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	pool := db
	return &Repository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

// ─────────────────────────────────────────────
// Workspaces
// ─────────────────────────────────────────────

func (r *Repository) CountWorkspacesByOwner(ctx context.Context, businessID int64) (int64, error) {
	count, err := r.queries.CountWorkspacesByOwner(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("CountWorkspacesByOwner failed", logger.Error(err))
		return 0, err
	}
	return count, nil
}

func (r *Repository) CreateWorkspace(ctx context.Context, input domain.WorkspaceInput, isDefault bool) (int64, error) {
	icon := input.Icon
	if icon == "" {
		icon = "🚀"
	}
	color := input.Color
	if color == "" {
		color = "#4583C4"
	}

	res, err := r.queries.CreateWorkspace(ctx, sqlc.CreateWorkspaceParams{
		OwnerID:     input.OwnerID,
		BusinessID:  input.BusinessID,
		Name:        input.Name,
		Description: sql.NullString{String: input.Description, Valid: input.Description != ""},
		Icon:        sql.NullString{String: icon, Valid: true},
		Color:       sql.NullString{String: color, Valid: true},
		IsDefault:   sql.NullBool{Bool: isDefault, Valid: true},
		CreatedBy:   input.CreatedBy,
	})
	if err != nil {
		logger.Ctx(ctx).Error("CreateWorkspace failed", logger.Error(err))
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) GetWorkspacesByOwner(ctx context.Context, businessID, limit, offset int64) ([]domain.Workspace, int64, error) {
	total, err := r.queries.CountWorkspacesByOwner(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("CountWorkspacesByOwner failed", logger.Error(err))
		return nil, 0, err
	}

	rows, err := r.queries.GetWorkspacesByOwner(ctx, sqlc.GetWorkspacesByOwnerParams{
		BusinessID: businessID,
		Limit:      int32(limit),  // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Offset:     int32(offset), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		logger.Ctx(ctx).Error("GetWorkspacesByOwner failed", logger.Error(err))
		return nil, 0, err
	}

	workspaces := make([]domain.Workspace, 0, len(rows))
	for _, row := range rows {
		workspaces = append(workspaces, domain.Workspace{
			ID:          row.ID,
			BusinessID:  row.BusinessID,
			OwnerID:     row.OwnerID,
			Name:        row.Name,
			Description: row.Description.String,
			Icon:        row.Icon.String,
			Color:       row.Color.String,
			IsDefault:   row.IsDefault.Bool,
			CreatedBy:   row.CreatedBy,
			CreatedAt:   row.CreatedAt.Time,
			UpdatedAt:   row.UpdatedAt.Time,
		})
	}
	return workspaces, total, nil
}

func (r *Repository) UpdateWorkspace(ctx context.Context, input domain.UpdateWorkspaceInput) error {
	err := r.queries.UpdateWorkspace(ctx, sqlc.UpdateWorkspaceParams{
		Name:        input.Name,
		Description: sql.NullString{String: input.Description, Valid: input.Description != ""},
		Icon:        sql.NullString{String: input.Icon, Valid: input.Icon != ""},
		Color:       sql.NullString{String: input.Color, Valid: input.Color != ""},
		ID:          input.ID,
		BusinessID:  input.BusinessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("UpdateWorkspace failed", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) DeleteWorkspace(ctx context.Context, id, businessID int64) error {
	err := r.queries.DeleteWorkspace(ctx, sqlc.DeleteWorkspaceParams{
		ID:         id,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("DeleteWorkspace failed", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) AddWorkspaceMember(ctx context.Context, input domain.AddWorkspaceMemberInput) error {
	role := input.Role
	if role == "" {
		role = "member"
	}

	err := r.queries.AddWorkspaceMember(ctx, sqlc.AddWorkspaceMemberParams{
		WorkspaceID: input.WorkspaceID,
		UserID:      input.UserID,
		Role:        sqlc.NullFlowWorkspaceMembersRole{FlowWorkspaceMembersRole: sqlc.FlowWorkspaceMembersRole(role), Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("AddWorkspaceMember failed", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) RemoveWorkspaceMember(ctx context.Context, input domain.RemoveWorkspaceMemberInput) error {
	err := r.queries.RemoveWorkspaceMember(ctx, sqlc.RemoveWorkspaceMemberParams{
		WorkspaceID: input.WorkspaceID,
		UserID:      input.UserID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("RemoveWorkspaceMember failed", logger.Error(err))
		return err
	}
	return nil
}

// ─────────────────────────────────────────────
// Projects
// ─────────────────────────────────────────────

func (r *Repository) CreateProject(ctx context.Context, input domain.ProjectInput) (int64, error) {
	icon := input.Icon
	if icon == "" {
		icon = "📋"
	}
	color := input.Color
	if color == "" {
		color = "#4583C4"
	}

	visibility := input.Visibility
	if visibility == "" {
		visibility = domain.VisibilityPrivate
	}
	res, err := r.queries.CreateProject(ctx, sqlc.CreateProjectParams{
		WorkspaceID: input.WorkspaceID,
		OwnerID:     input.OwnerID,
		BusinessID:  input.BusinessID,
		Visibility:  sqlc.FlowProjectsVisibility(visibility),
		IsPersonal:  input.IsPersonal,
		Name:        input.Name,
		Description: sql.NullString{String: input.Description, Valid: input.Description != ""},
		Icon:        sql.NullString{String: icon, Valid: true},
		Color:       sql.NullString{String: color, Valid: true},
		StartDate:   nullTimePtrToSql(input.StartDate),
		DueDate:     nullTimePtrToSql(input.DueDate),
		CreatedBy:   input.CreatedBy,
	})
	if err != nil {
		logger.Ctx(ctx).Error("CreateProject failed", logger.Error(err))
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) GetProjectsByWorkspace(ctx context.Context, workspaceID, businessID, limit, offset int64) ([]domain.Project, int64, error) {
	total, err := r.queries.CountProjectsByWorkspace(ctx, sqlc.CountProjectsByWorkspaceParams{
		WorkspaceID: workspaceID,
		BusinessID:  businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("CountProjectsByWorkspace failed", logger.Error(err))
		return nil, 0, err
	}

	rows, err := r.queries.GetProjectsByWorkspace(ctx, sqlc.GetProjectsByWorkspaceParams{
		WorkspaceID: workspaceID,
		BusinessID:  businessID,
		Limit:       int32(limit),  // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Offset:      int32(offset), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		logger.Ctx(ctx).Error("GetProjectsByWorkspace failed", logger.Error(err))
		return nil, 0, err
	}

	projects := make([]domain.Project, 0, len(rows))
	for _, row := range rows {
		p := domain.Project{
			ID:             row.ID,
			WorkspaceID:    row.WorkspaceID,
			OwnerID:        row.OwnerID,
			BusinessID:     row.BusinessID,
			Name:           row.Name,
			Description:    row.Description,
			Icon:           row.Icon,
			Color:          row.Color,
			Status:         string(row.Status),
			CreatedBy:      row.CreatedBy,
			TaskCount:      int32(countIfaceToInt64(row.TaskCount)),      // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
			CompletedCount: int32(countIfaceToInt64(row.CompletedCount)), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
			MemberCount:    int32(countIfaceToInt64(row.MemberCount)),    // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		}
		if row.StartDate.Valid {
			p.StartDate = &row.StartDate.Time
		}
		if row.DueDate.Valid {
			p.DueDate = &row.DueDate.Time
		}
		if row.CreatedAt.Valid {
			p.CreatedAt = row.CreatedAt.Time
		}
		if row.UpdatedAt.Valid {
			p.UpdatedAt = row.UpdatedAt.Time
		}
		projects = append(projects, p)
	}
	return projects, total, nil
}

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

// dynamic-query: squirrel
func (r *Repository) UpdateTask(ctx context.Context, input domain.UpdateTaskInput) error {
	var statusParam interface{}
	if input.Status != "" {
		statusParam = input.Status
	}
	var priorityParam interface{}
	if input.Priority != "" {
		priorityParam = input.Priority
	}
	// Guard on the input slice, not on the marshalled bytes. marshalLabels turns
	// a nil slice into "[]", which is never nil, so testing its output made
	// labelsParam non-NULL on every call and overwrote labels on updates that
	// never mentioned them. Only a nil slice means "leave the column alone"; an
	// explicit empty slice means "clear every label" and must still be written.
	// marshalLabels is left as-is because CreateTask depends on it defaulting a
	// nil slice to "[]" for the INSERT.
	var labelsParam interface{}
	if input.Labels != nil {
		labelsJSON, err := marshalLabels(input.Labels)
		if err != nil {
			return err
		}
		labelsParam = labelsJSON
	}
	var titleParam interface{}
	if input.Title != "" {
		titleParam = input.Title
	}
	var assigneeParam interface{}
	if input.AssigneeID != nil {
		assigneeParam = *input.AssigneeID
	}
	var dueDateParam interface{}
	if input.DueDate != nil {
		dueDateParam = *input.DueDate
	}
	var startDateParam interface{}
	if input.StartDate != nil {
		startDateParam = *input.StartDate
	}
	var archivedParam interface{}
	if input.Archived != nil {
		archivedParam = *input.Archived
	}

	// These five use a conditional Set rather than COALESCE so that a caller can
	// distinguish "leave it alone" (nil) from "clear it" (zero value). The
	// COALESCE fields above cannot express a clear, which is the reason a task
	// description can never be emptied once set.
	builder := sq.Update("flow_tasks")
	if input.EffortMinutes != nil {
		if *input.EffortMinutes == 0 {
			builder = builder.Set("effort_minutes", nil)
		} else {
			builder = builder.Set("effort_minutes", *input.EffortMinutes)
		}
	}
	if input.EstimatedHours != nil {
		if *input.EstimatedHours == 0 {
			builder = builder.Set("estimated_hours", nil)
		} else {
			// Format identically to CreateTask, deliberately: estimated_hours is a
			// DECIMAL(6,2) column, and leaving this as a raw float64 would give the
			// column two different marshalling paths that could drift from each
			// other depending on the driver's float formatting.
			builder = builder.Set("estimated_hours", strconv.FormatFloat(*input.EstimatedHours, 'f', 2, 64))
		}
	}
	if input.WaitingOn != nil {
		if *input.WaitingOn == "" {
			builder = builder.Set("waiting_on", nil)
		} else {
			builder = builder.Set("waiting_on", *input.WaitingOn)
		}
	}
	if input.DecisionNote != nil {
		if *input.DecisionNote == "" {
			builder = builder.Set("decision_note", nil)
		} else {
			builder = builder.Set("decision_note", *input.DecisionNote)
		}
	}

	// Partial updates (the client sends only the fields it changed, e.g. just
	// priority or assignee) must not overwrite the fields it omitted. COALESCE(?, col)
	// preserves the existing value when the param is NULL, and every nullable param
	// above is left NULL when the request omits that field. The completed_at CASE only
	// fires when a non-NULL status is explicitly supplied.
	q, args, err := builder.
		Set("title", sq.Expr("COALESCE(?, title)", titleParam)).
		Set("description", sq.Expr("COALESCE(?, description)", sql.NullString{String: input.Description, Valid: input.Description != ""})).
		Set("status", sq.Expr("COALESCE(?, status)", statusParam)).
		Set("priority", sq.Expr("COALESCE(?, priority)", priorityParam)).
		Set("assignee_id", sq.Expr("COALESCE(?, assignee_id)", assigneeParam)).
		Set("due_date", sq.Expr("COALESCE(?, due_date)", dueDateParam)).
		Set("start_date", sq.Expr("COALESCE(?, start_date)", startDateParam)).
		Set("labels", sq.Expr("COALESCE(?, labels)", labelsParam)).
		Set("archived", sq.Expr("COALESCE(?, archived)", archivedParam)).
		Set("completed_at", sq.Expr(
			`CASE
		        WHEN ? IS NOT NULL AND ? = 'done'  AND completed_at IS NULL THEN NOW()
		        WHEN ? IS NOT NULL AND ? != 'done' THEN NULL
		        ELSE completed_at
		    END`,
			statusParam, statusParam, // CASE WHEN ? IS NOT NULL AND ? = 'done'
			statusParam, statusParam, // CASE WHEN ? IS NOT NULL AND ? != 'done'
		)).
		Where(sq.Eq{"id": input.ID, "business_id": input.BusinessID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("UpdateTask build failed", logger.Error(err))
		return err
	}

	_, err = r.pool.ExecContext(ctx, q, args...)
	if err != nil {
		logger.Ctx(ctx).Error("UpdateTask failed", logger.Error(err))
		return err
	}
	return nil
}

// SetTaskFocusPin pins or unpins a task in one user's Today view. The pin is
// per-user by design: Today is a personal view, so a column on flow_tasks would
// put one member's planning in another member's list.
//
// businessID is written as a plain column value here, not a WHERE guard, so
// this method trusts the caller to have already confirmed the task belongs to
// that business. The API handler enforces that via RequireTaskRole before
// calling in.
func (r *Repository) SetTaskFocusPin(ctx context.Context, taskID, userID, businessID int64, pinned bool) error {
	if !pinned {
		_, err := r.pool.ExecContext(ctx,
			"DELETE FROM flow_task_focus_pins WHERE task_id = ? AND user_id = ? AND business_id = ?",
			taskID, userID, businessID,
		)
		if err != nil {
			logger.Ctx(ctx).Error("SetTaskFocusPin delete failed", logger.Error(err))
		}
		return err
	}
	// INSERT IGNORE rather than a read-then-write: the unique key on
	// (task_id, user_id) makes a repeat pin a no-op without a round trip.
	_, err := r.pool.ExecContext(ctx,
		"INSERT IGNORE INTO flow_task_focus_pins (task_id, user_id, business_id) VALUES (?, ?, ?)",
		taskID, userID, businessID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("SetTaskFocusPin insert failed", logger.Error(err))
	}
	return err
}

// AddTaskDependency records that taskID is blocked by dependsOnTaskID. The
// unique key on (task_id, depends_on_task_id) makes a repeat add a no-op.
//
// input.BusinessID is written as a plain column value here, not a WHERE
// guard, so this method trusts the caller to have already confirmed both
// tasks belong to that business (the same trust CreateTask already places in
// caller-supplied ids). The API handler enforces that via RequireTaskRole
// before calling in.
func (r *Repository) AddTaskDependency(ctx context.Context, input domain.AddTaskDependencyInput) error {
	_, err := r.pool.ExecContext(ctx,
		"INSERT IGNORE INTO flow_task_dependencies (task_id, depends_on_task_id, business_id) VALUES (?, ?, ?)",
		input.TaskID, input.DependsOnTaskID, input.BusinessID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("AddTaskDependency failed", logger.Error(err))
	}
	return err
}

func (r *Repository) RemoveTaskDependency(ctx context.Context, input domain.RemoveTaskDependencyInput) error {
	_, err := r.pool.ExecContext(ctx,
		"DELETE FROM flow_task_dependencies WHERE task_id = ? AND depends_on_task_id = ? AND business_id = ?",
		input.TaskID, input.DependsOnTaskID, input.BusinessID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("RemoveTaskDependency failed", logger.Error(err))
	}
	return err
}

// ListTaskDependencies returns dependency ids keyed by task id. It takes a slice
// so list views resolve every row in one query instead of N.
func (r *Repository) ListTaskDependencies(ctx context.Context, taskIDs []int64, businessID int64) (map[int64][]int64, error) {
	out := make(map[int64][]int64, len(taskIDs))
	if len(taskIDs) == 0 {
		return out, nil
	}
	q, args, err := sq.
		Select("task_id", "depends_on_task_id").
		From("flow_task_dependencies").
		Where(sq.Eq{"task_id": taskIDs, "business_id": businessID}).
		OrderBy("id").
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.pool.QueryContext(ctx, q, args...)
	if err != nil {
		logger.Ctx(ctx).Error("ListTaskDependencies failed", logger.Error(err))
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var taskID, dependsOn int64
		if err := rows.Scan(&taskID, &dependsOn); err != nil {
			return nil, err
		}
		out[taskID] = append(out[taskID], dependsOn)
	}
	return out, rows.Err()
}

func (r *Repository) UpdateTaskStatus(ctx context.Context, input domain.UpdateTaskStatusInput) error {
	err := r.queries.UpdateTaskStatus(ctx, sqlc.UpdateTaskStatusParams{
		Status:     sqlc.NullFlowTasksStatus{FlowTasksStatus: sqlc.FlowTasksStatus(input.Status), Valid: true},
		Column2:    input.Status,
		Column3:    input.Status,
		ID:         input.ID,
		BusinessID: input.BusinessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("UpdateTaskStatus failed", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) DeleteTask(ctx context.Context, taskID, businessID int64) error {
	err := r.queries.DeleteTask(ctx, sqlc.DeleteTaskParams{
		ID:         taskID,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("DeleteTask failed", logger.Error(err))
		return err
	}
	return nil
}

// ─────────────────────────────────────────────
// Dashboard
// ─────────────────────────────────────────────

func (r *Repository) GetDashboardStats(ctx context.Context, workspaceID, businessID int64) (domain.DashboardStats, error) {
	var stats domain.DashboardStats

	weekStart, _ := times.GetWeekRangeStartAndEnd()
	weekEnd := weekStart.AddDate(0, 0, 7)

	row, err := r.queries.GetDashboardStats(ctx, sqlc.GetDashboardStatsParams{
		CompletedAt:   sql.NullTime{Time: weekStart, Valid: true},
		CompletedAt_2: sql.NullTime{Time: weekEnd, Valid: true},
		WorkspaceID:   workspaceID,
		BusinessID:    businessID,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Ctx(ctx).Error("GetDashboardStats failed", logger.Error(err))
		return domain.DashboardStats{}, err
	}

	stats.OpenTasks = row.OpenTasks
	stats.InProgressTasks = row.InProgressTasks
	stats.CompletedThisWeek = row.CompletedThisWeek
	stats.OverdueTasks = row.OverdueTasks

	return stats, nil
}

// ─────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────

func nullInt64Ptr(v *int64) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *v, Valid: true}
}

func nullTimePtrToSql(v *time.Time) sql.NullTime {
	if v == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *v, Valid: true}
}

func marshalLabels(labels []int64) (json.RawMessage, error) {
	if labels == nil {
		return json.RawMessage("[]"), nil
	}
	return json.Marshal(labels)
}

// enrichedTaskColumns is the column list used by every task read path.
// It selects the raw flow_tasks row plus the JOINed display projections
// (project name/color, assignee name, subtask count) so the UI never has
// to do an N+1 lookup.
//
// Keep this list, scanEnrichedTask, and the JOINs in
// enrichedTaskSelect in sync.
func enrichedTaskColumns(includeDescription bool) []string {
	descriptionColumn := "t.description"
	if !includeDescription {
		descriptionColumn = "'' AS description"
	}

	return []string{
		"t.id", "t.project_id", "t.workspace_id", "t.owner_id", "t.business_id", "t.parent_task_id", "t.section_id",
		"t.title", descriptionColumn, "COALESCE(t.status, 'todo') AS status", "COALESCE(t.priority, 'none') AS priority",
		"t.assignee_id", "t.reporter_id", "t.due_date", "t.start_date", "t.completed_at",
		"t.estimated_hours", "t.actual_hours", "t.sort_order", "t.labels",
		"COALESCE(t.archived, FALSE) AS archived",
		"t.created_at", "t.updated_at",
		"COALESCE(p.name, '')  AS project_name",
		"COALESCE(p.color, '') AS project_color",
		"COALESCE(u.full_name, '') AS assignee_name",
		"COALESCE((SELECT COUNT(*) FROM flow_tasks st WHERE st.parent_task_id = t.id), 0) AS subtask_count",
		"t.status_id",
		"COALESCE(fs.name, '') AS status_name",
		"COALESCE(fs.color, '') AS status_color",
	}
}

// enrichedTaskSelect returns a squirrel builder with the enriched task
// SELECT + JOINs preconfigured. Callers add WHERE/ORDER/LIMIT clauses.
// All tenant-scope filtering is the caller's responsibility.
//
// dynamic-query: squirrel
func enrichedTaskSelect() sq.SelectBuilder {
	return enrichedTaskSelectForList(false)
}

// dynamic-query: squirrel
func enrichedTaskSelectForList(compact bool) sq.SelectBuilder {
	return sq.
		Select(enrichedTaskColumns(!compact)...).
		From("flow_tasks t").
		LeftJoin("flow_projects p ON p.id = t.project_id").
		LeftJoin("users u ON u.id = t.assignee_id").
		LeftJoin("flow_statuses fs ON fs.id = t.status_id").
		PlaceholderFormat(sq.Question)
}

// listEnrichedTasks runs an enriched task query built by the caller.
// dynamic-query: squirrel
func (r *Repository) listEnrichedTasks(ctx context.Context, builder sq.SelectBuilder) ([]domain.Task, error) {
	q, args, err := builder.ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("listEnrichedTasks build failed", logger.Error(err))
		return nil, err
	}
	rows, err := r.pool.QueryContext(ctx, q, args...)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			// Request cancellation is expected when a client navigates away or
			// disconnects; it is not an application/database failure.
			logger.Ctx(ctx).Debug("listEnrichedTasks canceled", logger.Error(err))
		} else {
			logger.Ctx(ctx).Error("listEnrichedTasks failed", logger.Error(err))
		}
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			logger.Ctx(ctx).Warn("listEnrichedTasks row close failed", logger.Error(closeErr))
		}
	}()

	var tasks []domain.Task
	for rows.Next() {
		t, scanErr := scanEnrichedTask(ctx, rows)
		if scanErr != nil {
			return nil, scanErr
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// scanEnrichedTask scans one row from a query using enrichedTaskColumns.
// Returns a fully-populated domain.Task including display projections.
func scanEnrichedTask(ctx context.Context, rows *sql.Rows) (domain.Task, error) {
	var t domain.Task
	var desc sql.NullString
	var parentID, sectionID, assigneeID, statusID sql.NullInt64
	var dueDate, startDate, completedAt sql.NullTime
	var estHours, actHours sql.NullFloat64
	var labelsRaw []byte

	err := rows.Scan(
		&t.ID, &t.ProjectID, &t.WorkspaceID, &t.OwnerID, &t.BusinessID, &parentID, &sectionID,
		&t.Title, &desc, &t.Status, &t.Priority,
		&assigneeID, &t.ReporterID,
		&dueDate, &startDate, &completedAt,
		&estHours, &actHours, &t.SortOrder, &labelsRaw, &t.Archived,
		&t.CreatedAt, &t.UpdatedAt,
		&t.ProjectName, &t.ProjectColor, &t.AssigneeName, &t.SubtaskCount,
		&statusID, &t.StatusName, &t.StatusColor,
	)
	if err != nil {
		logger.Ctx(ctx).Error("scanEnrichedTask row scan failed", logger.Error(err))
		return domain.Task{}, err
	}

	t.Description = desc.String
	if parentID.Valid {
		t.ParentTaskID = &parentID.Int64
	}
	if sectionID.Valid {
		t.SectionID = &sectionID.Int64
	}
	if assigneeID.Valid {
		t.AssigneeID = &assigneeID.Int64
	}
	if statusID.Valid {
		t.StatusID = &statusID.Int64
	}
	if dueDate.Valid {
		t.DueDate = &dueDate.Time
	}
	if startDate.Valid {
		t.StartDate = &startDate.Time
	}
	if completedAt.Valid {
		t.CompletedAt = &completedAt.Time
	}
	if estHours.Valid {
		t.EstimatedHours = &estHours.Float64
	}
	if actHours.Valid {
		t.ActualHours = &actHours.Float64
	}
	if len(labelsRaw) > 0 {
		if unErr := json.Unmarshal(labelsRaw, &t.Labels); unErr != nil {
			logger.Ctx(ctx).Warn("failed to unmarshal task labels", logger.Error(unErr))
		}
	}
	t.AssigneeInitials = initialsOf(t.AssigneeName)
	if t.Identifier == "" {
		t.Identifier = "T-" + strconv.FormatInt(t.ID, 10)
	}
	return t, nil
}

// initialsOf returns uppercase initials for a display name (up to 2 chars).
func initialsOf(name string) string {
	if name == "" {
		return ""
	}
	out := make([]rune, 0, 2)
	prev := ' '
	for _, r := range name {
		if (prev == ' ' || prev == '\t') && r != ' ' && r != '\t' {
			if r >= 'a' && r <= 'z' {
				r -= 32
			}
			out = append(out, r)
			if len(out) == 2 {
				break
			}
		}
		prev = r
	}
	return string(out)
}

// scanTasksRaw — legacy bare-row scanner kept for code paths not yet
// migrated to listEnrichedTasks (e.g. direct-write callbacks returning
// the inserted row). New callers should use listEnrichedTasks + scanEnrichedTask.
func scanTasksRaw(ctx context.Context, rows *sql.Rows) ([]domain.Task, error) {
	var tasks []domain.Task
	for rows.Next() {
		var t domain.Task
		var desc sql.NullString
		var parentID, sectionID, assigneeID sql.NullInt64
		var dueDate, startDate, completedAt sql.NullTime
		var estHours, actHours sql.NullFloat64
		var labelsRaw []byte

		err := rows.Scan(
			&t.ID, &t.ProjectID, &t.WorkspaceID, &t.OwnerID, &parentID, &sectionID,
			&t.Title, &desc, &t.Status, &t.Priority,
			&assigneeID, &t.ReporterID,
			&dueDate, &startDate, &completedAt,
			&estHours, &actHours, &t.SortOrder, &labelsRaw, &t.Archived,
			&t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			logger.Ctx(ctx).Error("scanTasksRaw row scan failed", logger.Error(err))
			return nil, err
		}

		t.Description = desc.String
		if parentID.Valid {
			t.ParentTaskID = &parentID.Int64
		}
		if sectionID.Valid {
			t.SectionID = &sectionID.Int64
		}
		if assigneeID.Valid {
			t.AssigneeID = &assigneeID.Int64
		}
		if dueDate.Valid {
			t.DueDate = &dueDate.Time
		}
		if startDate.Valid {
			t.StartDate = &startDate.Time
		}
		if completedAt.Valid {
			t.CompletedAt = &completedAt.Time
		}
		if estHours.Valid {
			t.EstimatedHours = &estHours.Float64
		}
		if actHours.Valid {
			t.ActualHours = &actHours.Float64
		}
		if len(labelsRaw) > 0 {
			if unErr := json.Unmarshal(labelsRaw, &t.Labels); unErr != nil {
				logger.Ctx(ctx).Warn("failed to unmarshal task labels", logger.Error(unErr))
			}
		}
		t.Identifier = "T-" + strconv.FormatInt(t.ID, 10)
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// ─────────────────────────────────────────────
// Workspace Members (extended)
// ─────────────────────────────────────────────

func (r *Repository) ListWorkspaceMembers(ctx context.Context, workspaceID, limit, offset int64) ([]domain.WorkspaceMember, int64, error) {
	total, err := r.queries.CountWorkspaceMembers(ctx, workspaceID)
	if err != nil {
		logger.Ctx(ctx).Error("CountWorkspaceMembers failed", logger.Error(err))
		return nil, 0, err
	}

	// cross-schema read from users — workspace member display names
	rows, err := r.queries.ListWorkspaceMembers(ctx, sqlc.ListWorkspaceMembersParams{
		WorkspaceID: workspaceID,
		Limit:       int32(limit),  // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Offset:      int32(offset), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		logger.Ctx(ctx).Error("ListWorkspaceMembers failed", logger.Error(err))
		return nil, 0, err
	}

	members := make([]domain.WorkspaceMember, 0, len(rows))
	for _, row := range rows {
		members = append(members, domain.WorkspaceMember{
			ID:          row.ID,
			WorkspaceID: row.WorkspaceID,
			UserID:      row.UserID,
			Role:        string(row.Role.FlowWorkspaceMembersRole),
			JoinedAt:    row.JoinedAt.Time,
			Name:        row.Name,
			Email:       row.Email,
		})
	}
	return members, total, nil
}

func (r *Repository) UpdateWorkspaceMemberRole(ctx context.Context, input domain.UpdateWorkspaceMemberRoleInput) error {
	err := r.queries.UpdateWorkspaceMemberRole(ctx, sqlc.UpdateWorkspaceMemberRoleParams{
		Role:        sqlc.NullFlowWorkspaceMembersRole{FlowWorkspaceMembersRole: sqlc.FlowWorkspaceMembersRole(input.Role), Valid: true},
		WorkspaceID: input.WorkspaceID,
		UserID:      input.UserID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("UpdateWorkspaceMemberRole failed", logger.Error(err))
	}
	return err
}

func (r *Repository) CreatePersonalWorkspace(ctx context.Context, userID, businessID int64) (int64, error) {
	existing, err := r.queries.GetPersonalWorkspaceByOwner(ctx, sqlc.GetPersonalWorkspaceByOwnerParams{
		OwnerID:    userID,
		BusinessID: businessID,
	})
	if err == nil && existing > 0 {
		return existing, nil
	}
	// Only a genuine miss may fall through to the INSERT. flow_workspaces has
	// no unique key on owner_id/business_id/is_personal, so treating a
	// transient lookup failure (pool exhaustion, statement timeout) as "no
	// workspace yet" inserts a second 'Personal' workspace for the same owner.
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Ctx(ctx).Error("CreatePersonalWorkspace lookup failed", logger.Error(err))
		return 0, err
	}

	res, err := r.queries.CreatePersonalWorkspace(ctx, sqlc.CreatePersonalWorkspaceParams{
		OwnerID:    userID,
		BusinessID: businessID,
		CreatedBy:  userID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("CreatePersonalWorkspace failed", logger.Error(err))
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		logger.Ctx(ctx).Error("CreatePersonalWorkspace last insert id failed", logger.Error(err))
		return 0, err
	}

	if addErr := r.queries.AddWorkspaceMember(ctx, sqlc.AddWorkspaceMemberParams{
		WorkspaceID: id,
		UserID:      userID,
		Role:        sqlc.NullFlowWorkspaceMembersRole{FlowWorkspaceMembersRole: sqlc.FlowWorkspaceMembersRoleAdmin, Valid: true},
	}); addErr != nil {
		// A workspace whose owner is not a member is not usable; reporting the
		// insert as a success left the caller with a half-created workspace.
		logger.Ctx(ctx).Error("CreatePersonalWorkspace add member failed", logger.Error(addErr))
		return 0, addErr
	}
	return id, nil
}

// ─────────────────────────────────────────────
// Project Members
// ─────────────────────────────────────────────

func (r *Repository) ListProjectMembers(ctx context.Context, projectID, limit, offset int64) ([]domain.ProjectMember, int64, error) {
	total, err := r.queries.CountProjectMembers(ctx, projectID)
	if err != nil {
		logger.Ctx(ctx).Error("CountProjectMembers failed", logger.Error(err))
		return nil, 0, err
	}

	rows, err := r.queries.ListProjectMembers(ctx, sqlc.ListProjectMembersParams{
		ProjectID: projectID,
		Limit:     int32(limit),  // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Offset:    int32(offset), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		logger.Ctx(ctx).Error("ListProjectMembers failed", logger.Error(err))
		return nil, 0, err
	}

	members := make([]domain.ProjectMember, 0, len(rows))
	for _, row := range rows {
		members = append(members, domain.ProjectMember{
			ID:        row.ID,
			ProjectID: row.ProjectID,
			UserID:    row.UserID,
			Role:      string(row.Role.FlowProjectMembersRole),
			AddedAt:   row.JoinedAt.Time,
			Name:      row.Name,
			Email:     row.Email,
		})
	}
	return members, total, nil
}

func (r *Repository) AddProjectMember(ctx context.Context, input domain.AddProjectMemberInput) error {
	role := input.Role
	if role == "" {
		role = "member"
	}
	err := r.queries.AddProjectMember(ctx, sqlc.AddProjectMemberParams{
		ProjectID: input.ProjectID,
		UserID:    input.UserID,
		Role:      sqlc.NullFlowProjectMembersRole{FlowProjectMembersRole: sqlc.FlowProjectMembersRole(role), Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("AddProjectMember failed", logger.Error(err))
	}
	return err
}

func (r *Repository) RemoveProjectMember(ctx context.Context, input domain.RemoveProjectMemberInput) error {
	err := r.queries.RemoveProjectMember(ctx, sqlc.RemoveProjectMemberParams{
		ProjectID: input.ProjectID,
		UserID:    input.UserID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("RemoveProjectMember failed", logger.Error(err))
	}
	return err
}

func (r *Repository) UpdateProjectMemberRole(ctx context.Context, input domain.UpdateProjectMemberRoleInput) error {
	err := r.queries.UpdateProjectMemberRole(ctx, sqlc.UpdateProjectMemberRoleParams{
		Role:      sqlc.NullFlowProjectMembersRole{FlowProjectMembersRole: sqlc.FlowProjectMembersRole(input.Role), Valid: true},
		ProjectID: input.ProjectID,
		UserID:    input.UserID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("UpdateProjectMemberRole failed", logger.Error(err))
	}
	return err
}

func (r *Repository) TransferProjectOwnership(ctx context.Context, input domain.TransferProjectOwnershipInput) error {
	err := r.queries.TransferProjectOwnership(ctx, sqlc.TransferProjectOwnershipParams{
		OwnerID:   input.NewOwnerID,
		ID:        input.ProjectID,
		OwnerID_2: input.CallerID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("TransferProjectOwnership failed", logger.Error(err))
	}
	return err
}

func (r *Repository) MoveProject(ctx context.Context, input domain.MoveProjectInput) error {
	err := r.queries.MoveProject(ctx, sqlc.MoveProjectParams{
		WorkspaceID: input.WorkspaceID,
		ID:          input.ProjectID,
		OwnerID:     input.CallerID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("MoveProject failed", logger.Error(err))
	}
	return err
}

func (r *Repository) GetUserWorkspaceRole(ctx context.Context, workspaceID, userID int64) (string, error) {
	role, err := r.queries.GetUserWorkspaceRole(ctx, sqlc.GetUserWorkspaceRoleParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
	})
	if err != nil {
		return "", err
	}
	if !role.Valid {
		return "", sql.ErrNoRows
	}
	return string(role.FlowWorkspaceMembersRole), nil
}

// GetProjectMemberRole returns the caller's role on a project, or
// sql.ErrNoRows when they hold no explicit membership row. A project owner and
// a workspace member both reach a project without one, so a not-found here is
// not the same as "no access" — see Service.RequireProjectRole.
func (r *Repository) GetProjectMemberRole(ctx context.Context, projectID, userID int64) (string, error) {
	role, err := r.queries.GetProjectMemberRole(ctx, sqlc.GetProjectMemberRoleParams{
		ProjectID: projectID,
		UserID:    userID,
	})
	if err != nil {
		return "", err
	}
	if !role.Valid {
		return "", sql.ErrNoRows
	}
	return string(role.FlowProjectMembersRole), nil
}

// GetProjectOwnerWorkspace returns a project's owner and workspace, the two
// facts needed to fall back from an absent project membership row.
func (r *Repository) GetProjectOwnerWorkspace(ctx context.Context, projectID int64) (int64, int64, error) {
	row, err := r.queries.GetProjectOwnerWorkspace(ctx, projectID)
	if err != nil {
		return 0, 0, err
	}
	return row.OwnerID, row.WorkspaceID, nil
}

// GetProjectProxyFacts returns the project facts the dept-manager visibility
// proxy needs, by id alone. projectRoleFor has no business in hand, and the
// proxy must resolve the project's real business rather than trust one supplied
// by the caller.
func (r *Repository) GetProjectProxyFacts(ctx context.Context, projectID int64) (domain.Project, error) {
	row, err := r.queries.GetProjectProxyFacts(ctx, projectID)
	if err != nil {
		return domain.Project{}, err
	}
	return domain.Project{
		ID:         projectID,
		BusinessID: row.BusinessID,
		OwnerID:    row.OwnerID,
		CreatedBy:  row.CreatedBy,
		IsPersonal: row.IsPersonal,
	}, nil
}

// GetWorkspaceOwner returns a workspace's owner. A workspace owner has no
// flow_workspace_members row of their own, so the roster read alone would miss
// them.
func (r *Repository) GetWorkspaceOwner(ctx context.Context, workspaceID int64) (int64, error) {
	return r.queries.GetWorkspaceOwner(ctx, workspaceID)
}

// ─── Project Columns ─────────────────────────────────────────────────────────

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

func (r *Repository) MoveTaskToColumn(ctx context.Context, taskID, columnID, actorID int64) error {
	var sectionID sql.NullInt64
	if columnID != 0 {
		sectionID = sql.NullInt64{Int64: columnID, Valid: true}
	}
	res, err := r.queries.MoveTaskToColumn(ctx, sqlc.MoveTaskToColumnParams{
		SectionID: sectionID,
		ID:        taskID,
		// One actor across every branch: project member, project owner, project
		// creator, then the two legacy workspace branches.
		UserID:    actorID,
		OwnerID:   actorID,
		CreatedBy: actorID,
		OwnerID_2: actorID,
		UserID_2:  actorID,
		ColumnID:  columnID,
	})
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
