package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/modules/flow_insights/internal/domain"
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
// Insights
// ─────────────────────────────────────────────

func (r *Repository) GetInsightsStats(ctx context.Context, ownerID int64) (domain.InsightsStats, error) {
	var stats domain.InsightsStats
	stats.TasksByStatus = make(map[string]int64)
	stats.TasksByPriority = make(map[string]int64)

	// Total, completed, overdue
	err := r.pool.QueryRowContext(ctx, `
		SELECT
			COUNT(*) AS total_tasks,
			COUNT(CASE WHEN status = 'done' THEN 1 END) AS completed_tasks,
			COUNT(CASE WHEN status NOT IN ('done','cancelled') AND due_date < CURDATE() THEN 1 END) AS overdue_count
		FROM flow_tasks WHERE owner_id = ?
	`, ownerID).Scan(&stats.TotalTasks, &stats.CompletedTasks, &stats.OverdueCount)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Ctx(ctx).Error("GetInsightsStats: total query failed", logger.Error(err))
		return stats, err
	}

	if stats.TotalTasks > 0 {
		stats.CompletionRate = float64(stats.CompletedTasks) / float64(stats.TotalTasks) * 100
	}

	// Avg completion days
	err = r.pool.QueryRowContext(ctx, `
		SELECT COALESCE(AVG(DATEDIFF(completed_at, created_at)), 0) AS avg_days
		FROM flow_tasks WHERE owner_id = ? AND status = 'done' AND completed_at IS NOT NULL
	`, ownerID).Scan(&stats.AvgCompletionDays)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Ctx(ctx).Error("GetInsightsStats: avg completion failed", logger.Error(err))
	}

	// Velocity this week
	err = r.pool.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM flow_tasks
		WHERE owner_id = ? AND status = 'done'
		  AND completed_at >= DATE_SUB(CURDATE(), INTERVAL WEEKDAY(CURDATE()) DAY)
		  AND completed_at < DATE_ADD(DATE_SUB(CURDATE(), INTERVAL WEEKDAY(CURDATE()) DAY), INTERVAL 7 DAY)
	`, ownerID).Scan(&stats.VelocityThisWeek)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Ctx(ctx).Error("GetInsightsStats: velocity this week failed", logger.Error(err))
	}

	// Velocity last week
	err = r.pool.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM flow_tasks
		WHERE owner_id = ? AND status = 'done'
		  AND completed_at >= DATE_SUB(DATE_SUB(CURDATE(), INTERVAL WEEKDAY(CURDATE()) DAY), INTERVAL 7 DAY)
		  AND completed_at < DATE_SUB(CURDATE(), INTERVAL WEEKDAY(CURDATE()) DAY)
	`, ownerID).Scan(&stats.VelocityLastWeek)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Ctx(ctx).Error("GetInsightsStats: velocity last week failed", logger.Error(err))
	}

	// Tasks by status
	statusRows, err := r.pool.QueryContext(ctx, `
		SELECT status, COUNT(*) AS cnt FROM flow_tasks WHERE owner_id = ? GROUP BY status
	`, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("GetInsightsStats: tasks by status failed", logger.Error(err))
	} else {
		defer statusRows.Close()
		for statusRows.Next() {
			var s string
			var c int64
			if err := statusRows.Scan(&s, &c); err == nil {
				stats.TasksByStatus[s] = c
			}
		}
		if err := statusRows.Err(); err != nil {
			logger.Ctx(ctx).Error("GetInsightsStats: tasks by status rows failed", logger.Error(err))
		}
	}

	// Tasks by priority
	priorityRows, err := r.pool.QueryContext(ctx, `
		SELECT priority, COUNT(*) AS cnt FROM flow_tasks WHERE owner_id = ? GROUP BY priority
	`, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("GetInsightsStats: tasks by priority failed", logger.Error(err))
	} else {
		defer priorityRows.Close()
		for priorityRows.Next() {
			var s string
			var c int64
			if err := priorityRows.Scan(&s, &c); err == nil {
				stats.TasksByPriority[s] = c
			}
		}
		if err := priorityRows.Err(); err != nil {
			logger.Ctx(ctx).Error("GetInsightsStats: tasks by priority rows failed", logger.Error(err))
		}
	}

	return stats, nil
}

// ─────────────────────────────────────────────
// Team Workload
// ─────────────────────────────────────────────

func (r *Repository) GetTeamWorkload(ctx context.Context, ownerID int64) (domain.TeamWorkloadResult, error) {
	rows, err := r.pool.QueryContext(ctx, `
		SELECT
			ft.assignee_id AS user_id,
			COALESCE(u.full_name, CONCAT('User ', ft.assignee_id)) AS name,
			COUNT(*) AS total_tasks,
			COUNT(CASE WHEN ft.status = 'done' THEN 1 END) AS completed_tasks,
			COUNT(CASE WHEN ft.status = 'in_progress' THEN 1 END) AS in_progress_tasks,
			COUNT(CASE WHEN ft.status NOT IN ('done','cancelled') AND ft.due_date < CURDATE() THEN 1 END) AS overdue_tasks
		FROM flow_tasks ft
		LEFT JOIN users u ON ft.assignee_id = u.id
		WHERE ft.owner_id = ? AND ft.assignee_id IS NOT NULL
		GROUP BY ft.assignee_id, u.full_name
	`, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("GetTeamWorkload failed", logger.Error(err))
		return domain.TeamWorkloadResult{}, err
	}
	defer rows.Close()

	var result domain.TeamWorkloadResult
	for rows.Next() {
		var m domain.TeamMemberWorkload
		if err := rows.Scan(&m.UserID, &m.Name, &m.TotalTasks, &m.CompletedTasks, &m.InProgressTasks, &m.OverdueTasks); err != nil {
			logger.Ctx(ctx).Error("GetTeamWorkload scan failed", logger.Error(err))
			continue
		}
		result.Members = append(result.Members, m)
	}

	if result.Members == nil {
		result.Members = []domain.TeamMemberWorkload{}
	}

	return result, rows.Err()
}

// ─────────────────────────────────────────────
// Roadmap
// ─────────────────────────────────────────────

func (r *Repository) GetRoadmap(ctx context.Context, ownerID int64) (domain.RoadmapResult, error) {
	projRows, err := r.pool.QueryContext(ctx, `
		SELECT
			fp.id,
			fp.name,
			COALESCE(DATE_FORMAT(fp.start_date, '%Y-%m-%d'), '') AS start_date,
			COALESCE(DATE_FORMAT(fp.due_date, '%Y-%m-%d'), '') AS due_date,
			COUNT(ft.id) AS total_tasks,
			COUNT(CASE WHEN ft.status = 'done' THEN 1 END) AS done_tasks
		FROM flow_projects fp
		LEFT JOIN flow_tasks ft ON ft.project_id = fp.id AND ft.parent_task_id IS NULL
		WHERE fp.owner_id = ? AND fp.status != 'archived'
		GROUP BY fp.id, fp.name, fp.start_date, fp.due_date
		ORDER BY fp.created_at ASC
	`, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("GetRoadmap: projects query failed", logger.Error(err))
		return domain.RoadmapResult{}, err
	}
	defer projRows.Close()

	var result domain.RoadmapResult
	for projRows.Next() {
		var p domain.RoadmapProject
		var totalTasks, doneTasks int64
		if err := projRows.Scan(&p.ID, &p.Name, &p.StartDate, &p.DueDate, &totalTasks, &doneTasks); err != nil {
			logger.Ctx(ctx).Error("GetRoadmap: project scan failed", logger.Error(err))
			continue
		}
		if totalTasks > 0 {
			p.Progress = float64(doneTasks) / float64(totalTasks) * 100
		}
		p.Milestones = []domain.RoadmapMilestone{}
		result.Projects = append(result.Projects, p)
	}
	if err := projRows.Err(); err != nil {
		return result, err
	}

	// Fetch milestones per project
	for i := range result.Projects {
		ms, err := r.getMilestonesByProject(ctx, result.Projects[i].ID)
		if err != nil {
			logger.Ctx(ctx).Error("GetRoadmap: milestones query failed",
				logger.Error(err), logger.Any("project_id", result.Projects[i].ID))
			continue
		}
		result.Projects[i].Milestones = ms
	}

	if result.Projects == nil {
		result.Projects = []domain.RoadmapProject{}
	}

	return result, nil
}

func (r *Repository) getMilestonesByProject(ctx context.Context, projectID int64) ([]domain.RoadmapMilestone, error) {
	rows, err := r.pool.QueryContext(ctx, `
		SELECT
			fm.id,
			fm.name,
			COALESCE(DATE_FORMAT(fm.target_date, '%Y-%m-%d'), '') AS due_date,
			COUNT(ft.id) AS task_count,
			COUNT(CASE WHEN ft.status = 'done' THEN 1 END) AS done_count
		FROM flow_milestones fm
		LEFT JOIN flow_tasks ft ON ft.project_id = fm.project_id AND ft.parent_task_id IS NULL
		WHERE fm.project_id = ?
		GROUP BY fm.id, fm.name, fm.target_date
		ORDER BY fm.target_date ASC
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var milestones []domain.RoadmapMilestone
	for rows.Next() {
		var m domain.RoadmapMilestone
		if err := rows.Scan(&m.ID, &m.Name, &m.DueDate, &m.TaskCount, &m.DoneCount); err != nil {
			continue
		}
		milestones = append(milestones, m)
	}

	if milestones == nil {
		milestones = []domain.RoadmapMilestone{}
	}

	return milestones, rows.Err()
}
