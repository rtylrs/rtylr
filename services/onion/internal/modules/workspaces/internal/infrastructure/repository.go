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
	rows, err := r.queries.GetProjectsByWorkspace(ctx, sqlc.GetProjectsByWorkspaceParams{
		WorkspaceID: workspaceID,
		BusinessID:  businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("GetProjectsByWorkspace failed", logger.Error(err))
		return nil, err
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
			TaskCount:      int32(countIfaceToInt64(row.TaskCount)),
			CompletedCount: int32(countIfaceToInt64(row.CompletedCount)),
			MemberCount:    int32(countIfaceToInt64(row.MemberCount)),
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
	return projects, nil
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

// dynamic-query: squirrel
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

// dynamic-query: squirrel
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

// dynamic-query: squirrel
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

// dynamic-query: squirrel
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
	return r.queries.GetCommentTaskID(ctx, commentID)
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
		TaskID:     input.TaskID,
		OwnerID:    input.OwnerID,
		BusinessID: input.BusinessID,
		Name:       input.Name,
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
	color := input.Color
	if color == "" {
		color = "#6366f1"
	}
	res, err := r.queries.UpdateLabel(ctx, sqlc.UpdateLabelParams{
		Name:    input.Name,
		Color:   color,
		ID:      id,
		OwnerID: actorID,
		UserID:  actorID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("UpdateLabel failed", logger.Error(err))
		return err
	}
	return requireAffected(res)
}

func (r *Repository) DeleteLabel(ctx context.Context, id, actorID int64) error {
	res, err := r.queries.DeleteLabel(ctx, sqlc.DeleteLabelParams{
		ID:      id,
		OwnerID: actorID,
		UserID:  actorID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("DeleteLabel failed", logger.Error(err))
		return err
	}
	return requireAffected(res)
}

func (r *Repository) SetTaskLabels(ctx context.Context, taskID, actorID int64, labelIDs []int64) error {
	labelsJSON, err := json.Marshal(labelIDs)
	if err != nil {
		return err
	}
	res, err := r.queries.SetTaskLabels(ctx, sqlc.SetTaskLabelsParams{
		Labels:  labelsJSON,
		ID:      taskID,
		OwnerID: actorID,
		UserID:  actorID,
	})
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
	return r.queries.GetTimeLogTaskID(ctx, id)
}

func (r *Repository) DeleteTimeLog(ctx context.Context, id, userID int64) error {
	err := r.queries.DeleteTimeLog(ctx, sqlc.DeleteTimeLogParams{ID: id, UserID: userID})
	if err != nil {
		logger.Ctx(ctx).Error("DeleteTimeLog failed", logger.Error(err))
	}
	return err
}

func (r *Repository) RecalcActualHours(ctx context.Context, taskID int64) error {
	err := r.queries.RecalcActualHours(ctx, sqlc.RecalcActualHoursParams{
		TaskID: taskID,
		ID:     taskID,
	})
	if err != nil {
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
	// cross-schema read from users — workspace member display names
	rows, err := r.queries.ListWorkspaceMembers(ctx, workspaceID)
	if err != nil {
		logger.Ctx(ctx).Error("ListWorkspaceMembers failed", logger.Error(err))
		return nil, err
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
	return members, nil
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

func (r *Repository) CreatePersonalWorkspace(ctx context.Context, userID int64) (int64, error) {
	existing, err := r.queries.GetPersonalWorkspaceByOwner(ctx, userID)
	if err == nil && existing > 0 {
		return existing, nil
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Ctx(ctx).Warn("CreatePersonalWorkspace lookup failed", logger.Error(err))
	}

	res, err := r.queries.CreatePersonalWorkspace(ctx, sqlc.CreatePersonalWorkspaceParams{
		OwnerID:   userID,
		CreatedBy: userID,
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
		logger.Ctx(ctx).Warn("CreatePersonalWorkspace add member failed", logger.Error(addErr))
	}
	return id, nil
}

// ─────────────────────────────────────────────
// Project Members
// ─────────────────────────────────────────────

func (r *Repository) ListProjectMembers(ctx context.Context, projectID int64) ([]domain.ProjectMember, error) {
	rows, err := r.queries.ListProjectMembers(ctx, projectID)
	if err != nil {
		logger.Ctx(ctx).Error("ListProjectMembers failed", logger.Error(err))
		return nil, err
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
	return members, nil
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

// ─── Project Columns ─────────────────────────────────────────────────────────

func (r *Repository) GetColumnsByProject(ctx context.Context, projectID int64) ([]domain.ProjectColumn, error) {
	rows, err := r.queries.GetColumnsByProject(ctx, projectID)
	if err != nil {
		logger.Ctx(ctx).Error("GetColumnsByProject failed", logger.Error(err))
		return nil, err
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
	return cols, nil
}

func (r *Repository) CreateColumn(ctx context.Context, input domain.ProjectColumnInput) (int64, error) {
	res, err := r.queries.CreateColumn(ctx, sqlc.CreateColumnParams{
		ProjectID:   input.ProjectID,
		WorkspaceID: input.WorkspaceID,
		Name:        input.Name,
		Color:       sql.NullString{String: input.Color, Valid: input.Color != ""},
		SortOrder:   sql.NullInt32{Int32: int32(input.SortOrder), Valid: true},
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
		SortOrder: sql.NullInt32{Int32: int32(input.SortOrder), Valid: true},
		ID:        input.ID,
		OwnerID:   input.ActorID,
		UserID:    input.ActorID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("UpdateColumn failed", logger.Error(err))
		return err
	}
	return requireAffected(res)
}

func (r *Repository) DeleteColumn(ctx context.Context, id, actorID int64) error {
	if err := r.queries.ClearTasksSectionID(ctx, sqlc.ClearTasksSectionIDParams{
		SectionID: sql.NullInt64{Int64: id, Valid: true},
		OwnerID:   actorID,
		UserID:    actorID,
	}); err != nil {
		return err
	}

	res, err := r.queries.DeleteColumn(ctx, sqlc.DeleteColumnParams{
		ID:      id,
		OwnerID: actorID,
		UserID:  actorID,
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
		OwnerID:   actorID,
		UserID:    actorID,
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
			SortOrder:   sql.NullInt32{Int32: int32(d.sortOrder), Valid: true},
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
func (r *Repository) SearchTasks(ctx context.Context, workspaceID int64, q string) ([]domain.Task, error) {
	q = strings.TrimSpace(q)
	if q == "" {
		return []domain.Task{}, nil
	}
	searchQuery := searchutil.BooleanPrefixQuery(q)
	if searchQuery == "" {
		return []domain.Task{}, nil
	}
	query := `SELECT id, project_id, workspace_id, owner_id, parent_task_id, section_id, title, description,
	           status, priority, assignee_id, reporter_id, due_date, start_date, completed_at,
	           estimated_hours, actual_hours, sort_order, labels, created_at, updated_at
	          FROM flow_tasks
	          WHERE workspace_id = ?
	            AND archived = FALSE
	            AND MATCH(title, description) AGAINST (? IN BOOLEAN MODE)
	          ORDER BY updated_at DESC
	          LIMIT 50`

	rows, err := r.pool.QueryContext(ctx, query, workspaceID, searchQuery)
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
	id, err := r.queries.LookupUserByEmail(ctx, email)
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
	member, err := r.queries.IsWorkspaceMember(ctx, sqlc.IsWorkspaceMemberParams{
		ID:          workspaceID,
		OwnerID:     userID,
		WorkspaceID: workspaceID,
		UserID:      userID,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Ctx(ctx).Warn("IsWorkspaceMember check failed", logger.Error(err))
	}
	return member.Bool
}

func (r *Repository) GetMyTasksProject(ctx context.Context, userID, workspaceID int64) (int64, string) {
	row, err := r.queries.GetMyTasksProject(ctx, sqlc.GetMyTasksProjectParams{
		OwnerID:     userID,
		WorkspaceID: workspaceID,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Ctx(ctx).Warn("GetMyTasksProject failed", logger.Error(err))
		return 0, ""
	}
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ""
	}
	return row.ID, row.Name
}

func (r *Repository) RenameProject(ctx context.Context, projectID int64, name string) error {
	return r.queries.RenameProject(ctx, sqlc.RenameProjectParams{Name: name, ID: projectID})
}

func (r *Repository) GetProjectWorkspace(ctx context.Context, projectID int64) int64 {
	workspaceID, err := r.queries.GetProjectWorkspace(ctx, projectID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Ctx(ctx).Warn("GetProjectWorkspace failed", logger.Error(err))
	}
	return workspaceID
}

func (r *Repository) GetColumnProject(ctx context.Context, columnID int64) int64 {
	projectID, err := r.queries.GetColumnProject(ctx, columnID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Ctx(ctx).Warn("GetColumnProject failed", logger.Error(err))
	}
	return projectID
}

func (r *Repository) IsProjectMember(ctx context.Context, userID, projectID int64) bool {
	if userID == 0 || projectID == 0 {
		return false
	}
	member, err := r.queries.IsProjectMember(ctx, sqlc.IsProjectMemberParams{
		ProjectID: projectID,
		UserID:    userID,
		UserID_2:  userID,
		ID:        projectID,
		OwnerID:   userID,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Ctx(ctx).Warn("IsProjectMember check failed", logger.Error(err))
	}
	return member.Bool
}
