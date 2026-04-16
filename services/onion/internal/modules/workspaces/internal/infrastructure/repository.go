package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/workspaces/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	pool := dbclient.Pool()
	return &Repository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

// ─────────────────────────────────────────────
// Workspaces
// ─────────────────────────────────────────────

func (r *Repository) CountWorkspacesByOwner(ctx context.Context, ownerID int64) (int64, error) {
	count, err := r.queries.CountWorkspacesByOwner(ctx, ownerID)
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

func (r *Repository) GetWorkspacesByOwner(ctx context.Context, ownerID int64) ([]domain.Workspace, error) {
	rows, err := r.queries.GetWorkspacesByOwner(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("GetWorkspacesByOwner failed", logger.Error(err))
		return nil, err
	}

	var workspaces []domain.Workspace
	for _, row := range rows {
		workspaces = append(workspaces, domain.Workspace{
			ID:          row.ID,
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
		OwnerID:     input.OwnerID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("UpdateWorkspace failed", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) DeleteWorkspace(ctx context.Context, id, ownerID int64) error {
	err := r.queries.DeleteWorkspace(ctx, sqlc.DeleteWorkspaceParams{
		ID:      id,
		OwnerID: ownerID,
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

func (r *Repository) GetProjectsByWorkspace(ctx context.Context, workspaceID, ownerID int64) ([]domain.Project, error) {
	rows, err := r.queries.GetProjectsByWorkspace(ctx, sqlc.GetProjectsByWorkspaceParams{
		WorkspaceID: workspaceID,
		OwnerID:     ownerID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("GetProjectsByWorkspace failed", logger.Error(err))
		return nil, err
	}

	var projects []domain.Project
	for _, row := range rows {
		p := domain.Project{
			ID:          row.ID,
			WorkspaceID: row.WorkspaceID,
			OwnerID:     row.OwnerID,
			Name:        row.Name,
			Description: row.Description.String,
			Icon:        row.Icon.String,
			Color:       row.Color.String,
			Status:      string(row.Status.FlowProjectsStatus),
			CreatedBy:   row.CreatedBy,
			CreatedAt:   row.CreatedAt.Time,
			UpdatedAt:   row.UpdatedAt.Time,
		}
		if row.StartDate.Valid {
			p.StartDate = &row.StartDate.Time
		}
		if row.DueDate.Valid {
			p.DueDate = &row.DueDate.Time
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
		OwnerID:     input.OwnerID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("UpdateProject failed", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) ArchiveProject(ctx context.Context, id, ownerID int64) error {
	err := r.queries.ArchiveProject(ctx, sqlc.ArchiveProjectParams{
		ID:      id,
		OwnerID: ownerID,
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
	// sqlc query doesn't support optional filters, so for complex cases
	// we use the basic query and filter in-memory, or keep raw SQL for status/assignee/priority.
	if filter.Status != "" || filter.AssigneeID != nil || filter.Priority != "" {
		return r.getTasksByProjectFiltered(ctx, filter)
	}

	rows, err := r.queries.GetTasksByProject(ctx, filter.ProjectID)
	if err != nil {
		logger.Ctx(ctx).Error("GetTasksByProject failed", logger.Error(err))
		return nil, err
	}
	return tasksFromRows(rows), nil
}

func (r *Repository) getTasksByProjectFiltered(ctx context.Context, filter domain.ListTasksFilter) ([]domain.Task, error) {
	query := `SELECT id, project_id, workspace_id, owner_id, parent_task_id, section_id, title, description,
	           status, priority, assignee_id, reporter_id, due_date, start_date, completed_at,
	           estimated_hours, actual_hours, sort_order, labels, created_at, updated_at
	          FROM flow_tasks WHERE project_id = ? AND parent_task_id IS NULL`
	args := []interface{}{filter.ProjectID}

	if filter.Status != "" {
		query += " AND status = ?"
		args = append(args, filter.Status)
	}
	if filter.AssigneeID != nil {
		query += " AND assignee_id = ?"
		args = append(args, *filter.AssigneeID)
	}
	if filter.Priority != "" {
		query += " AND priority = ?"
		args = append(args, filter.Priority)
	}

	query += " ORDER BY sort_order ASC, created_at ASC"

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Ctx(ctx).Error("GetTasksByProject filtered failed", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	return scanTasksRaw(ctx, rows)
}

func (r *Repository) GetTasksByWorkspace(ctx context.Context, filter domain.ListTasksFilter) ([]domain.Task, error) {
	query := `SELECT id, project_id, workspace_id, owner_id, parent_task_id, section_id, title, description,
	           status, priority, assignee_id, reporter_id, due_date, start_date, completed_at,
	           estimated_hours, actual_hours, sort_order, labels, created_at, updated_at
	          FROM flow_tasks WHERE workspace_id = ? AND parent_task_id IS NULL`
	args := []interface{}{filter.WorkspaceID}

	if filter.Status != "" {
		query += " AND status = ?"
		args = append(args, filter.Status)
	}
	if filter.AssigneeID != nil {
		query += " AND assignee_id = ?"
		args = append(args, *filter.AssigneeID)
	}
	if filter.Priority != "" {
		query += " AND priority = ?"
		args = append(args, filter.Priority)
	}

	query += " ORDER BY created_at DESC LIMIT 500"

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Ctx(ctx).Error("GetTasksByWorkspace failed", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	return scanTasksRaw(ctx, rows)
}

func (r *Repository) GetTaskByID(ctx context.Context, taskID int64) (domain.Task, error) {
	row, err := r.queries.GetTaskByID(ctx, taskID)
	if err != nil {
		return domain.Task{}, err
	}
	return taskFromFlowTask(row), nil
}

func (r *Repository) GetSubtasks(ctx context.Context, parentTaskID int64) ([]domain.Task, error) {
	rows, err := r.queries.GetSubtasks(ctx, sql.NullInt64{Int64: parentTaskID, Valid: true})
	if err != nil {
		logger.Ctx(ctx).Error("GetSubtasks failed", logger.Error(err))
		return nil, err
	}
	return tasksFromRows(rows), nil
}

func (r *Repository) GetMyTasks(ctx context.Context, userID int64) ([]domain.Task, error) {
	rows, err := r.queries.GetMyTasks(ctx, sql.NullInt64{Int64: userID, Valid: true})
	if err != nil {
		logger.Ctx(ctx).Error("GetMyTasks failed", logger.Error(err))
		return nil, err
	}
	return tasksFromRows(rows), nil
}

func (r *Repository) UpdateTask(ctx context.Context, input domain.UpdateTaskInput) error {
	labelsJSON, err := marshalLabels(input.Labels)
	if err != nil {
		return err
	}

	err = r.queries.UpdateTask(ctx, sqlc.UpdateTaskParams{
		Title:       input.Title,
		Description: sql.NullString{String: input.Description, Valid: input.Description != ""},
		Status:      sqlc.NullFlowTasksStatus{FlowTasksStatus: sqlc.FlowTasksStatus(input.Status), Valid: input.Status != ""},
		Priority:    sqlc.NullFlowTasksPriority{FlowTasksPriority: sqlc.FlowTasksPriority(input.Priority), Valid: input.Priority != ""},
		AssigneeID:  nullInt64Ptr(input.AssigneeID),
		DueDate:     nullTimePtrToSql(input.DueDate),
		StartDate:   nullTimePtrToSql(input.StartDate),
		Labels:      labelsJSON,
		Column9:     input.Status,
		Column10:    input.Status,
		ID:          input.ID,
		OwnerID:     input.OwnerID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("UpdateTask failed", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) UpdateTaskStatus(ctx context.Context, input domain.UpdateTaskStatusInput) error {
	err := r.queries.UpdateTaskStatus(ctx, sqlc.UpdateTaskStatusParams{
		Status:  sqlc.NullFlowTasksStatus{FlowTasksStatus: sqlc.FlowTasksStatus(input.Status), Valid: true},
		Column2: input.Status,
		Column3: input.Status,
		ID:      input.ID,
		OwnerID: input.OwnerID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("UpdateTaskStatus failed", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) DeleteTask(ctx context.Context, taskID, ownerID int64) error {
	err := r.queries.DeleteTask(ctx, sqlc.DeleteTaskParams{
		ID:      taskID,
		OwnerID: ownerID,
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

	var comments []domain.TaskComment
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

	var activities []domain.TaskActivity
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

	var attachments []domain.TaskAttachment
	for _, row := range rows {
		attachments = append(attachments, domain.TaskAttachment{
			ID:         row.ID,
			TaskID:     row.TaskID,
			OwnerID:    row.OwnerID,
			Name:       row.Name,
			URL:        row.Url,
			Size:       row.Size.Int64,
			UploadedBy: row.UploadedBy,
			CreatedAt:  row.CreatedAt.Time,
		})
	}
	return attachments, nil
}

func (r *Repository) DeleteAttachment(ctx context.Context, attachmentID, ownerID int64) error {
	err := r.queries.DeleteAttachment(ctx, sqlc.DeleteAttachmentParams{
		ID:      attachmentID,
		OwnerID: ownerID,
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

	var logs []domain.TimeLog
	for _, row := range rows {
		logs = append(logs, domain.TimeLog{
			ID:              row.ID,
			OwnerID:         row.OwnerID,
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

	var labels []domain.Label
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

func (r *Repository) UpdateLabel(ctx context.Context, id int64, input domain.LabelInput) error {
	_, err := r.pool.ExecContext(ctx,
		`UPDATE flow_labels SET name = ?, color = ? WHERE id = ?`,
		input.Name, input.Color, id,
	)
	if err != nil {
		logger.Ctx(ctx).Error("UpdateLabel failed", logger.Error(err))
	}
	return err
}

func (r *Repository) DeleteLabel(ctx context.Context, id int64) error {
	_, err := r.pool.ExecContext(ctx,
		`DELETE FROM flow_labels WHERE id = ?`, id,
	)
	if err != nil {
		logger.Ctx(ctx).Error("DeleteLabel failed", logger.Error(err))
	}
	return err
}

func (r *Repository) SetTaskLabels(ctx context.Context, taskID int64, labelIDs []int64) error {
	b, err := json.Marshal(labelIDs)
	if err != nil {
		return err
	}
	_, err = r.pool.ExecContext(ctx,
		`UPDATE flow_tasks SET labels = ? WHERE id = ?`, b, taskID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("SetTaskLabels failed", logger.Error(err))
	}
	return err
}

// ─────────────────────────────────────────────
// Time Logs (extended)
// ─────────────────────────────────────────────

func (r *Repository) GetTimeLogTaskID(ctx context.Context, id int64) (int64, error) {
	var taskID int64
	err := r.pool.QueryRowContext(ctx,
		`SELECT task_id FROM flow_time_logs WHERE id = ?`, id,
	).Scan(&taskID)
	return taskID, err
}

func (r *Repository) DeleteTimeLog(ctx context.Context, id, userID int64) error {
	_, err := r.pool.ExecContext(ctx,
		`DELETE FROM flow_time_logs WHERE id = ? AND user_id = ?`, id, userID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("DeleteTimeLog failed", logger.Error(err))
	}
	return err
}

func (r *Repository) RecalcActualHours(ctx context.Context, taskID int64) error {
	_, err := r.pool.ExecContext(ctx,
		`UPDATE flow_tasks
		 SET actual_hours = (SELECT COALESCE(SUM(duration_minutes), 0) / 60.0
		                     FROM flow_time_logs WHERE task_id = ?)
		 WHERE id = ?`,
		taskID, taskID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("RecalcActualHours failed", logger.Error(err))
	}
	return err
}

// ─────────────────────────────────────────────
// Dashboard
// ─────────────────────────────────────────────

func (r *Repository) GetDashboardStats(ctx context.Context, workspaceID, ownerID int64) (domain.DashboardStats, error) {
	var stats domain.DashboardStats

	weekStart := time.Now().AddDate(0, 0, -int(time.Now().Weekday()))
	weekEnd := weekStart.AddDate(0, 0, 7)

	row, err := r.queries.GetDashboardStats(ctx, sqlc.GetDashboardStatsParams{
		CompletedAt:   sql.NullTime{Time: weekStart, Valid: true},
		CompletedAt_2: sql.NullTime{Time: weekEnd, Valid: true},
		WorkspaceID:   workspaceID,
		OwnerID:       ownerID,
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

func taskFromFlowTask(row sqlc.FlowTask) domain.Task {
	t := domain.Task{
		ID:          row.ID,
		ProjectID:   row.ProjectID,
		WorkspaceID: row.WorkspaceID,
		OwnerID:     row.OwnerID,
		Title:       row.Title,
		Description: row.Description.String,
		Status:      string(row.Status.FlowTasksStatus),
		Priority:    string(row.Priority.FlowTasksPriority),
		ReporterID:  row.ReporterID,
		SortOrder:   int(row.SortOrder.Int32),
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}

	if row.ParentTaskID.Valid {
		t.ParentTaskID = &row.ParentTaskID.Int64
	}
	if row.AssigneeID.Valid {
		t.AssigneeID = &row.AssigneeID.Int64
	}
	if row.DueDate.Valid {
		t.DueDate = &row.DueDate.Time
	}
	if row.StartDate.Valid {
		t.StartDate = &row.StartDate.Time
	}
	if row.CompletedAt.Valid {
		t.CompletedAt = &row.CompletedAt.Time
	}
	if row.EstimatedHours.Valid {
		f, _ := strconv.ParseFloat(row.EstimatedHours.String, 64)
		t.EstimatedHours = &f
	}
	if row.ActualHours.Valid {
		f, _ := strconv.ParseFloat(row.ActualHours.String, 64)
		t.ActualHours = &f
	}
	if len(row.Labels) > 0 {
		_ = json.Unmarshal(row.Labels, &t.Labels)
	}

	return t
}

func tasksFromRows(rows []sqlc.FlowTask) []domain.Task {
	var tasks []domain.Task
	for _, row := range rows {
		tasks = append(tasks, taskFromFlowTask(row))
	}
	return tasks
}

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
			_ = json.Unmarshal(labelsRaw, &t.Labels)
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// ─────────────────────────────────────────────
// Workspace Members (extended)
// ─────────────────────────────────────────────

func (r *Repository) ListWorkspaceMembers(ctx context.Context, workspaceID int64) ([]domain.WorkspaceMember, error) {
	q := `
		SELECT wm.id, wm.workspace_id, wm.user_id, wm.role, wm.invited_by, wm.joined_at,
		       COALESCE(u.full_name, '') AS name, COALESCE(u.email, '') AS email
		FROM flow_workspace_members wm
		LEFT JOIN users u ON u.id = wm.user_id
		WHERE wm.workspace_id = ?
		ORDER BY wm.joined_at ASC`

	rows, err := r.pool.QueryContext(ctx, q, workspaceID)
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
	_, err := r.pool.ExecContext(ctx,
		`UPDATE flow_workspace_members SET role = ? WHERE workspace_id = ? AND user_id = ?`,
		input.Role, input.WorkspaceID, input.UserID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("UpdateWorkspaceMemberRole failed", logger.Error(err))
	}
	return err
}

func (r *Repository) CreatePersonalWorkspace(ctx context.Context, userID int64) (int64, error) {
	// Idempotent: do nothing if personal workspace already exists for this user
	var existing int64
	_ = r.pool.QueryRowContext(ctx,
		`SELECT id FROM flow_workspaces WHERE owner_id = ? AND is_personal = TRUE LIMIT 1`, userID,
	).Scan(&existing)
	if existing > 0 {
		return existing, nil
	}

	res, err := r.pool.ExecContext(ctx,
		`INSERT INTO flow_workspaces (owner_id, name, icon, color, is_default, is_personal, created_by)
		 VALUES (?, 'Personal', '🏠', '#64C5B9', TRUE, TRUE, ?)`,
		userID, userID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("CreatePersonalWorkspace failed", logger.Error(err))
		return 0, err
	}
	id, _ := res.LastInsertId()

	// Add creator as manager
	_, _ = r.pool.ExecContext(ctx,
		`INSERT INTO flow_workspace_members (workspace_id, user_id, role) VALUES (?, ?, 'manager')
		 ON DUPLICATE KEY UPDATE role = VALUES(role)`,
		id, userID,
	)
	return id, nil
}

// ─────────────────────────────────────────────
// Project Members
// ─────────────────────────────────────────────

func (r *Repository) ListProjectMembers(ctx context.Context, projectID int64) ([]domain.ProjectMember, error) {
	q := `
		SELECT pm.id, pm.project_id, pm.user_id, pm.role, pm.added_by, pm.joined_at,
		       COALESCE(u.full_name, '') AS name, COALESCE(u.email, '') AS email
		FROM flow_project_members pm
		LEFT JOIN users u ON u.id = pm.user_id
		WHERE pm.project_id = ?
		ORDER BY pm.joined_at ASC`

	rows, err := r.pool.QueryContext(ctx, q, projectID)
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
	_, err := r.pool.ExecContext(ctx,
		`INSERT INTO flow_project_members (project_id, user_id, role, added_by)
		 VALUES (?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE role = VALUES(role), added_by = VALUES(added_by)`,
		input.ProjectID, input.UserID, role, input.AddedBy,
	)
	if err != nil {
		logger.Ctx(ctx).Error("AddProjectMember failed", logger.Error(err))
	}
	return err
}

func (r *Repository) RemoveProjectMember(ctx context.Context, input domain.RemoveProjectMemberInput) error {
	_, err := r.pool.ExecContext(ctx,
		`DELETE FROM flow_project_members WHERE project_id = ? AND user_id = ?`,
		input.ProjectID, input.UserID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("RemoveProjectMember failed", logger.Error(err))
	}
	return err
}

func (r *Repository) UpdateProjectMemberRole(ctx context.Context, input domain.UpdateProjectMemberRoleInput) error {
	_, err := r.pool.ExecContext(ctx,
		`UPDATE flow_project_members SET role = ? WHERE project_id = ? AND user_id = ?`,
		input.Role, input.ProjectID, input.UserID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("UpdateProjectMemberRole failed", logger.Error(err))
	}
	return err
}

func (r *Repository) TransferProjectOwnership(ctx context.Context, input domain.TransferProjectOwnershipInput) error {
	_, err := r.pool.ExecContext(ctx,
		`UPDATE flow_projects SET owner_id = ? WHERE id = ? AND owner_id = ?`,
		input.NewOwnerID, input.ProjectID, input.CallerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("TransferProjectOwnership failed", logger.Error(err))
	}
	return err
}

func (r *Repository) GetUserWorkspaceRole(ctx context.Context, workspaceID, userID int64) (string, error) {
	var role string
	err := r.pool.QueryRowContext(ctx,
		`SELECT role FROM flow_workspace_members WHERE workspace_id = ? AND user_id = ?`,
		workspaceID, userID,
	).Scan(&role)
	if err != nil {
		return "", err
	}
	return role, nil
}

// ─── Project Columns ─────────────────────────────────────────────────────────

func (r *Repository) GetColumnsByProject(ctx context.Context, projectID int64) ([]domain.ProjectColumn, error) {
	rows, err := r.pool.QueryContext(ctx,
		`SELECT id, project_id, workspace_id, name, color, sort_order, created_by, created_at, updated_at
		 FROM flow_project_columns WHERE project_id = ? ORDER BY sort_order ASC`,
		projectID,
	)
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
	res, err := r.pool.ExecContext(ctx,
		`INSERT INTO flow_project_columns (project_id, workspace_id, name, color, sort_order, created_by)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		input.ProjectID, input.WorkspaceID, input.Name, input.Color, input.SortOrder, input.CreatedBy,
	)
	if err != nil {
		logger.Ctx(ctx).Error("CreateColumn failed", logger.Error(err))
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) UpdateColumn(ctx context.Context, input domain.UpdateProjectColumnInput) error {
	_, err := r.pool.ExecContext(ctx,
		`UPDATE flow_project_columns SET name = ?, color = ?, sort_order = ?, updated_at = NOW()
		 WHERE id = ?`,
		input.Name, input.Color, input.SortOrder, input.ID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("UpdateColumn failed", logger.Error(err))
	}
	return err
}

func (r *Repository) DeleteColumn(ctx context.Context, id int64) error {
	// Null out section_id for tasks in this column before deleting
	if _, err := r.pool.ExecContext(ctx,
		`UPDATE flow_tasks SET section_id = NULL WHERE section_id = ?`, id,
	); err != nil {
		return err
	}
	_, err := r.pool.ExecContext(ctx,
		`DELETE FROM flow_project_columns WHERE id = ?`, id,
	)
	if err != nil {
		logger.Ctx(ctx).Error("DeleteColumn failed", logger.Error(err))
	}
	return err
}

func (r *Repository) MoveTaskToColumn(ctx context.Context, taskID, columnID int64) error {
	var sectionVal interface{}
	if columnID == 0 {
		sectionVal = nil
	} else {
		sectionVal = columnID
	}
	_, err := r.pool.ExecContext(ctx,
		`UPDATE flow_tasks SET section_id = ?, updated_at = NOW() WHERE id = ?`,
		sectionVal, taskID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("MoveTaskToColumn failed", logger.Error(err))
	}
	return err
}

func (r *Repository) SeedDefaultColumns(ctx context.Context, projectID, workspaceID, createdBy int64) error {
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

// ─────────────────────────────────────────────
// Businesses
// ─────────────────────────────────────────────

func (r *Repository) ListBusinesses(ctx context.Context, ownerID int64) ([]domain.Business, error) {
	rows, err := r.pool.QueryContext(ctx,
		`SELECT id, owner_id, name, color, icon, created_at, updated_at
		 FROM flow_businesses WHERE owner_id = ? ORDER BY created_at ASC`,
		ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("ListBusinesses failed", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	var businesses []domain.Business
	for rows.Next() {
		var b domain.Business
		if err := rows.Scan(&b.ID, &b.OwnerID, &b.Name, &b.Color, &b.Icon, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, err
		}
		businesses = append(businesses, b)
	}
	return businesses, rows.Err()
}

func (r *Repository) CreateBusiness(ctx context.Context, input domain.BusinessInput) (int64, error) {
	color := input.Color
	if color == "" {
		color = "#64C5B9"
	}
	icon := input.Icon
	if icon == "" {
		icon = "🏢"
	}

	res, err := r.pool.ExecContext(ctx,
		`INSERT INTO flow_businesses (owner_id, name, color, icon) VALUES (?, ?, ?, ?)`,
		input.OwnerID, input.Name, color, icon,
	)
	if err != nil {
		logger.Ctx(ctx).Error("CreateBusiness failed", logger.Error(err))
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) UpdateBusiness(ctx context.Context, id int64, name, color, icon string) error {
	_, err := r.pool.ExecContext(ctx,
		`UPDATE flow_businesses SET name = ?, color = ?, icon = ?, updated_at = NOW() WHERE id = ?`,
		name, color, icon, id,
	)
	if err != nil {
		logger.Ctx(ctx).Error("UpdateBusiness failed", logger.Error(err))
	}
	return err
}

func (r *Repository) DeleteBusiness(ctx context.Context, id int64, ownerID int64) error {
	_, err := r.pool.ExecContext(ctx,
		`DELETE FROM flow_businesses WHERE id = ? AND owner_id = ?`,
		id, ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("DeleteBusiness failed", logger.Error(err))
	}
	return err
}

func (r *Repository) SetWorkspaceBusiness(ctx context.Context, workspaceID int64, businessID *int64) error {
	_, err := r.pool.ExecContext(ctx,
		`UPDATE flow_workspaces SET business_id = ? WHERE id = ?`,
		businessID, workspaceID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("SetWorkspaceBusiness failed", logger.Error(err))
	}
	return err
}
