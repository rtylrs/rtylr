package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/flow_insights/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		pool: db,
	}
}

// ─────────────────────────────────────────────
// Insights
// ─────────────────────────────────────────────

func (r *Repository) GetInsightsStats(ctx context.Context, businessID int64) (domain.InsightsStats, error) {
	var stats domain.InsightsStats
	stats.TasksByStatus = make(map[string]int64)
	stats.TasksByPriority = make(map[string]int64)

	// Total, completed, overdue
	totalQuery, totalArgs, err := sq.
		Select(
			"COUNT(*) AS total_tasks",
			"COUNT(CASE WHEN status = 'done' THEN 1 END) AS completed_tasks",
			"COUNT(CASE WHEN status NOT IN ('done','cancelled') AND due_date < CURDATE() THEN 1 END) AS overdue_count",
		).
		From("flow_tasks").
		Where(sq.Eq{"business_id": businessID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("GetInsightsStats: build total query failed", logger.Error(err))
		return stats, err
	}
	err = r.pool.QueryRowContext(ctx, totalQuery, totalArgs...).Scan(&stats.TotalTasks, &stats.CompletedTasks, &stats.OverdueCount)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Ctx(ctx).Error("GetInsightsStats: total query failed", logger.Error(err))
		return stats, err
	}

	if stats.TotalTasks > 0 {
		stats.CompletionRate = float64(stats.CompletedTasks) / float64(stats.TotalTasks) * 100
	}

	// Avg completion days
	avgQuery, avgArgs, err := sq.
		Select("COALESCE(AVG(DATEDIFF(completed_at, created_at)), 0) AS avg_days").
		From("flow_tasks").
		Where(sq.Eq{"business_id": businessID, "status": "done"}).
		Where(sq.Expr("completed_at IS NOT NULL")).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("GetInsightsStats: build avg completion query failed", logger.Error(err))
	} else {
		err = r.pool.QueryRowContext(ctx, avgQuery, avgArgs...).Scan(&stats.AvgCompletionDays)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			logger.Ctx(ctx).Error("GetInsightsStats: avg completion failed", logger.Error(err))
		}
	}

	// Velocity this week
	velThisQuery, velThisArgs, err := sq.
		Select("COUNT(*)").
		From("flow_tasks").
		Where(sq.Eq{"business_id": businessID, "status": "done"}).
		Where(sq.Expr("completed_at >= DATE_SUB(CURDATE(), INTERVAL WEEKDAY(CURDATE()) DAY)")).
		Where(sq.Expr("completed_at < DATE_ADD(DATE_SUB(CURDATE(), INTERVAL WEEKDAY(CURDATE()) DAY), INTERVAL 7 DAY)")).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("GetInsightsStats: build velocity this week query failed", logger.Error(err))
	} else {
		err = r.pool.QueryRowContext(ctx, velThisQuery, velThisArgs...).Scan(&stats.VelocityThisWeek)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			logger.Ctx(ctx).Error("GetInsightsStats: velocity this week failed", logger.Error(err))
		}
	}

	// Velocity last week
	velLastQuery, velLastArgs, err := sq.
		Select("COUNT(*)").
		From("flow_tasks").
		Where(sq.Eq{"business_id": businessID, "status": "done"}).
		Where(sq.Expr("completed_at >= DATE_SUB(DATE_SUB(CURDATE(), INTERVAL WEEKDAY(CURDATE()) DAY), INTERVAL 7 DAY)")).
		Where(sq.Expr("completed_at < DATE_SUB(CURDATE(), INTERVAL WEEKDAY(CURDATE()) DAY)")).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("GetInsightsStats: build velocity last week query failed", logger.Error(err))
	} else {
		err = r.pool.QueryRowContext(ctx, velLastQuery, velLastArgs...).Scan(&stats.VelocityLastWeek)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			logger.Ctx(ctx).Error("GetInsightsStats: velocity last week failed", logger.Error(err))
		}
	}

	// Tasks by status
	statusQuery, statusArgs, err := sq.
		Select("status", "COUNT(*) AS cnt").
		From("flow_tasks").
		Where(sq.Eq{"business_id": businessID}).
		GroupBy("status").
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("GetInsightsStats: build tasks by status query failed", logger.Error(err))
	} else {
		statusRows, err := r.pool.QueryContext(ctx, statusQuery, statusArgs...)
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
	}

	// Tasks by priority
	priorityQuery, priorityArgs, err := sq.
		Select("priority", "COUNT(*) AS cnt").
		From("flow_tasks").
		Where(sq.Eq{"business_id": businessID}).
		GroupBy("priority").
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("GetInsightsStats: build tasks by priority query failed", logger.Error(err))
	} else {
		priorityRows, err := r.pool.QueryContext(ctx, priorityQuery, priorityArgs...)
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
	}

	return stats, nil
}

// ─────────────────────────────────────────────
// Team Workload
// ─────────────────────────────────────────────

func (r *Repository) GetTeamWorkload(ctx context.Context, businessID int64) (domain.TeamWorkloadResult, error) {
	// cross-schema read from users — joins flow_tasks assignees to user display names
	query, args, err := sq.
		Select(
			"ft.assignee_id AS user_id",
			"COALESCE(u.full_name, CONCAT('User ', ft.assignee_id)) AS name",
			"COUNT(*) AS total_tasks",
			"COUNT(CASE WHEN ft.status = 'done' THEN 1 END) AS completed_tasks",
			"COUNT(CASE WHEN ft.status = 'in_progress' THEN 1 END) AS in_progress_tasks",
			"COUNT(CASE WHEN ft.status NOT IN ('done','cancelled') AND ft.due_date < CURDATE() THEN 1 END) AS overdue_tasks",
		).
		From("flow_tasks ft").
		LeftJoin("users u ON ft.assignee_id = u.id").
		Where(sq.Eq{"ft.business_id": businessID}).
		Where(sq.Expr("ft.assignee_id IS NOT NULL")).
		GroupBy("ft.assignee_id", "u.full_name").
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("GetTeamWorkload build query failed", logger.Error(err))
		return domain.TeamWorkloadResult{}, err
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
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

func (r *Repository) GetRoadmap(ctx context.Context, businessID int64) (domain.RoadmapResult, error) {
	projQuery, projArgs, err := sq.
		Select(
			"fp.id",
			"fp.name",
			"COALESCE(DATE_FORMAT(fp.start_date, '%Y-%m-%d'), '') AS start_date",
			"COALESCE(DATE_FORMAT(fp.due_date, '%Y-%m-%d'), '') AS due_date",
			"COUNT(ft.id) AS total_tasks",
			"COUNT(CASE WHEN ft.status = 'done' THEN 1 END) AS done_tasks",
		).
		From("flow_projects fp").
		LeftJoin("flow_tasks ft ON ft.project_id = fp.id AND ft.parent_task_id IS NULL").
		Where(sq.Eq{"fp.business_id": businessID}).
		Where(sq.NotEq{"fp.status": "archived"}).
		GroupBy("fp.id", "fp.name", "fp.start_date", "fp.due_date").
		OrderBy("fp.created_at ASC").
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("GetRoadmap: build projects query failed", logger.Error(err))
		return domain.RoadmapResult{}, err
	}

	projRows, err := r.pool.QueryContext(ctx, projQuery, projArgs...)
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

	// Fetch milestones for every project in a single batched query, then map
	// them back by project ID, instead of issuing one query per project.
	if len(result.Projects) > 0 {
		projectIDs := make([]int64, len(result.Projects))
		for i := range result.Projects {
			projectIDs[i] = result.Projects[i].ID
		}

		byProject, err := r.getMilestonesByProjectIDs(ctx, projectIDs)
		if err != nil {
			logger.Ctx(ctx).Error("GetRoadmap: milestones query failed", logger.Error(err))
		} else {
			for i := range result.Projects {
				if ms := byProject[result.Projects[i].ID]; ms != nil {
					result.Projects[i].Milestones = ms
				}
			}
		}
	}

	if result.Projects == nil {
		result.Projects = []domain.RoadmapProject{}
	}

	return result, nil
}

// getMilestonesByProjectIDs fetches milestones for several projects at once and
// returns them grouped by project ID.
func (r *Repository) getMilestonesByProjectIDs(ctx context.Context, projectIDs []int64) (map[int64][]domain.RoadmapMilestone, error) {
	if len(projectIDs) == 0 {
		return map[int64][]domain.RoadmapMilestone{}, nil
	}

	query, args, err := sq.
		Select(
			"fm.project_id",
			"fm.id",
			"fm.name",
			"COALESCE(DATE_FORMAT(fm.target_date, '%Y-%m-%d'), '') AS due_date",
			"COUNT(ft.id) AS task_count",
			"COUNT(CASE WHEN ft.status = 'done' THEN 1 END) AS done_count",
		).
		From("flow_milestones fm").
		LeftJoin("flow_tasks ft ON ft.project_id = fm.project_id AND ft.parent_task_id IS NULL").
		Where(sq.Eq{"fm.project_id": projectIDs}).
		GroupBy("fm.project_id", "fm.id", "fm.name", "fm.target_date").
		OrderBy("fm.target_date ASC").
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	byProject := make(map[int64][]domain.RoadmapMilestone, len(projectIDs))
	for rows.Next() {
		var projectID int64
		var m domain.RoadmapMilestone
		if err := rows.Scan(&projectID, &m.ID, &m.Name, &m.DueDate, &m.TaskCount, &m.DoneCount); err != nil {
			continue
		}
		byProject[projectID] = append(byProject[projectID], m)
	}

	return byProject, rows.Err()
}
