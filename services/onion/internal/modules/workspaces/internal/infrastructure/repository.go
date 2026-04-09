package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{
		pool: dbclient.Pool(),
	}
}

// ─────────────────────────────────────────────
// Workspaces
// ─────────────────────────────────────────────

func (r *Repository) CountWorkspacesByOwner(ctx context.Context, ownerID int64) (int64, error) {
	var count int64
	err := r.pool.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM flow_workspaces WHERE owner_id = ?`, ownerID,
	).Scan(&count)
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

	res, err := r.pool.ExecContext(ctx,
		`INSERT INTO flow_workspaces (owner_id, name, description, icon, color, is_default, created_by)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		input.OwnerID, input.Name, input.Description, icon, color, isDefault, input.CreatedBy,
	)
	if err != nil {
		logger.Ctx(ctx).Error("CreateWorkspace failed", logger.Error(err))
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *Repository) GetWorkspacesByOwner(ctx context.Context, ownerID int64) ([]domain.Workspace, error) {
	rows, err := r.pool.QueryContext(ctx,
		`SELECT id, owner_id, name, description, icon, color, is_default, created_by, created_at, updated_at
		 FROM flow_workspaces WHERE owner_id = ? ORDER BY is_default DESC, created_at ASC`,
		ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("GetWorkspacesByOwner failed", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	var workspaces []domain.Workspace
	for rows.Next() {
		var w domain.Workspace
		var desc sql.NullString

		if scanErr := rows.Scan(
			&w.ID, &w.OwnerID, &w.Name, &desc, &w.Icon, &w.Color, &w.IsDefault, &w.CreatedBy, &w.CreatedAt, &w.UpdatedAt,
		); scanErr != nil {
			logger.Ctx(ctx).Error("GetWorkspacesByOwner scan failed", logger.Error(scanErr))
			return nil, scanErr
		}

		w.Description = desc.String
		workspaces = append(workspaces, w)
	}

	return workspaces, rows.Err()
}

func (r *Repository) UpdateWorkspace(ctx context.Context, input domain.UpdateWorkspaceInput) error {
	_, err := r.pool.ExecContext(ctx,
		`UPDATE flow_workspaces SET name = ?, description = ?, icon = ?, color = ?
		 WHERE id = ? AND owner_id = ?`,
		input.Name, input.Description, input.Icon, input.Color, input.ID, input.OwnerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("UpdateWorkspace failed", logger.Error(err))
		return err
	}

	return nil
}

func (r *Repository) DeleteWorkspace(ctx context.Context, id, ownerID int64) error {
	_, err := r.pool.ExecContext(ctx,
		`DELETE FROM flow_workspaces WHERE id = ? AND owner_id = ?`, id, ownerID,
	)
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

	_, err := r.pool.ExecContext(ctx,
		`INSERT INTO flow_workspace_members (workspace_id, user_id, role) VALUES (?, ?, ?)
		 ON DUPLICATE KEY UPDATE role = VALUES(role)`,
		input.WorkspaceID, input.UserID, role,
	)
	if err != nil {
		logger.Ctx(ctx).Error("AddWorkspaceMember failed", logger.Error(err))
		return err
	}

	return nil
}

func (r *Repository) RemoveWorkspaceMember(ctx context.Context, input domain.RemoveWorkspaceMemberInput) error {
	_, err := r.pool.ExecContext(ctx,
		`DELETE FROM flow_workspace_members WHERE workspace_id = ? AND user_id = ?`,
		input.WorkspaceID, input.UserID,
	)
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

	res, err := r.pool.ExecContext(ctx,
		`INSERT INTO flow_projects (workspace_id, owner_id, name, description, icon, color, start_date, due_date, created_by)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		input.WorkspaceID, input.OwnerID, input.Name, input.Description, icon, color,
		nullTime(input.StartDate), nullTime(input.DueDate), input.CreatedBy,
	)
	if err != nil {
		logger.Ctx(ctx).Error("CreateProject failed", logger.Error(err))
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *Repository) GetProjectsByWorkspace(ctx context.Context, workspaceID, ownerID int64) ([]domain.Project, error) {
	rows, err := r.pool.QueryContext(ctx,
		`SELECT id, workspace_id, owner_id, name, description, icon, color, status, start_date, due_date, created_by, created_at, updated_at
		 FROM flow_projects WHERE workspace_id = ? AND owner_id = ? AND status != 'archived' ORDER BY created_at ASC`,
		workspaceID, ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("GetProjectsByWorkspace failed", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	var projects []domain.Project
	for rows.Next() {
		var p domain.Project
		var desc sql.NullString
		var startDate, dueDate sql.NullTime

		if scanErr := rows.Scan(
			&p.ID, &p.WorkspaceID, &p.OwnerID, &p.Name, &desc, &p.Icon, &p.Color, &p.Status,
			&startDate, &dueDate, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt,
		); scanErr != nil {
			logger.Ctx(ctx).Error("GetProjectsByWorkspace scan failed", logger.Error(scanErr))
			return nil, scanErr
		}

		p.Description = desc.String

		if startDate.Valid {
			p.StartDate = &startDate.Time
		}

		if dueDate.Valid {
			p.DueDate = &dueDate.Time
		}

		projects = append(projects, p)
	}

	return projects, rows.Err()
}

func (r *Repository) UpdateProject(ctx context.Context, input domain.UpdateProjectInput) error {
	_, err := r.pool.ExecContext(ctx,
		`UPDATE flow_projects SET name = ?, description = ?, icon = ?, color = ?, status = ?, start_date = ?, due_date = ?
		 WHERE id = ? AND owner_id = ?`,
		input.Name, input.Description, input.Icon, input.Color, input.Status,
		nullTime(input.StartDate), nullTime(input.DueDate),
		input.ID, input.OwnerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("UpdateProject failed", logger.Error(err))
		return err
	}

	return nil
}

func (r *Repository) ArchiveProject(ctx context.Context, id, ownerID int64) error {
	_, err := r.pool.ExecContext(ctx,
		`UPDATE flow_projects SET status = 'archived' WHERE id = ? AND owner_id = ?`, id, ownerID,
	)
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

	res, err := r.pool.ExecContext(ctx,
		`INSERT INTO flow_tasks
		 (project_id, workspace_id, owner_id, parent_task_id, title, description, priority,
		  assignee_id, reporter_id, due_date, start_date, estimated_hours, labels)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		input.ProjectID, input.WorkspaceID, input.OwnerID, nullInt64(input.ParentTaskID),
		input.Title, input.Description, priority,
		nullInt64(input.AssigneeID), input.ReporterID,
		nullTime(input.DueDate), nullTime(input.StartDate),
		nullFloat64(input.EstimatedHours), labelsJSON,
	)
	if err != nil {
		logger.Ctx(ctx).Error("CreateTask failed", logger.Error(err))
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *Repository) GetTasksByProject(ctx context.Context, filter domain.ListTasksFilter) ([]domain.Task, error) {
	query := `SELECT id, project_id, workspace_id, owner_id, parent_task_id, title, description,
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
		logger.Ctx(ctx).Error("GetTasksByProject failed", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	return scanTasks(ctx, rows)
}

func (r *Repository) GetTaskByID(ctx context.Context, taskID int64) (domain.Task, error) {
	row := r.pool.QueryRowContext(ctx,
		`SELECT id, project_id, workspace_id, owner_id, parent_task_id, title, description,
		  status, priority, assignee_id, reporter_id, due_date, start_date, completed_at,
		  estimated_hours, actual_hours, sort_order, labels, created_at, updated_at
		 FROM flow_tasks WHERE id = ?`, taskID,
	)

	return scanTask(row)
}

func (r *Repository) GetSubtasks(ctx context.Context, parentTaskID int64) ([]domain.Task, error) {
	rows, err := r.pool.QueryContext(ctx,
		`SELECT id, project_id, workspace_id, owner_id, parent_task_id, title, description,
		  status, priority, assignee_id, reporter_id, due_date, start_date, completed_at,
		  estimated_hours, actual_hours, sort_order, labels, created_at, updated_at
		 FROM flow_tasks WHERE parent_task_id = ? ORDER BY sort_order ASC, created_at ASC`,
		parentTaskID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("GetSubtasks failed", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	return scanTasks(ctx, rows)
}

func (r *Repository) GetMyTasks(ctx context.Context, userID int64) ([]domain.Task, error) {
	rows, err := r.pool.QueryContext(ctx,
		`SELECT id, project_id, workspace_id, owner_id, parent_task_id, title, description,
		  status, priority, assignee_id, reporter_id, due_date, start_date, completed_at,
		  estimated_hours, actual_hours, sort_order, labels, created_at, updated_at
		 FROM flow_tasks
		 WHERE assignee_id = ? AND status NOT IN ('done','cancelled')
		 ORDER BY
		   FIELD(priority,'urgent','high','medium','low','none'),
		   due_date ASC`,
		userID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("GetMyTasks failed", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	return scanTasks(ctx, rows)
}

func (r *Repository) UpdateTask(ctx context.Context, input domain.UpdateTaskInput) error {
	labelsJSON, err := marshalLabels(input.Labels)
	if err != nil {
		return err
	}

	completedAt := "NULL"
	_ = completedAt

	query := `UPDATE flow_tasks
	          SET title = ?, description = ?, status = ?, priority = ?, assignee_id = ?,
	              due_date = ?, start_date = ?, labels = ?,
	              completed_at = CASE WHEN ? = 'done' AND completed_at IS NULL THEN NOW() WHEN ? != 'done' THEN NULL ELSE completed_at END
	          WHERE id = ? AND owner_id = ?`

	_, err = r.pool.ExecContext(ctx, query,
		input.Title, input.Description, input.Status, input.Priority,
		nullInt64(input.AssigneeID),
		nullTime(input.DueDate), nullTime(input.StartDate),
		labelsJSON,
		input.Status, input.Status,
		input.ID, input.OwnerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("UpdateTask failed", logger.Error(err))
		return err
	}

	return nil
}

func (r *Repository) UpdateTaskStatus(ctx context.Context, input domain.UpdateTaskStatusInput) error {
	_, err := r.pool.ExecContext(ctx,
		`UPDATE flow_tasks
		 SET status = ?,
		     completed_at = CASE WHEN ? = 'done' AND completed_at IS NULL THEN NOW() WHEN ? != 'done' THEN NULL ELSE completed_at END
		 WHERE id = ? AND owner_id = ?`,
		input.Status, input.Status, input.Status, input.ID, input.OwnerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("UpdateTaskStatus failed", logger.Error(err))
		return err
	}

	return nil
}

func (r *Repository) DeleteTask(ctx context.Context, taskID, ownerID int64) error {
	_, err := r.pool.ExecContext(ctx,
		`DELETE FROM flow_tasks WHERE id = ? AND owner_id = ?`, taskID, ownerID,
	)
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
	res, err := r.pool.ExecContext(ctx,
		`INSERT INTO flow_task_comments (task_id, user_id, content) VALUES (?, ?, ?)`,
		taskID, userID, content,
	)
	if err != nil {
		logger.Ctx(ctx).Error("CreateComment failed", logger.Error(err))
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *Repository) GetCommentsByTask(ctx context.Context, taskID int64) ([]domain.TaskComment, error) {
	rows, err := r.pool.QueryContext(ctx,
		`SELECT id, task_id, user_id, content, is_edited, created_at, updated_at
		 FROM flow_task_comments WHERE task_id = ? ORDER BY created_at ASC`,
		taskID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("GetCommentsByTask failed", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	var comments []domain.TaskComment
	for rows.Next() {
		var c domain.TaskComment
		if scanErr := rows.Scan(&c.ID, &c.TaskID, &c.UserID, &c.Content, &c.IsEdited, &c.CreatedAt, &c.UpdatedAt); scanErr != nil {
			logger.Ctx(ctx).Error("GetCommentsByTask scan failed", logger.Error(scanErr))
			return nil, scanErr
		}

		comments = append(comments, c)
	}

	return comments, rows.Err()
}

func (r *Repository) DeleteComment(ctx context.Context, commentID, userID int64) error {
	_, err := r.pool.ExecContext(ctx,
		`DELETE FROM flow_task_comments WHERE id = ? AND user_id = ?`, commentID, userID,
	)
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
	_, err := r.pool.ExecContext(ctx,
		`INSERT INTO flow_task_activity (task_id, user_id, action, old_value, new_value) VALUES (?, ?, ?, ?, ?)`,
		taskID, userID, action, oldValue, newValue,
	)
	if err != nil {
		logger.Ctx(ctx).Error("LogActivity failed", logger.Error(err))
		return err
	}

	return nil
}

func (r *Repository) GetActivityByTask(ctx context.Context, taskID int64, limit int) ([]domain.TaskActivity, error) {
	rows, err := r.pool.QueryContext(ctx,
		`SELECT id, task_id, user_id, action, COALESCE(old_value,''), COALESCE(new_value,''), created_at
		 FROM flow_task_activity WHERE task_id = ? ORDER BY created_at DESC LIMIT ?`,
		taskID, limit,
	)
	if err != nil {
		logger.Ctx(ctx).Error("GetActivityByTask failed", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	var activities []domain.TaskActivity
	for rows.Next() {
		var a domain.TaskActivity
		if scanErr := rows.Scan(&a.ID, &a.TaskID, &a.UserID, &a.Action, &a.OldValue, &a.NewValue, &a.CreatedAt); scanErr != nil {
			logger.Ctx(ctx).Error("GetActivityByTask scan failed", logger.Error(scanErr))
			return nil, scanErr
		}

		activities = append(activities, a)
	}

	return activities, rows.Err()
}

// ─────────────────────────────────────────────
// Labels
// ─────────────────────────────────────────────

func (r *Repository) CreateLabel(ctx context.Context, input domain.LabelInput) (int64, error) {
	color := input.Color
	if color == "" {
		color = "#6366f1"
	}

	res, err := r.pool.ExecContext(ctx,
		`INSERT INTO flow_labels (workspace_id, name, color) VALUES (?, ?, ?)`,
		input.WorkspaceID, input.Name, color,
	)
	if err != nil {
		logger.Ctx(ctx).Error("CreateLabel failed", logger.Error(err))
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *Repository) GetLabelsByWorkspace(ctx context.Context, workspaceID int64) ([]domain.Label, error) {
	rows, err := r.pool.QueryContext(ctx,
		`SELECT id, workspace_id, name, color, created_at
		 FROM flow_labels WHERE workspace_id = ? ORDER BY name ASC`,
		workspaceID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("GetLabelsByWorkspace failed", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	var labels []domain.Label
	for rows.Next() {
		var l domain.Label
		if scanErr := rows.Scan(&l.ID, &l.WorkspaceID, &l.Name, &l.Color, &l.CreatedAt); scanErr != nil {
			logger.Ctx(ctx).Error("GetLabelsByWorkspace scan failed", logger.Error(scanErr))
			return nil, scanErr
		}

		labels = append(labels, l)
	}

	return labels, rows.Err()
}

// ─────────────────────────────────────────────
// Dashboard
// ─────────────────────────────────────────────

func (r *Repository) GetDashboardStats(ctx context.Context, workspaceID, ownerID int64) (domain.DashboardStats, error) {
	var stats domain.DashboardStats

	weekStart := time.Now().AddDate(0, 0, -int(time.Now().Weekday()))
	weekEnd := weekStart.AddDate(0, 0, 7)

	err := r.pool.QueryRowContext(ctx,
		`SELECT
		   COUNT(CASE WHEN status = 'todo' THEN 1 END),
		   COUNT(CASE WHEN status = 'in_progress' THEN 1 END),
		   COUNT(CASE WHEN status = 'done' AND completed_at >= ? AND completed_at < ? THEN 1 END),
		   COUNT(CASE WHEN status NOT IN ('done','cancelled') AND due_date < NOW() THEN 1 END)
		 FROM flow_tasks WHERE workspace_id = ? AND owner_id = ?`,
		weekStart, weekEnd, workspaceID, ownerID,
	).Scan(&stats.OpenTasks, &stats.InProgressTasks, &stats.CompletedThisWeek, &stats.OverdueTasks)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Ctx(ctx).Error("GetDashboardStats failed", logger.Error(err))
		return domain.DashboardStats{}, err
	}

	return stats, nil
}

// ─────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────

func nullInt64(v *int64) interface{} {
	if v == nil {
		return nil
	}

	return *v
}

func nullFloat64(v *float64) interface{} {
	if v == nil {
		return nil
	}

	return *v
}

func nullTime(v *time.Time) interface{} {
	if v == nil {
		return nil
	}

	return *v
}

func marshalLabels(labels []int64) ([]byte, error) {
	if labels == nil {
		return []byte("[]"), nil
	}

	return json.Marshal(labels)
}

type taskScanner interface {
	Scan(dest ...any) error
}

func scanTask(row taskScanner) (domain.Task, error) {
	var t domain.Task
	var desc sql.NullString
	var parentID sql.NullInt64
	var assigneeID sql.NullInt64
	var dueDate, startDate, completedAt sql.NullTime
	var estHours, actHours sql.NullFloat64
	var labelsRaw []byte

	err := row.Scan(
		&t.ID, &t.ProjectID, &t.WorkspaceID, &t.OwnerID, &parentID,
		&t.Title, &desc, &t.Status, &t.Priority,
		&assigneeID, &t.ReporterID,
		&dueDate, &startDate, &completedAt,
		&estHours, &actHours, &t.SortOrder, &labelsRaw,
		&t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return domain.Task{}, err
	}

	t.Description = desc.String

	if parentID.Valid {
		t.ParentTaskID = &parentID.Int64
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

	return t, nil
}

func scanTasks(ctx context.Context, rows *sql.Rows) ([]domain.Task, error) {
	var tasks []domain.Task
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			logger.Ctx(ctx).Error("scanTasks row scan failed", logger.Error(err))
			return nil, err
		}

		tasks = append(tasks, t)
	}

	return tasks, rows.Err()
}
