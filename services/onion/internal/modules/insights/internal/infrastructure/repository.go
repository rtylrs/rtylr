package infrastructure

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{pool: dbclient.Pool()}
}

type InsightsData struct {
	TotalTasks      int64
	CompletedTasks  int64
	InProgressTasks int64
	OverdueTasks    int64
	AvgCycleTimeDays float64
	TasksThisWeek   int64
	TasksLastWeek   int64
}

func (r *Repository) GetInsights(ctx context.Context, ownerID int64, workspaceID *int64) (InsightsData, error) {
	var data InsightsData

	query := `SELECT
		COUNT(*) as total_tasks,
		COUNT(CASE WHEN status = 'done' THEN 1 END) as completed_tasks,
		COUNT(CASE WHEN status = 'in_progress' THEN 1 END) as in_progress_tasks,
		COUNT(CASE WHEN status NOT IN ('done','cancelled') AND due_date < NOW() THEN 1 END) as overdue_tasks,
		COALESCE(AVG(CASE WHEN status = 'done' AND completed_at IS NOT NULL THEN TIMESTAMPDIFF(HOUR, created_at, completed_at) / 24.0 END), 0) as avg_cycle_time_days
	FROM flow_tasks WHERE owner_id = ?`
	args := []interface{}{ownerID}

	if workspaceID != nil {
		query += " AND workspace_id = ?"
		args = append(args, *workspaceID)
	}

	err := r.pool.QueryRowContext(ctx, query, args...).Scan(
		&data.TotalTasks, &data.CompletedTasks, &data.InProgressTasks,
		&data.OverdueTasks, &data.AvgCycleTimeDays,
	)
	if err != nil {
		logger.Ctx(ctx).Error("GetInsights failed", logger.Error(err))
		return InsightsData{}, err
	}

	weekStart := time.Now().AddDate(0, 0, -int(time.Now().Weekday()))
	lastWeekStart := weekStart.AddDate(0, 0, -7)

	throughputQuery := `SELECT
		COUNT(CASE WHEN completed_at >= ? THEN 1 END),
		COUNT(CASE WHEN completed_at >= ? AND completed_at < ? THEN 1 END)
	FROM flow_tasks WHERE owner_id = ? AND status = 'done'`
	throughputArgs := []interface{}{weekStart, lastWeekStart, weekStart, ownerID}

	if workspaceID != nil {
		throughputQuery += " AND workspace_id = ?"
		throughputArgs = append(throughputArgs, *workspaceID)
	}

	err = r.pool.QueryRowContext(ctx, throughputQuery, throughputArgs...).Scan(
		&data.TasksThisWeek, &data.TasksLastWeek,
	)
	if err != nil {
		logger.Ctx(ctx).Error("GetInsights throughput failed", logger.Error(err))
	}

	return data, nil
}

type WorkloadEntry struct {
	UserID          int64
	OpenTasks       int64
	InProgressTasks int64
	CompletedTasks  int64
	OverdueTasks    int64
}

func (r *Repository) GetTeamWorkload(ctx context.Context, ownerID int64, workspaceID *int64) ([]WorkloadEntry, error) {
	query := `SELECT
		assignee_id,
		COUNT(CASE WHEN status = 'todo' THEN 1 END) as open_tasks,
		COUNT(CASE WHEN status = 'in_progress' THEN 1 END) as in_progress_tasks,
		COUNT(CASE WHEN status = 'done' THEN 1 END) as completed_tasks,
		COUNT(CASE WHEN status NOT IN ('done','cancelled') AND due_date < NOW() THEN 1 END) as overdue_tasks
	FROM flow_tasks
	WHERE owner_id = ? AND assignee_id IS NOT NULL`
	args := []interface{}{ownerID}

	if workspaceID != nil {
		query += " AND workspace_id = ?"
		args = append(args, *workspaceID)
	}

	query += " GROUP BY assignee_id ORDER BY (open_tasks + in_progress_tasks) DESC"

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Ctx(ctx).Error("GetTeamWorkload failed", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	var entries []WorkloadEntry
	for rows.Next() {
		var e WorkloadEntry
		if scanErr := rows.Scan(&e.UserID, &e.OpenTasks, &e.InProgressTasks, &e.CompletedTasks, &e.OverdueTasks); scanErr != nil {
			logger.Ctx(ctx).Error("GetTeamWorkload scan failed", logger.Error(scanErr))
			return nil, scanErr
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

type RoadmapProject struct {
	ProjectID      int64
	ProjectName    string
	TotalTasks     int64
	CompletedTasks int64
	StartDate      *time.Time
	DueDate        *time.Time
	Status         string
}

func (r *Repository) GetRoadmap(ctx context.Context, ownerID int64, workspaceID *int64) ([]RoadmapProject, error) {
	query := `SELECT
		p.id, p.name, p.status, p.start_date, p.due_date,
		COUNT(t.id) as total_tasks,
		COUNT(CASE WHEN t.status = 'done' THEN 1 END) as completed_tasks
	FROM flow_projects p
	LEFT JOIN flow_tasks t ON t.project_id = p.id
	WHERE p.owner_id = ? AND p.status != 'archived'`
	args := []interface{}{ownerID}

	if workspaceID != nil {
		query += " AND p.workspace_id = ?"
		args = append(args, *workspaceID)
	}

	query += " GROUP BY p.id, p.name, p.status, p.start_date, p.due_date ORDER BY p.start_date ASC, p.created_at ASC"

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Ctx(ctx).Error("GetRoadmap failed", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	var projects []RoadmapProject
	for rows.Next() {
		var rp RoadmapProject
		var startDate, dueDate sql.NullTime
		if scanErr := rows.Scan(&rp.ProjectID, &rp.ProjectName, &rp.Status, &startDate, &dueDate, &rp.TotalTasks, &rp.CompletedTasks); scanErr != nil {
			logger.Ctx(ctx).Error("GetRoadmap scan failed", logger.Error(scanErr))
			return nil, scanErr
		}
		if startDate.Valid {
			rp.StartDate = &startDate.Time
		}
		if dueDate.Valid {
			rp.DueDate = &dueDate.Time
		}
		projects = append(projects, rp)
	}
	return projects, rows.Err()
}
