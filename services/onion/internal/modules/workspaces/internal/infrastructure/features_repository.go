package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

// ─────────────────────────────────────────────
// Cycles
// ─────────────────────────────────────────────

func (r *Repository) ListCycles(ctx context.Context, workspaceID int64) ([]domain.Cycle, error) {
	q := `
		SELECT c.id, c.workspace_id, c.project_id, c.name, c.description, c.status,
		       c.start_date, c.end_date, c.created_by, c.created_at,
		       COUNT(ct.task_id) as scope,
		       COUNT(CASE WHEN t.status = 'done' THEN 1 END) as completed
		FROM flow_cycles c
		LEFT JOIN flow_cycle_tasks ct ON ct.cycle_id = c.id
		LEFT JOIN flow_tasks t ON t.id = ct.task_id
		WHERE c.workspace_id = ?
		GROUP BY c.id
		ORDER BY c.start_date DESC`

	rows, err := r.pool.QueryContext(ctx, q, workspaceID)
	if err != nil {
		logger.Ctx(ctx).Error("ListCycles failed", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	var cycles []domain.Cycle
	for rows.Next() {
		var c domain.Cycle
		var projID sql.NullInt64
		var desc sql.NullString
		if err := rows.Scan(&c.ID, &c.WorkspaceID, &projID, &c.Name, &desc, &c.Status,
			&c.StartDate, &c.EndDate, &c.CreatedBy, &c.CreatedAt, &c.Scope, &c.Completed); err != nil {
			return nil, err
		}
		c.Description = desc.String
		if projID.Valid {
			c.ProjectID = &projID.Int64
		}
		// Auto-update status based on dates
		now := time.Now()
		if c.Status != "completed" {
			if now.Before(c.StartDate) {
				c.Status = "upcoming"
			} else if now.After(c.EndDate) {
				c.Status = "completed"
			} else {
				c.Status = "active"
			}
		}
		cycles = append(cycles, c)
	}
	return cycles, rows.Err()
}

func (r *Repository) CreateCycle(ctx context.Context, input domain.CycleInput) (int64, error) {
	res, err := r.pool.ExecContext(ctx,
		`INSERT INTO flow_cycles (workspace_id, project_id, name, description, start_date, end_date, created_by)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		input.WorkspaceID,
		nullInt64PtrFromPtr(input.ProjectID),
		input.Name,
		sql.NullString{String: input.Description, Valid: input.Description != ""},
		input.StartDate.Format("2006-01-02"),
		input.EndDate.Format("2006-01-02"),
		input.CreatedBy,
	)
	if err != nil {
		logger.Ctx(ctx).Error("CreateCycle failed", logger.Error(err))
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) UpdateCycle(ctx context.Context, input domain.UpdateCycleInput) error {
	parts := []string{"name = ?"}
	args := []interface{}{input.Name}

	if input.Status != "" {
		parts = append(parts, "status = ?")
		args = append(args, input.Status)
	}
	if !input.EndDate.IsZero() {
		parts = append(parts, "end_date = ?")
		args = append(args, input.EndDate.Format("2006-01-02"))
	}
	args = append(args, input.ID, input.ActorID, input.ActorID)

	res, err := r.pool.ExecContext(ctx,
		fmt.Sprintf("UPDATE flow_cycles SET %s WHERE id = ? AND %s", strings.Join(parts, ", "), workspaceMemberPredicate("flow_cycles")),
		args...,
	)
	if err != nil {
		logger.Ctx(ctx).Error("UpdateCycle failed", logger.Error(err))
		return err
	}
	return requireAffected(res)
}

func nullInt64PtrFromPtr(v *int64) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *v, Valid: true}
}

// ─────────────────────────────────────────────
// Milestones
// ─────────────────────────────────────────────

func (r *Repository) ListMilestones(ctx context.Context, workspaceID int64, projectID *int64) ([]domain.Milestone, error) {
	q := `
		SELECT m.id, m.project_id, m.workspace_id, m.name, m.description, m.status,
		       m.due_date, m.completed_at, m.created_by, m.created_at,
		       p.name as project_name, p.color as project_color,
		       COUNT(mt.task_id) as task_count,
		       COUNT(CASE WHEN t.status = 'done' THEN 1 END) as done_count
		FROM flow_milestones m
		LEFT JOIN flow_projects p ON p.id = m.project_id
		LEFT JOIN flow_milestone_tasks mt ON mt.milestone_id = m.id
		LEFT JOIN flow_tasks t ON t.id = mt.task_id
		WHERE m.workspace_id = ?`

	args := []interface{}{workspaceID}
	if projectID != nil {
		q += " AND m.project_id = ?"
		args = append(args, *projectID)
	}
	q += " GROUP BY m.id ORDER BY m.due_date ASC"

	rows, err := r.pool.QueryContext(ctx, q, args...)
	if err != nil {
		logger.Ctx(ctx).Error("ListMilestones failed", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	var milestones []domain.Milestone
	now := time.Now()
	for rows.Next() {
		var m domain.Milestone
		var desc sql.NullString
		var completedAt sql.NullTime
		var projName, projColor sql.NullString
		if err := rows.Scan(&m.ID, &m.ProjectID, &m.WorkspaceID, &m.Name, &desc, &m.Status,
			&m.DueDate, &completedAt, &m.CreatedBy, &m.CreatedAt,
			&projName, &projColor, &m.TaskCount, &m.DoneCount); err != nil {
			return nil, err
		}
		m.Description = desc.String
		if completedAt.Valid {
			m.CompletedAt = &completedAt.Time
		}
		if projName.Valid {
			m.Project = &domain.MilestoneProject{Name: projName.String, Color: projColor.String}
		}
		if m.TaskCount > 0 {
			m.Progress = int(float64(m.DoneCount) / float64(m.TaskCount) * 100)
		}
		// Auto-derive status
		if m.Status != "completed" && m.DueDate.Before(now) {
			m.Status = "at_risk"
		}
		milestones = append(milestones, m)
	}
	return milestones, rows.Err()
}

func (r *Repository) CreateMilestone(ctx context.Context, input domain.MilestoneInput) (int64, error) {
	res, err := r.pool.ExecContext(ctx,
		`INSERT INTO flow_milestones (project_id, workspace_id, name, description, due_date, created_by)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		input.ProjectID, input.WorkspaceID, input.Name,
		sql.NullString{String: input.Description, Valid: input.Description != ""},
		input.DueDate.Format("2006-01-02"), input.CreatedBy,
	)
	if err != nil {
		logger.Ctx(ctx).Error("CreateMilestone failed", logger.Error(err))
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) UpdateMilestone(ctx context.Context, input domain.UpdateMilestoneInput) error {
	parts := []string{"name = ?"}
	args := []interface{}{input.Name}

	if input.Description != "" {
		parts = append(parts, "description = ?")
		args = append(args, input.Description)
	}
	if input.Status != "" {
		parts = append(parts, "status = ?")
		args = append(args, input.Status)
		if input.Status == "completed" {
			parts = append(parts, "completed_at = NOW()")
		}
	}
	if !input.DueDate.IsZero() {
		parts = append(parts, "due_date = ?")
		args = append(args, input.DueDate.Format("2006-01-02"))
	}
	args = append(args, input.ID, input.ActorID, input.ActorID)

	res, err := r.pool.ExecContext(ctx,
		fmt.Sprintf("UPDATE flow_milestones SET %s WHERE id = ? AND %s", strings.Join(parts, ", "), workspaceMemberPredicate("flow_milestones")),
		args...,
	)
	if err != nil {
		logger.Ctx(ctx).Error("UpdateMilestone failed", logger.Error(err))
		return err
	}
	return requireAffected(res)
}

// ─────────────────────────────────────────────
// Views
// ─────────────────────────────────────────────

func (r *Repository) ListViews(ctx context.Context, workspaceID int64) ([]domain.View, error) {
	rows, err := r.pool.QueryContext(ctx,
		`SELECT id, workspace_id, name, icon, filters, group_by, sort_by, sort_dir, created_by, created_at
		 FROM flow_views WHERE workspace_id = ? ORDER BY created_at ASC`,
		workspaceID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("ListViews failed", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	var views []domain.View
	for rows.Next() {
		var v domain.View
		var icon, groupBy, sortBy, sortDir sql.NullString
		var filtersRaw []byte
		if err := rows.Scan(&v.ID, &v.WorkspaceID, &v.Name, &icon, &filtersRaw,
			&groupBy, &sortBy, &sortDir, &v.CreatedBy, &v.CreatedAt); err != nil {
			return nil, err
		}
		v.Icon = icon.String
		v.GroupBy = groupBy.String
		v.SortBy = sortBy.String
		v.SortDir = sortDir.String
		if len(filtersRaw) > 0 {
			_ = json.Unmarshal(filtersRaw, &v.Filters)
		}
		if v.Filters == nil {
			v.Filters = map[string]interface{}{}
		}
		views = append(views, v)
	}
	return views, rows.Err()
}

func (r *Repository) CreateView(ctx context.Context, input domain.ViewInput) (int64, error) {
	filtersJSON, err := json.Marshal(input.Filters)
	if err != nil {
		return 0, err
	}
	res, err := r.pool.ExecContext(ctx,
		`INSERT INTO flow_views (workspace_id, name, icon, filters, group_by, sort_by, sort_dir, created_by)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		input.WorkspaceID, input.Name,
		sql.NullString{String: input.Icon, Valid: input.Icon != ""},
		filtersJSON,
		sql.NullString{String: input.GroupBy, Valid: input.GroupBy != ""},
		sql.NullString{String: input.SortBy, Valid: input.SortBy != ""},
		sql.NullString{String: input.SortDir, Valid: input.SortDir != ""},
		input.CreatedBy,
	)
	if err != nil {
		logger.Ctx(ctx).Error("CreateView failed", logger.Error(err))
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) UpdateView(ctx context.Context, input domain.UpdateViewInput) error {
	filtersJSON, err := json.Marshal(input.Filters)
	if err != nil {
		return err
	}
	res, err := r.pool.ExecContext(ctx,
		fmt.Sprintf(`UPDATE flow_views SET name = ?, icon = ?, filters = ?, group_by = ?, sort_by = ?, sort_dir = ? WHERE id = ? AND %s`, workspaceMemberPredicate("flow_views")),
		input.Name,
		sql.NullString{String: input.Icon, Valid: input.Icon != ""},
		filtersJSON,
		sql.NullString{String: input.GroupBy, Valid: input.GroupBy != ""},
		sql.NullString{String: input.SortBy, Valid: input.SortBy != ""},
		sql.NullString{String: input.SortDir, Valid: input.SortDir != ""},
		input.ID,
		input.ActorID,
		input.ActorID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("UpdateView failed", logger.Error(err))
		return err
	}
	return requireAffected(res)
}

func (r *Repository) DeleteView(ctx context.Context, id, actorID int64) error {
	res, err := r.pool.ExecContext(ctx, fmt.Sprintf(`DELETE FROM flow_views WHERE id = ? AND %s`, workspaceMemberPredicate("flow_views")), id, actorID, actorID)
	if err != nil {
		logger.Ctx(ctx).Error("DeleteView failed", logger.Error(err))
		return err
	}
	return requireAffected(res)
}

func workspaceMemberPredicate(table string) string {
	return fmt.Sprintf(`(
		EXISTS (SELECT 1 FROM flow_workspaces w WHERE w.id = %s.workspace_id AND w.owner_id = ?)
		OR EXISTS (SELECT 1 FROM flow_workspace_members wm WHERE wm.workspace_id = %s.workspace_id AND wm.user_id = ?)
	)`, table, table)
}

func requireAffected(res sql.Result) error {
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrUnauthorized
	}
	return nil
}

// ─────────────────────────────────────────────
// Insights (computed aggregations)
// ─────────────────────────────────────────────

func (r *Repository) GetInsights(ctx context.Context, workspaceID int64) (domain.Insights, error) {
	var ins domain.Insights

	// Total / done tasks
	_ = r.pool.QueryRowContext(ctx,
		`SELECT COUNT(*), SUM(status = 'done') FROM flow_tasks WHERE workspace_id = ? AND parent_task_id IS NULL`,
		workspaceID,
	).Scan(&ins.TotalTasks, &ins.DoneTasks)

	// Cycle time: avg days from created_at to completed_at
	var avgDays sql.NullFloat64
	_ = r.pool.QueryRowContext(ctx,
		`SELECT AVG(DATEDIFF(completed_at, created_at))
		 FROM flow_tasks WHERE workspace_id = ? AND status = 'done' AND completed_at IS NOT NULL`,
		workspaceID,
	).Scan(&avgDays)
	if avgDays.Valid {
		ins.CycleTime = math.Round(avgDays.Float64*10) / 10
	}

	// Velocity: tasks completed per 2-week block, last 8 weeks
	velRows, err := r.pool.QueryContext(ctx, `
		SELECT
			DATE_FORMAT(DATE_SUB(completed_at, INTERVAL WEEKDAY(completed_at) DAY), '%Y-%m-%d') as week_start,
			COUNT(*) as completed
		FROM flow_tasks
		WHERE workspace_id = ?
		  AND status = 'done'
		  AND completed_at >= DATE_SUB(NOW(), INTERVAL 8 WEEK)
		GROUP BY YEARWEEK(completed_at, 1)
		ORDER BY week_start ASC`, workspaceID)
	if err == nil {
		defer velRows.Close()
		for velRows.Next() {
			var weekStart string
			var completed int
			if scanErr := velRows.Scan(&weekStart, &completed); scanErr == nil {
				ins.Velocity = append(ins.Velocity, domain.InsightsVelocityPoint{
					Cycle:          weekStart,
					Completed:      completed,
					Total:          completed + int(float64(completed)*0.3), // estimate scope
					CompletionRate: math.Min(1.0, float64(completed)/math.Max(1, float64(completed)+float64(completed)*0.3)),
				})
			}
		}
		if err := velRows.Err(); err != nil {
			logger.Ctx(ctx).Error("GetInsights velocity rows failed", logger.Error(err))
		}
	}

	// Throughput: completed per week last 8 weeks (same as velocity for now)
	ins.Throughput = make([]domain.InsightsThroughput, len(ins.Velocity))
	for i, v := range ins.Velocity {
		ins.Throughput[i] = domain.InsightsThroughput{Week: v.Cycle, Completed: v.Completed}
	}

	// Burndown: 14-day ideal vs actual remaining
	var totalOpen int
	_ = r.pool.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM flow_tasks WHERE workspace_id = ? AND status NOT IN ('done','cancelled') AND parent_task_id IS NULL`,
		workspaceID,
	).Scan(&totalOpen)
	idealPerDay := float64(totalOpen) / 14.0
	ins.Burndown = make([]domain.InsightsBurndownPoint, 15)
	for d := 0; d <= 14; d++ {
		remaining := math.Max(0, float64(totalOpen)-float64(d)*idealPerDay*1.15)
		ins.Burndown[d] = domain.InsightsBurndownPoint{
			Day:    d,
			Ideal:  math.Round((float64(totalOpen)-idealPerDay*float64(d))*10) / 10,
			Actual: math.Round(remaining*10) / 10,
		}
	}

	// Time by member (from time logs)
	tlRows, err := r.pool.QueryContext(ctx, `
		SELECT CAST(u.id AS CHAR), COALESCE(u.full_name, u.email), SUM(tl.duration_minutes)/60.0
		FROM flow_time_logs tl
		JOIN users u ON u.id = tl.user_id
		JOIN flow_tasks t ON t.id = tl.task_id
		WHERE t.workspace_id = ?
		GROUP BY u.id, u.full_name, u.email
		ORDER BY SUM(tl.duration_minutes) DESC
		LIMIT 10`, workspaceID)
	if err == nil {
		defer tlRows.Close()
		for tlRows.Next() {
			var m domain.InsightsMemberTime
			if scanErr := tlRows.Scan(&m.ID, &m.Name, &m.Hours); scanErr == nil {
				m.Hours = math.Round(m.Hours*10) / 10
				ins.TimeByMember = append(ins.TimeByMember, m)
			}
		}
		if err := tlRows.Err(); err != nil {
			logger.Ctx(ctx).Error("GetInsights time log rows failed", logger.Error(err))
		}
	}
	if ins.TimeByMember == nil {
		ins.TimeByMember = []domain.InsightsMemberTime{}
	}
	if ins.Velocity == nil {
		ins.Velocity = []domain.InsightsVelocityPoint{}
	}
	if ins.Throughput == nil {
		ins.Throughput = []domain.InsightsThroughput{}
	}

	return ins, nil
}

// ─────────────────────────────────────────────
// Team Workload
// ─────────────────────────────────────────────

func (r *Repository) GetTeamWorkload(ctx context.Context, workspaceID int64) ([]domain.TeamMemberWorkload, error) {
	rows, err := r.pool.QueryContext(ctx, `
		SELECT CAST(u.id AS CHAR), COALESCE(u.full_name, u.email), COUNT(t.id) as task_count
		FROM flow_workspace_members wm
		JOIN users u ON u.id = wm.user_id
		LEFT JOIN flow_tasks t ON t.assignee_id = u.id
			AND t.workspace_id = ?
			AND t.status NOT IN ('done', 'cancelled')
			AND t.parent_task_id IS NULL
		WHERE wm.workspace_id = ?
		GROUP BY u.id, u.full_name, u.email
		ORDER BY task_count DESC`, workspaceID, workspaceID)
	if err != nil {
		logger.Ctx(ctx).Error("GetTeamWorkload failed", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	var workloads []domain.TeamMemberWorkload
	for rows.Next() {
		var w domain.TeamMemberWorkload
		var taskCount int
		if err := rows.Scan(&w.ID, &w.Name, &taskCount); err != nil {
			return nil, err
		}
		// Initials from name
		parts := strings.Fields(w.Name)
		if len(parts) >= 2 {
			w.Initials = strings.ToUpper(string(parts[0][0])) + strings.ToUpper(string(parts[len(parts)-1][0]))
		} else if len(w.Name) > 0 {
			w.Initials = strings.ToUpper(string(w.Name[0]))
		}
		// Estimate: assume 8 tasks = full capacity (40h week, 5h/task)
		w.Capacity = 8
		w.AssignedHours = float64(taskCount) * 5
		w.Load = math.Min(1.0, float64(taskCount)/float64(w.Capacity))
		workloads = append(workloads, w)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Load tasks for each member
	for i := range workloads {
		taskRows, err := r.pool.QueryContext(ctx, `
			SELECT id, project_id, workspace_id, owner_id, parent_task_id,
			       title, description, status, priority,
			       assignee_id, reporter_id, due_date, start_date, completed_at,
			       estimated_hours, actual_hours, sort_order, labels,
			       created_at, updated_at
			FROM flow_tasks
			WHERE workspace_id = ? AND assignee_id = ? AND status NOT IN ('done','cancelled') AND parent_task_id IS NULL
			ORDER BY priority DESC, due_date ASC
			LIMIT 10`, workspaceID, workloads[i].ID)
		if err == nil {
			tasks, _ := scanTasksRaw(ctx, taskRows)
			taskRows.Close()
			workloads[i].Tasks = tasks
		}
		if workloads[i].Tasks == nil {
			workloads[i].Tasks = []domain.Task{}
		}
	}

	return workloads, nil
}

// ─────────────────────────────────────────────
// Roadmap (derived from projects + milestones)
// ─────────────────────────────────────────────

func (r *Repository) GetRoadmap(ctx context.Context, workspaceID int64) ([]domain.Project, error) {
	// Return active projects with progress data
	rows, err := r.pool.QueryContext(ctx, `
		SELECT p.id, p.workspace_id, p.owner_id, p.name, p.description, p.icon, p.color,
		       p.status, p.start_date, p.due_date, p.created_by, p.created_at, p.updated_at
		FROM flow_projects p
		WHERE p.workspace_id = ? AND p.status != 'archived'
		ORDER BY p.start_date ASC, p.created_at ASC`, workspaceID)
	if err != nil {
		logger.Ctx(ctx).Error("GetRoadmap failed", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	var projects []domain.Project
	for rows.Next() {
		var p domain.Project
		var desc, icon, color, status sql.NullString
		var startDate, dueDate sql.NullTime
		if err := rows.Scan(&p.ID, &p.WorkspaceID, &p.OwnerID, &p.Name, &desc, &icon, &color,
			&status, &startDate, &dueDate, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		p.Description = desc.String
		p.Icon = icon.String
		p.Color = color.String
		p.Status = status.String
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

// ─────────────────────────────────────────────
// Notifications
// ─────────────────────────────────────────────

func (r *Repository) ListNotifications(ctx context.Context, userID int64) ([]domain.Notification, error) {
	rows, err := r.pool.QueryContext(ctx, `
		SELECT id, user_id, workspace_id, type, title, COALESCE(description,''),
		       task_id, COALESCE(task_title,''), actor_id, COALESCE(actor_initials,''),
		       COALESCE(project,''), COALESCE(project_color,''), is_read, created_at
		FROM flow_notifications
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT 100`, userID)
	if err != nil {
		logger.Ctx(ctx).Error("ListNotifications failed", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	var notifs []domain.Notification
	for rows.Next() {
		var n domain.Notification
		var wsID, taskID, actorID sql.NullInt64
		if err := rows.Scan(&n.ID, &n.UserID, &wsID, &n.Type, &n.Title, &n.Description,
			&taskID, &n.TaskTitle, &actorID, &n.ActorInitials,
			&n.Project, &n.ProjectColor, &n.IsRead, &n.CreatedAt); err != nil {
			return nil, err
		}
		if wsID.Valid {
			n.WorkspaceID = &wsID.Int64
		}
		if taskID.Valid {
			n.TaskID = &taskID.Int64
		}
		if actorID.Valid {
			n.ActorID = &actorID.Int64
		}
		notifs = append(notifs, n)
	}
	return notifs, rows.Err()
}

func (r *Repository) CreateNotification(ctx context.Context, input domain.CreateNotificationInput) error {
	_, err := r.pool.ExecContext(ctx, `
		INSERT INTO flow_notifications
		  (user_id, workspace_id, type, title, description, task_id, task_title, actor_id, actor_initials, project, project_color)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		input.UserID,
		nullInt64PtrFromPtr(input.WorkspaceID),
		input.Type, input.Title,
		sql.NullString{String: input.Description, Valid: input.Description != ""},
		nullInt64PtrFromPtr(input.TaskID),
		sql.NullString{String: input.TaskTitle, Valid: input.TaskTitle != ""},
		nullInt64PtrFromPtr(input.ActorID),
		sql.NullString{String: input.ActorInitials, Valid: input.ActorInitials != ""},
		sql.NullString{String: input.Project, Valid: input.Project != ""},
		sql.NullString{String: input.ProjectColor, Valid: input.ProjectColor != ""},
	)
	if err != nil {
		logger.Ctx(ctx).Error("CreateNotification failed", logger.Error(err))
	}
	return err
}

func (r *Repository) MarkNotificationRead(ctx context.Context, id, userID int64) error {
	_, err := r.pool.ExecContext(ctx,
		`UPDATE flow_notifications SET is_read = TRUE WHERE id = ? AND user_id = ?`,
		id, userID,
	)
	return err
}

func (r *Repository) MarkAllNotificationsRead(ctx context.Context, userID int64) error {
	_, err := r.pool.ExecContext(ctx,
		`UPDATE flow_notifications SET is_read = TRUE WHERE user_id = ?`,
		userID,
	)
	return err
}

// ─────────────────────────────────────────────
// User Settings
// ─────────────────────────────────────────────

func (r *Repository) GetUserSettings(ctx context.Context, userID int64) (domain.UserSettings, error) {
	var s domain.UserSettings
	s.UserID = userID
	// Defaults
	s.TaskAssigned = true
	s.TaskComments = true
	s.Mentions = true
	s.WeeklyDigest = true

	err := r.pool.QueryRowContext(ctx, `
		SELECT task_assigned, task_comments, mentions, status_changes, weekly_digest, email_notifications
		FROM flow_user_settings WHERE user_id = ?`, userID,
	).Scan(&s.TaskAssigned, &s.TaskComments, &s.Mentions, &s.StatusChanges, &s.WeeklyDigest, &s.EmailNotifications)
	if errors.Is(err, sql.ErrNoRows) {
		return s, nil // return defaults
	}
	return s, err
}

func (r *Repository) UpsertUserSettings(ctx context.Context, s domain.UserSettings) error {
	_, err := r.pool.ExecContext(ctx, `
		INSERT INTO flow_user_settings (user_id, task_assigned, task_comments, mentions, status_changes, weekly_digest, email_notifications)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
		  task_assigned = VALUES(task_assigned),
		  task_comments = VALUES(task_comments),
		  mentions = VALUES(mentions),
		  status_changes = VALUES(status_changes),
		  weekly_digest = VALUES(weekly_digest),
		  email_notifications = VALUES(email_notifications)`,
		s.UserID, s.TaskAssigned, s.TaskComments, s.Mentions, s.StatusChanges, s.WeeklyDigest, s.EmailNotifications,
	)
	if err != nil {
		logger.Ctx(ctx).Error("UpsertUserSettings failed", logger.Error(err))
	}
	return err
}

// ─────────────────────────────────────────────
// Roadmap task/progress helpers
// ─────────────────────────────────────────────

func (r *Repository) GetProjectTaskCounts(ctx context.Context, projectID int64) (total, done int) {
	_ = r.pool.QueryRowContext(ctx,
		`SELECT COUNT(*), SUM(status='done') FROM flow_tasks WHERE project_id = ? AND parent_task_id IS NULL`,
		projectID,
	).Scan(&total, &done)
	return total, done
}
