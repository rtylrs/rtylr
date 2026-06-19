package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/workspaces/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
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

func (r *Repository) GetWorkspacesByOwner(ctx context.Context, businessID int64) ([]domain.Workspace, error) {
	rows, err := r.queries.GetWorkspacesByOwner(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("GetWorkspacesByOwner failed", logger.Error(err))
		return nil, err
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
	return workspaces, nil
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

	res, err := r.queries.CreateProject(ctx, sqlc.CreateProjectParams{
		WorkspaceID: input.WorkspaceID,
		OwnerID:     input.OwnerID,
		BusinessID:  input.BusinessID,
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

func (r *Repository) GetProjectsByWorkspace(ctx context.Context, workspaceID, businessID int64) ([]domain.Project, error) {
	// Enriched projection: task_count, completed_count, and member_count
	// are computed via correlated subqueries so the list view doesn't
	// need separate fetches per project.
	q, args, err := sq.
		Select(
			"p.id", "p.workspace_id", "p.owner_id", "p.business_id", "p.name",
			"COALESCE(p.description, '')", "COALESCE(p.icon, '')", "COALESCE(p.color, '')",
			"COALESCE(p.status, 'active')",
			"p.start_date", "p.due_date",
			"p.created_by", "p.created_at", "p.updated_at",
			"COALESCE((SELECT COUNT(*) FROM flow_tasks t WHERE t.project_id = p.id), 0) AS task_count",
			"COALESCE((SELECT COUNT(*) FROM flow_tasks t WHERE t.project_id = p.id AND t.status = 'done'), 0) AS completed_count",
			"COALESCE((SELECT COUNT(*) FROM flow_project_members pm WHERE pm.project_id = p.id), 0) AS member_count",
		).
		From("flow_projects p").
		Where(sq.Eq{"p.workspace_id": workspaceID, "p.business_id": businessID}).
		OrderBy("p.created_at DESC").
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("GetProjectsByWorkspace build failed", logger.Error(err))
		return nil, err
	}

	rows, err := r.pool.QueryContext(ctx, q, args...)
	if err != nil {
		logger.Ctx(ctx).Error("GetProjectsByWorkspace failed", logger.Error(err))
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			logger.Ctx(ctx).Warn("GetProjectsByWorkspace close failed", logger.Error(closeErr))
		}
	}()

	var projects []domain.Project
	for rows.Next() {
		var p domain.Project
		var startDate, dueDate sql.NullTime
		var createdAt, updatedAt sql.NullTime
		if scanErr := rows.Scan(
			&p.ID, &p.WorkspaceID, &p.OwnerID, &p.BusinessID, &p.Name,
			&p.Description, &p.Icon, &p.Color, &p.Status,
			&startDate, &dueDate,
			&p.CreatedBy, &createdAt, &updatedAt,
			&p.TaskCount, &p.CompletedCount, &p.MemberCount,
		); scanErr != nil {
			logger.Ctx(ctx).Error("GetProjectsByWorkspace scan failed", logger.Error(scanErr))
			return nil, scanErr
		}
		if startDate.Valid {
			p.StartDate = &startDate.Time
		}
		if dueDate.Valid {
			p.DueDate = &dueDate.Time
		}
		if createdAt.Valid {
			p.CreatedAt = createdAt.Time
		}
		if updatedAt.Valid {
			p.UpdatedAt = updatedAt.Time
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
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

func (r *Repository) ArchiveProject(ctx context.Context, id, businessID int64) error {
	err := r.queries.ArchiveProject(ctx, sqlc.ArchiveProjectParams{
		ID:         id,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("ArchiveProject failed", logger.Error(err))
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

func (r *Repository) GetTasksByProject(ctx context.Context, filter domain.ListTasksFilter) ([]domain.Task, error) {
	builder := enrichedTaskSelectForList(filter.Compact).
		Where(sq.Eq{"t.project_id": filter.ProjectID}).
		Where("t.parent_task_id IS NULL")

	if filter.Status != "" {
		builder = builder.Where(sq.Eq{"t.status": filter.Status})
	} else if filter.ActiveOnly {
		builder = builder.Where("t.status NOT IN ('done', 'cancelled')")
	}
	if filter.AssigneeID != nil {
		builder = builder.Where(sq.Eq{"t.assignee_id": *filter.AssigneeID})
	}
	if filter.Priority != "" {
		builder = builder.Where(sq.Eq{"t.priority": filter.Priority})
	}

	builder = builder.OrderBy("t.sort_order ASC", "t.created_at ASC")

	return r.listEnrichedTasks(ctx, builder)
}

func (r *Repository) GetTasksByWorkspace(ctx context.Context, filter domain.ListTasksFilter) ([]domain.Task, error) {
	builder := enrichedTaskSelectForList(filter.Compact).
		Where(sq.Eq{"t.workspace_id": filter.WorkspaceID}).
		Where("t.parent_task_id IS NULL")

	if filter.Status != "" {
		builder = builder.Where(sq.Eq{"t.status": filter.Status})
	} else if filter.ActiveOnly {
		builder = builder.Where("t.status NOT IN ('done', 'cancelled')")
	}
	if filter.AssigneeID != nil {
		builder = builder.Where(sq.Eq{"t.assignee_id": *filter.AssigneeID})
	}
	if filter.Priority != "" {
		builder = builder.Where(sq.Eq{"t.priority": filter.Priority})
	}

	builder = builder.OrderBy("t.created_at DESC").Limit(500)

	return r.listEnrichedTasks(ctx, builder)
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

func (r *Repository) GetSubtasks(ctx context.Context, parentTaskID int64) ([]domain.Task, error) {
	return r.listEnrichedTasks(ctx,
		enrichedTaskSelect().
			Where(sq.Eq{"t.parent_task_id": parentTaskID}).
			OrderBy("t.sort_order ASC", "t.created_at ASC"),
	)
}

// myTasksDoneWindowDays bounds how far back completed tasks surface in the global
// "My tasks" board, so the Done column shows recent work without unbounded history.
const myTasksDoneWindowDays = 14

func (r *Repository) GetMyTasks(ctx context.Context, userID int64) ([]domain.Task, error) {
	// All active work assigned to the user, plus tasks they completed within the
	// recent window so the board's Done column isn't always empty. Cancelled tasks
	// stay hidden. completed_at is stamped by both the status and partial-update
	// paths when a task moves to 'done'.
	doneCutoff := time.Now().AddDate(0, 0, -myTasksDoneWindowDays)
	return r.listEnrichedTasks(ctx,
		enrichedTaskSelect().
			Where(sq.Eq{"t.assignee_id": userID}).
			Where("t.parent_task_id IS NULL").
			Where(sq.Or{
				sq.Expr("t.status NOT IN ('done','cancelled')"),
				sq.And{
					sq.Eq{"t.status": "done"},
					sq.GtOrEq{"t.completed_at": doneCutoff},
				},
			}).
			OrderBy(
				"CASE WHEN t.due_date IS NULL THEN 1 ELSE 0 END",
				"t.due_date ASC",
				"t.created_at DESC",
			).
			Limit(500),
	)
}

func (r *Repository) UpdateTask(ctx context.Context, input domain.UpdateTaskInput) error {
	labelsJSON, err := marshalLabels(input.Labels)
	if err != nil {
		return err
	}

	var statusParam interface{}
	if input.Status != "" {
		statusParam = input.Status
	}
	var priorityParam interface{}
	if input.Priority != "" {
		priorityParam = input.Priority
	}
	var labelsParam interface{}
	if labelsJSON != nil {
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

	// Partial updates (the client sends only the fields it changed, e.g. just
	// priority or assignee) must not overwrite the fields it omitted. COALESCE(?, col)
	// preserves the existing value when the param is NULL, and every nullable param
	// above is left NULL when the request omits that field. The completed_at CASE only
	// fires when a non-NULL status is explicitly supplied.
	q, args, err := sq.
		Update("flow_tasks").
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
// Comments
// ─────────────────────────────────────────────

func (r *Repository) CreateComment(ctx context.Context, taskID, userID int64, content string) (int64, error) {
	res, err := r.queries.CreateComment(ctx, sqlc.CreateCommentParams{
		TaskID:  taskID,
		UserID:  userID,
		Content: content,
	})
	if err != nil {
		logger.Ctx(ctx).Error("CreateComment failed", logger.Error(err))
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) GetCommentsByTask(ctx context.Context, taskID int64) ([]domain.TaskComment, error) {
	rows, err := r.queries.GetCommentsByTask(ctx, taskID)
	if err != nil {
		logger.Ctx(ctx).Error("GetCommentsByTask failed", logger.Error(err))
		return nil, err
	}

	comments := make([]domain.TaskComment, 0, len(rows))
	for _, row := range rows {
		comments = append(comments, domain.TaskComment{
			ID:        row.ID,
			TaskID:    row.TaskID,
			UserID:    row.UserID,
			Content:   row.Content,
			IsEdited:  row.IsEdited.Bool,
			CreatedAt: row.CreatedAt.Time,
			UpdatedAt: row.UpdatedAt.Time,
		})
	}
	return comments, nil
}

func (r *Repository) DeleteComment(ctx context.Context, commentID, userID int64) error {
	err := r.queries.DeleteComment(ctx, sqlc.DeleteCommentParams{
		ID:     commentID,
		UserID: userID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("DeleteComment failed", logger.Error(err))
		return err
	}
	return nil
}

// GetCommentTaskID returns the task a comment belongs to. Used by the
// comment handlers to route realtime events to the correct task topic
// before the row disappears.
func (r *Repository) GetCommentTaskID(ctx context.Context, commentID int64) (int64, error) {
	q, args, err := sq.
		Select("task_id").
		From("flow_comments").
		Where(sq.Eq{"id": commentID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return 0, err
	}
	var taskID int64
	if err := r.pool.QueryRowContext(ctx, q, args...).Scan(&taskID); err != nil {
		return 0, err
	}
	return taskID, nil
}

// ─────────────────────────────────────────────
// Activity
// ─────────────────────────────────────────────

func (r *Repository) LogActivity(ctx context.Context, taskID, userID int64, action, oldValue, newValue string) error {
	err := r.queries.LogActivity(ctx, sqlc.LogActivityParams{
		TaskID:   taskID,
		UserID:   userID,
		Action:   action,
		OldValue: sql.NullString{String: oldValue, Valid: oldValue != ""},
		NewValue: sql.NullString{String: newValue, Valid: newValue != ""},
	})
	if err != nil {
		logger.Ctx(ctx).Error("LogActivity failed", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) GetActivityByTask(ctx context.Context, taskID int64, limit int) ([]domain.TaskActivity, error) {
	rows, err := r.queries.GetActivityByTask(ctx, sqlc.GetActivityByTaskParams{
		TaskID: taskID,
		Limit:  int32(limit),
	})
	if err != nil {
		logger.Ctx(ctx).Error("GetActivityByTask failed", logger.Error(err))
		return nil, err
	}

	activities := make([]domain.TaskActivity, 0, len(rows))
	for _, row := range rows {
		activities = append(activities, domain.TaskActivity{
			ID:        row.ID,
			TaskID:    row.TaskID,
			UserID:    row.UserID,
			Action:    row.Action,
			OldValue:  row.OldValue.String,
			NewValue:  row.NewValue.String,
			CreatedAt: row.CreatedAt.Time,
		})
	}
	return activities, nil
}

// ─────────────────────────────────────────────
// Attachments
// ─────────────────────────────────────────────

func (r *Repository) CreateAttachment(ctx context.Context, input domain.TaskAttachmentInput) (int64, error) {
	res, err := r.queries.CreateAttachment(ctx, sqlc.CreateAttachmentParams{
		TaskID:      input.TaskID,
		OwnerID:     input.OwnerID,
		BusinessID:  input.BusinessID,
		Name:        input.Name,
		Url:        input.URL,
		Size:       sql.NullInt64{Int64: input.Size, Valid: true},
		UploadedBy: input.UploadedBy,
	})
	if err != nil {
		logger.Ctx(ctx).Error("CreateAttachment failed", logger.Error(err))
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) GetAttachmentsByTask(ctx context.Context, taskID int64) ([]domain.TaskAttachment, error) {
	rows, err := r.queries.GetAttachmentsByTask(ctx, taskID)
	if err != nil {
		logger.Ctx(ctx).Error("GetAttachmentsByTask failed", logger.Error(err))
		return nil, err
	}

	attachments := make([]domain.TaskAttachment, 0, len(rows))
	for _, row := range rows {
		attachments = append(attachments, domain.TaskAttachment{
			ID:         row.ID,
			TaskID:     row.TaskID,
			OwnerID:    row.OwnerID,
			BusinessID: row.BusinessID,
			Name:       row.Name,
			URL:        row.Url,
			Size:       row.Size.Int64,
			UploadedBy: row.UploadedBy,
			CreatedAt:  row.CreatedAt.Time,
		})
	}
	return attachments, nil
}

func (r *Repository) DeleteAttachment(ctx context.Context, attachmentID, businessID int64) error {
	err := r.queries.DeleteAttachment(ctx, sqlc.DeleteAttachmentParams{
		ID:         attachmentID,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("DeleteAttachment failed", logger.Error(err))
		return err
	}
	return nil
}

// ─────────────────────────────────────────────
// Time Logs
// ─────────────────────────────────────────────

func (r *Repository) CreateTimeLog(ctx context.Context, input domain.TimeLogInput) (int64, error) {
	res, err := r.queries.CreateTimeLog(ctx, sqlc.CreateTimeLogParams{
		OwnerID:         input.OwnerID,
		BusinessID:      input.BusinessID,
		TaskID:          input.TaskID,
		UserID:          input.UserID,
		DurationMinutes: int32(input.DurationMinutes),
		Description:     sql.NullString{String: input.Description, Valid: input.Description != ""},
	})
	if err != nil {
		logger.Ctx(ctx).Error("CreateTimeLog failed", logger.Error(err))
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) GetTimeLogsByTask(ctx context.Context, taskID int64) ([]domain.TimeLog, error) {
	rows, err := r.queries.GetTimeLogsByTask(ctx, taskID)
	if err != nil {
		logger.Ctx(ctx).Error("GetTimeLogsByTask failed", logger.Error(err))
		return nil, err
	}

	logs := make([]domain.TimeLog, 0, len(rows))
	for _, row := range rows {
		logs = append(logs, domain.TimeLog{
			ID:              row.ID,
			OwnerID:         row.OwnerID,
			BusinessID:      row.BusinessID,
			TaskID:          row.TaskID,
			UserID:          row.UserID,
			DurationMinutes: int(row.DurationMinutes),
			Description:     row.Description.String,
			LoggedAt:        row.LoggedAt.Time,
			CreatedAt:       row.CreatedAt.Time,
		})
	}
	return logs, nil
}

// ─────────────────────────────────────────────
// Labels
// ─────────────────────────────────────────────

func (r *Repository) CreateLabel(ctx context.Context, input domain.LabelInput) (int64, error) {
	color := input.Color
	if color == "" {
		color = "#6366f1"
	}

	res, err := r.queries.CreateLabel(ctx, sqlc.CreateLabelParams{
		WorkspaceID: input.WorkspaceID,
		Name:        input.Name,
		Color:       color,
	})
	if err != nil {
		logger.Ctx(ctx).Error("CreateLabel failed", logger.Error(err))
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) GetLabelsByWorkspace(ctx context.Context, workspaceID int64) ([]domain.Label, error) {
	rows, err := r.queries.GetLabelsByWorkspace(ctx, workspaceID)
	if err != nil {
		logger.Ctx(ctx).Error("GetLabelsByWorkspace failed", logger.Error(err))
		return nil, err
	}

	labels := make([]domain.Label, 0, len(rows))
	for _, row := range rows {
		labels = append(labels, domain.Label{
			ID:          row.ID,
			WorkspaceID: row.WorkspaceID,
			Name:        row.Name,
			Color:       row.Color,
			CreatedAt:   row.CreatedAt.Time,
		})
	}
	return labels, nil
}

func (r *Repository) UpdateLabel(ctx context.Context, id, actorID int64, input domain.LabelInput) error {
	q, args, err := sq.
		Update("flow_labels").
		Set("name", input.Name).
		Set("color", input.Color).
		Where(sq.Eq{"id": id}).
		Where(workspaceMemberPredicate("flow_labels"), actorID, actorID).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("UpdateLabel build failed", logger.Error(err))
		return err
	}
	res, err := r.pool.ExecContext(ctx, q, args...)
	if err != nil {
		logger.Ctx(ctx).Error("UpdateLabel failed", logger.Error(err))
		return err
	}
	return requireAffected(res)
}

func (r *Repository) DeleteLabel(ctx context.Context, id, actorID int64) error {
	q, args, err := sq.
		Delete("flow_labels").
		Where(sq.Eq{"id": id}).
		Where(workspaceMemberPredicate("flow_labels"), actorID, actorID).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("DeleteLabel build failed", logger.Error(err))
		return err
	}
	res, err := r.pool.ExecContext(ctx, q, args...)
	if err != nil {
		logger.Ctx(ctx).Error("DeleteLabel failed", logger.Error(err))
		return err
	}
	return requireAffected(res)
}

func (r *Repository) SetTaskLabels(ctx context.Context, taskID, actorID int64, labelIDs []int64) error {
	b, err := json.Marshal(labelIDs)
	if err != nil {
		return err
	}
	q, args, err := sq.
		Update("flow_tasks").
		Set("labels", b).
		Where(sq.Eq{"id": taskID}).
		Where(workspaceMemberPredicate("flow_tasks"), actorID, actorID).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("SetTaskLabels build failed", logger.Error(err))
		return err
	}
	res, err := r.pool.ExecContext(ctx, q, args...)
	if err != nil {
		logger.Ctx(ctx).Error("SetTaskLabels failed", logger.Error(err))
		return err
	}
	return requireAffected(res)
}

// ─────────────────────────────────────────────
// Time Logs (extended)
// ─────────────────────────────────────────────

func (r *Repository) GetTimeLogTaskID(ctx context.Context, id int64) (int64, error) {
	q, args, err := sq.
		Select("task_id").
		From("flow_time_logs").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return 0, err
	}
	var taskID int64
	err = r.pool.QueryRowContext(ctx, q, args...).Scan(&taskID)
	return taskID, err
}

func (r *Repository) DeleteTimeLog(ctx context.Context, id, userID int64) error {
	q, args, err := sq.
		Delete("flow_time_logs").
		Where(sq.Eq{"id": id, "user_id": userID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("DeleteTimeLog build failed", logger.Error(err))
		return err
	}
	if _, err = r.pool.ExecContext(ctx, q, args...); err != nil {
		logger.Ctx(ctx).Error("DeleteTimeLog failed", logger.Error(err))
	}
	return err
}

func (r *Repository) RecalcActualHours(ctx context.Context, taskID int64) error {
	q, args, err := sq.
		Update("flow_tasks").
		Set("actual_hours", sq.Expr(
			`(SELECT COALESCE(SUM(duration_minutes), 0) / 60.0 FROM flow_time_logs WHERE task_id = ?)`,
			taskID,
		)).
		Where(sq.Eq{"id": taskID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("RecalcActualHours build failed", logger.Error(err))
		return err
	}
	if _, err = r.pool.ExecContext(ctx, q, args...); err != nil {
		logger.Ctx(ctx).Error("RecalcActualHours failed", logger.Error(err))
	}
	return err
}

// ─────────────────────────────────────────────
// Dashboard
// ─────────────────────────────────────────────

func (r *Repository) GetDashboardStats(ctx context.Context, workspaceID, businessID int64) (domain.DashboardStats, error) {
	var stats domain.DashboardStats

	weekStart := time.Now().AddDate(0, 0, -int(time.Now().Weekday()))
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
		"t.id", "t.project_id", "t.workspace_id", "t.owner_id", "t.parent_task_id", "t.section_id",
		"t.title", descriptionColumn, "COALESCE(t.status, 'todo') AS status", "COALESCE(t.priority, 'none') AS priority",
		"t.assignee_id", "t.reporter_id", "t.due_date", "t.start_date", "t.completed_at",
		"t.estimated_hours", "t.actual_hours", "t.sort_order", "t.labels",
		"t.created_at", "t.updated_at",
		"COALESCE(p.name, '')  AS project_name",
		"COALESCE(p.color, '') AS project_color",
		"COALESCE(u.full_name, '') AS assignee_name",
		"COALESCE((SELECT COUNT(*) FROM flow_tasks st WHERE st.parent_task_id = t.id), 0) AS subtask_count",
	}
}

// enrichedTaskSelect returns a squirrel builder with the enriched task
// SELECT + JOINs preconfigured. Callers add WHERE/ORDER/LIMIT clauses.
// All tenant-scope filtering is the caller's responsibility.
func enrichedTaskSelect() sq.SelectBuilder {
	return enrichedTaskSelectForList(false)
}

func enrichedTaskSelectForList(compact bool) sq.SelectBuilder {
	return sq.
		Select(enrichedTaskColumns(!compact)...).
		From("flow_tasks t").
		LeftJoin("flow_projects p ON p.id = t.project_id").
		LeftJoin("users u ON u.id = t.assignee_id").
		PlaceholderFormat(sq.Question)
}

// listEnrichedTasks runs an enriched task query built by the caller.
func (r *Repository) listEnrichedTasks(ctx context.Context, builder sq.SelectBuilder) ([]domain.Task, error) {
	q, args, err := builder.ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("listEnrichedTasks build failed", logger.Error(err))
		return nil, err
	}
	rows, err := r.pool.QueryContext(ctx, q, args...)
	if err != nil {
		logger.Ctx(ctx).Error("listEnrichedTasks failed", logger.Error(err))
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
	var parentID, sectionID, assigneeID sql.NullInt64
	var dueDate, startDate, completedAt sql.NullTime
	var estHours, actHours sql.NullFloat64
	var labelsRaw []byte

	err := rows.Scan(
		&t.ID, &t.ProjectID, &t.WorkspaceID, &t.OwnerID, &parentID, &sectionID,
		&t.Title, &desc, &t.Status, &t.Priority,
		&assigneeID, &t.ReporterID,
		&dueDate, &startDate, &completedAt,
		&estHours, &actHours, &t.SortOrder, &labelsRaw,
		&t.CreatedAt, &t.UpdatedAt,
		&t.ProjectName, &t.ProjectColor, &t.AssigneeName, &t.SubtaskCount,
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
			&estHours, &actHours, &t.SortOrder, &labelsRaw,
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

func (r *Repository) ListWorkspaceMembers(ctx context.Context, workspaceID int64) ([]domain.WorkspaceMember, error) {
	q, args, err := sq.
		Select(
			"wm.id", "wm.workspace_id", "wm.user_id", "wm.role", "wm.invited_by", "wm.joined_at",
			"COALESCE(u.full_name, '') AS name", "COALESCE(u.email, '') AS email",
		).
		From("flow_workspace_members wm").
		LeftJoin("users u ON u.id = wm.user_id").
		Where(sq.Eq{"wm.workspace_id": workspaceID}).
		OrderBy("wm.joined_at ASC").
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("ListWorkspaceMembers build failed", logger.Error(err))
		return nil, err
	}

	rows, err := r.pool.QueryContext(ctx, q, args...)
	if err != nil {
		logger.Ctx(ctx).Error("ListWorkspaceMembers failed", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	var members []domain.WorkspaceMember
	for rows.Next() {
		var m domain.WorkspaceMember
		var invitedBy sql.NullInt64
		if err := rows.Scan(&m.ID, &m.WorkspaceID, &m.UserID, &m.Role, &invitedBy, &m.JoinedAt, &m.Name, &m.Email); err != nil {
			return nil, err
		}
		if invitedBy.Valid {
			m.InvitedBy = &invitedBy.Int64
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

func (r *Repository) UpdateWorkspaceMemberRole(ctx context.Context, input domain.UpdateWorkspaceMemberRoleInput) error {
	q, args, err := sq.
		Update("flow_workspace_members").
		Set("role", input.Role).
		Where(sq.Eq{"workspace_id": input.WorkspaceID, "user_id": input.UserID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("UpdateWorkspaceMemberRole build failed", logger.Error(err))
		return err
	}
	if _, err = r.pool.ExecContext(ctx, q, args...); err != nil {
		logger.Ctx(ctx).Error("UpdateWorkspaceMemberRole failed", logger.Error(err))
	}
	return err
}

func (r *Repository) CreatePersonalWorkspace(ctx context.Context, userID int64) (int64, error) {
	// Idempotent: do nothing if personal workspace already exists for this user
	existingQ, existingArgs, err := sq.
		Select("id").
		From("flow_workspaces").
		Where(sq.Eq{"owner_id": userID}).
		Where("is_personal = TRUE").
		Limit(1).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("CreatePersonalWorkspace build failed", logger.Error(err))
		return 0, err
	}
	var existing int64
	if scanErr := r.pool.QueryRowContext(ctx, existingQ, existingArgs...).Scan(&existing); scanErr != nil && !errors.Is(scanErr, sql.ErrNoRows) {
		logger.Ctx(ctx).Warn("CreatePersonalWorkspace lookup failed", logger.Error(scanErr))
	}
	if existing > 0 {
		return existing, nil
	}

	insertQ, insertArgs, err := sq.
		Insert("flow_workspaces").
		Columns("owner_id", "name", "icon", "color", "is_default", "is_personal", "created_by").
		Values(userID, "Personal", "🏠", "#64C5B9", true, true, userID).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("CreatePersonalWorkspace insert build failed", logger.Error(err))
		return 0, err
	}
	res, err := r.pool.ExecContext(ctx, insertQ, insertArgs...)
	if err != nil {
		logger.Ctx(ctx).Error("CreatePersonalWorkspace failed", logger.Error(err))
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		logger.Ctx(ctx).Error("CreatePersonalWorkspace last insert id failed", logger.Error(err))
		return 0, err
	}

	// Add creator as manager
	memberQ, memberArgs, err := sq.
		Insert("flow_workspace_members").
		Columns("workspace_id", "user_id", "role").
		Values(id, userID, "manager").
		Suffix("ON DUPLICATE KEY UPDATE role = VALUES(role)").
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Warn("CreatePersonalWorkspace member build failed", logger.Error(err))
		return id, nil
	}
	if _, memberErr := r.pool.ExecContext(ctx, memberQ, memberArgs...); memberErr != nil {
		logger.Ctx(ctx).Warn("CreatePersonalWorkspace add member failed", logger.Error(memberErr))
	}
	return id, nil
}

// ─────────────────────────────────────────────
// Project Members
// ─────────────────────────────────────────────

func (r *Repository) ListProjectMembers(ctx context.Context, projectID int64) ([]domain.ProjectMember, error) {
	q, args, err := sq.
		Select(
			"pm.id", "pm.project_id", "pm.user_id", "pm.role", "pm.added_by", "pm.joined_at",
			"COALESCE(u.full_name, '') AS name", "COALESCE(u.email, '') AS email",
		).
		From("flow_project_members pm").
		LeftJoin("users u ON u.id = pm.user_id").
		Where(sq.Eq{"pm.project_id": projectID}).
		OrderBy("pm.joined_at ASC").
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("ListProjectMembers build failed", logger.Error(err))
		return nil, err
	}

	rows, err := r.pool.QueryContext(ctx, q, args...)
	if err != nil {
		logger.Ctx(ctx).Error("ListProjectMembers failed", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	var members []domain.ProjectMember
	for rows.Next() {
		var m domain.ProjectMember
		var addedBy sql.NullInt64
		if err := rows.Scan(&m.ID, &m.ProjectID, &m.UserID, &m.Role, &addedBy, &m.AddedAt, &m.Name, &m.Email); err != nil {
			return nil, err
		}
		if addedBy.Valid {
			m.AddedBy = &addedBy.Int64
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

func (r *Repository) AddProjectMember(ctx context.Context, input domain.AddProjectMemberInput) error {
	role := input.Role
	if role == "" {
		role = "member"
	}
	q, args, err := sq.
		Insert("flow_project_members").
		Columns("project_id", "user_id", "role", "added_by").
		Values(input.ProjectID, input.UserID, role, input.AddedBy).
		Suffix("ON DUPLICATE KEY UPDATE role = VALUES(role), added_by = VALUES(added_by)").
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("AddProjectMember build failed", logger.Error(err))
		return err
	}
	if _, err = r.pool.ExecContext(ctx, q, args...); err != nil {
		logger.Ctx(ctx).Error("AddProjectMember failed", logger.Error(err))
	}
	return err
}

func (r *Repository) RemoveProjectMember(ctx context.Context, input domain.RemoveProjectMemberInput) error {
	q, args, err := sq.
		Delete("flow_project_members").
		Where(sq.Eq{"project_id": input.ProjectID, "user_id": input.UserID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("RemoveProjectMember build failed", logger.Error(err))
		return err
	}
	if _, err = r.pool.ExecContext(ctx, q, args...); err != nil {
		logger.Ctx(ctx).Error("RemoveProjectMember failed", logger.Error(err))
	}
	return err
}

func (r *Repository) UpdateProjectMemberRole(ctx context.Context, input domain.UpdateProjectMemberRoleInput) error {
	q, args, err := sq.
		Update("flow_project_members").
		Set("role", input.Role).
		Where(sq.Eq{"project_id": input.ProjectID, "user_id": input.UserID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("UpdateProjectMemberRole build failed", logger.Error(err))
		return err
	}
	if _, err = r.pool.ExecContext(ctx, q, args...); err != nil {
		logger.Ctx(ctx).Error("UpdateProjectMemberRole failed", logger.Error(err))
	}
	return err
}

func (r *Repository) TransferProjectOwnership(ctx context.Context, input domain.TransferProjectOwnershipInput) error {
	q, args, err := sq.
		Update("flow_projects").
		Set("owner_id", input.NewOwnerID).
		Where(sq.Eq{"id": input.ProjectID, "owner_id": input.CallerID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("TransferProjectOwnership build failed", logger.Error(err))
		return err
	}
	if _, err = r.pool.ExecContext(ctx, q, args...); err != nil {
		logger.Ctx(ctx).Error("TransferProjectOwnership failed", logger.Error(err))
	}
	return err
}

func (r *Repository) MoveProject(ctx context.Context, input domain.MoveProjectInput) error {
	q, args, err := sq.
		Update("flow_projects").
		Set("workspace_id", input.WorkspaceID).
		Where(sq.Eq{"id": input.ProjectID, "owner_id": input.CallerID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("MoveProject build failed", logger.Error(err))
		return err
	}
	if _, err = r.pool.ExecContext(ctx, q, args...); err != nil {
		logger.Ctx(ctx).Error("MoveProject failed", logger.Error(err))
	}
	return err
}

func (r *Repository) GetUserWorkspaceRole(ctx context.Context, workspaceID, userID int64) (string, error) {
	q, args, err := sq.
		Select("role").
		From("flow_workspace_members").
		Where(sq.Eq{"workspace_id": workspaceID, "user_id": userID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return "", err
	}
	var role string
	if err := r.pool.QueryRowContext(ctx, q, args...).Scan(&role); err != nil {
		return "", err
	}
	return role, nil
}

// ─── Project Columns ─────────────────────────────────────────────────────────

func (r *Repository) GetColumnsByProject(ctx context.Context, projectID int64) ([]domain.ProjectColumn, error) {
	q, args, err := sq.
		Select("id", "project_id", "workspace_id", "name", "color", "sort_order", "created_by", "created_at", "updated_at").
		From("flow_project_columns").
		Where(sq.Eq{"project_id": projectID}).
		OrderBy("sort_order ASC").
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("GetColumnsByProject build failed", logger.Error(err))
		return nil, err
	}

	rows, err := r.pool.QueryContext(ctx, q, args...)
	if err != nil {
		logger.Ctx(ctx).Error("GetColumnsByProject failed", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	var cols []domain.ProjectColumn
	for rows.Next() {
		var c domain.ProjectColumn
		if err := rows.Scan(&c.ID, &c.ProjectID, &c.WorkspaceID, &c.Name, &c.Color,
			&c.SortOrder, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		cols = append(cols, c)
	}
	return cols, rows.Err()
}

func (r *Repository) CreateColumn(ctx context.Context, input domain.ProjectColumnInput) (int64, error) {
	q, args, err := sq.
		Insert("flow_project_columns").
		Columns("project_id", "workspace_id", "name", "color", "sort_order", "created_by").
		Values(input.ProjectID, input.WorkspaceID, input.Name, input.Color, input.SortOrder, input.CreatedBy).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("CreateColumn build failed", logger.Error(err))
		return 0, err
	}
	res, err := r.pool.ExecContext(ctx, q, args...)
	if err != nil {
		logger.Ctx(ctx).Error("CreateColumn failed", logger.Error(err))
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) UpdateColumn(ctx context.Context, input domain.UpdateProjectColumnInput) error {
	q, args, err := sq.
		Update("flow_project_columns").
		Set("name", input.Name).
		Set("color", input.Color).
		Set("sort_order", input.SortOrder).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": input.ID}).
		Where(workspaceMemberPredicate("flow_project_columns"), input.ActorID, input.ActorID).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("UpdateColumn build failed", logger.Error(err))
		return err
	}
	res, err := r.pool.ExecContext(ctx, q, args...)
	if err != nil {
		logger.Ctx(ctx).Error("UpdateColumn failed", logger.Error(err))
		return err
	}
	return requireAffected(res)
}

func (r *Repository) DeleteColumn(ctx context.Context, id, actorID int64) error {
	// Null out section_id for tasks in this column before deleting
	clearQ, clearArgs, err := sq.
		Update("flow_tasks").
		Set("section_id", nil).
		Where(sq.Eq{"section_id": id}).
		Where(workspaceMemberPredicate("flow_tasks"), actorID, actorID).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("DeleteColumn clear build failed", logger.Error(err))
		return err
	}
	if _, err := r.pool.ExecContext(ctx, clearQ, clearArgs...); err != nil {
		return err
	}

	delQ, delArgs, err := sq.
		Delete("flow_project_columns").
		Where(sq.Eq{"id": id}).
		Where(workspaceMemberPredicate("flow_project_columns"), actorID, actorID).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("DeleteColumn build failed", logger.Error(err))
		return err
	}
	res, err := r.pool.ExecContext(ctx, delQ, delArgs...)
	if err != nil {
		logger.Ctx(ctx).Error("DeleteColumn failed", logger.Error(err))
		return err
	}
	return requireAffected(res)
}

func (r *Repository) MoveTaskToColumn(ctx context.Context, taskID, columnID, actorID int64) error {
	var sectionVal interface{}
	if columnID == 0 {
		sectionVal = nil
	} else {
		sectionVal = columnID
	}
	q, args, err := sq.
		Update("flow_tasks").
		Set("section_id", sectionVal).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": taskID}).
		Where(workspaceMemberPredicate("flow_tasks"), actorID, actorID).
		Where("(? = 0 OR EXISTS (SELECT 1 FROM flow_project_columns c WHERE c.id = ? AND c.workspace_id = flow_tasks.workspace_id))", columnID, columnID).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("MoveTaskToColumn build failed", logger.Error(err))
		return err
	}
	res, err := r.pool.ExecContext(ctx, q, args...)
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
		if _, err := r.pool.ExecContext(ctx,
			`INSERT INTO flow_project_columns (project_id, workspace_id, name, color, sort_order, created_by)
			 VALUES (?, ?, ?, ?, ?, ?)`,
			projectID, workspaceID, d.name, d.color, d.sortOrder, createdBy,
		); err != nil {
			return err
		}
	}
	return nil
}

// ─────────────────────────────────────────────
// Task Search
// ─────────────────────────────────────────────

func (r *Repository) SearchTasks(ctx context.Context, workspaceID int64, q string) ([]domain.Task, error) {
	like := "%" + q + "%"
	query := `SELECT id, project_id, workspace_id, owner_id, parent_task_id, section_id, title, description,
	           status, priority, assignee_id, reporter_id, due_date, start_date, completed_at,
	           estimated_hours, actual_hours, sort_order, labels, created_at, updated_at
	          FROM flow_tasks
	          WHERE workspace_id = ?
	            AND archived = FALSE
	            AND (title LIKE ? OR description LIKE ?)
	          ORDER BY updated_at DESC
	          LIMIT 50`

	rows, err := r.pool.QueryContext(ctx, query, workspaceID, like, like)
	if err != nil {
		logger.Ctx(ctx).Error("SearchTasks failed", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	return scanTasksRaw(ctx, rows)
}

// ─────────────────────────────────────────────
// User lookup
// ─────────────────────────────────────────────

func (r *Repository) LookupUserByEmail(ctx context.Context, email string) (int64, bool, error) {
	var id int64
	err := r.pool.QueryRowContext(ctx, `SELECT id FROM users WHERE email = ? LIMIT 1`, email).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		logger.Ctx(ctx).Error("LookupUserByEmail failed", logger.Error(err))
		return 0, false, err
	}
	return id, true, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Membership checks (used by realtime topic auth).
// Kept cheap — a single indexed EXISTS query each; callers (the WS hub)
// run them per subscribe, which is infrequent compared to per-event dispatch.
// ─────────────────────────────────────────────────────────────────────────────

// IsWorkspaceMember returns true when the user either owns the workspace
// or has a row in flow_workspace_members.
func (r *Repository) IsWorkspaceMember(ctx context.Context, userID, workspaceID int64) bool {
	if userID == 0 || workspaceID == 0 {
		return false
	}
	var exists int
	err := r.pool.QueryRowContext(ctx, `
		SELECT 1 FROM flow_workspaces WHERE id = ? AND owner_id = ?
		UNION
		SELECT 1 FROM flow_workspace_members WHERE workspace_id = ? AND user_id = ?
		LIMIT 1`,
		workspaceID, userID, workspaceID, userID,
	).Scan(&exists)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Ctx(ctx).Warn("IsWorkspaceMember check failed", logger.Error(err))
	}
	return exists == 1
}

// GetMyTasksProject returns the id and current name of the caller's
// auto-provisioned personal project inside a workspace, or (0, "") if none
// exists. Identified by name — "My Tasks" now, with "Inbox" recognised as the
// legacy name so existing data is reused (the service migrates it on access).
func (r *Repository) GetMyTasksProject(ctx context.Context, userID, workspaceID int64) (int64, string) {
	var id int64
	var name string
	err := r.pool.QueryRowContext(ctx,
		`SELECT id, name FROM flow_projects
		 WHERE owner_id = ? AND workspace_id = ? AND name IN ('My Tasks', 'Inbox')
		 ORDER BY (name = 'My Tasks') DESC, created_at ASC LIMIT 1`,
		userID, workspaceID,
	).Scan(&id, &name)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Ctx(ctx).Warn("GetMyTasksProject failed", logger.Error(err))
	}
	return id, name
}

// RenameProject updates a project's name. Used to migrate the legacy "Inbox"
// personal project to "My Tasks" on access.
func (r *Repository) RenameProject(ctx context.Context, projectID int64, name string) error {
	_, err := r.pool.ExecContext(ctx,
		`UPDATE flow_projects SET name = ? WHERE id = ?`, name, projectID)
	return err
}

// GetProjectWorkspace returns the workspace id of the given project, or
// 0 if the project doesn't exist. Used by the realtime publish paths.
func (r *Repository) GetProjectWorkspace(ctx context.Context, projectID int64) int64 {
	var workspaceID int64
	err := r.pool.QueryRowContext(ctx,
		`SELECT workspace_id FROM flow_projects WHERE id = ? LIMIT 1`,
		projectID,
	).Scan(&workspaceID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Ctx(ctx).Warn("GetProjectWorkspace failed", logger.Error(err))
	}
	return workspaceID
}

// GetColumnProject returns the project id that owns the given column,
// or 0 if the column is missing. Used by the realtime publish paths to
// find the correct topic when the handler only has the column id
// (update/delete).
func (r *Repository) GetColumnProject(ctx context.Context, columnID int64) int64 {
	var projectID int64
	err := r.pool.QueryRowContext(ctx,
		`SELECT project_id FROM flow_project_columns WHERE id = ? LIMIT 1`,
		columnID,
	).Scan(&projectID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Ctx(ctx).Warn("GetColumnProject failed", logger.Error(err))
	}
	return projectID
}

// IsProjectMember returns true when the user can access the project —
// either because they're listed on flow_project_members or because they
// belong to the project's workspace. This matches the read-access model
// the REST endpoints enforce.
func (r *Repository) IsProjectMember(ctx context.Context, userID, projectID int64) bool {
	if userID == 0 || projectID == 0 {
		return false
	}
	var exists int
	err := r.pool.QueryRowContext(ctx, `
		SELECT 1 FROM flow_project_members WHERE project_id = ? AND user_id = ?
		UNION
		SELECT 1
		FROM flow_projects p
		LEFT JOIN flow_workspace_members wm
		  ON wm.workspace_id = p.workspace_id AND wm.user_id = ?
		WHERE p.id = ? AND (p.owner_id = ? OR wm.user_id IS NOT NULL)
		LIMIT 1`,
		projectID, userID, userID, projectID, userID,
	).Scan(&exists)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Ctx(ctx).Warn("IsProjectMember check failed", logger.Error(err))
	}
	return exists == 1
}
