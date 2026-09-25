package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/voxire/onion/internal/modules/flow_insights/internal/domain"
	"github.com/voxire/onion/internal/modules/flow_insights/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		queries: sqlc.New(db),
		pool:    db,
	}
}

func (r *Repository) GetInsightsStats(ctx context.Context, businessID int64) (domain.InsightsStats, error) {
	var stats domain.InsightsStats
	stats.TasksByStatus = make(map[string]int64)
	stats.TasksByPriority = make(map[string]int64)

	totals, err := r.queries.GetTotalAndCompletedTasks(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("GetInsightsStats: total query failed", logger.Error(err))
		return stats, err
	}
	stats.TotalTasks = totals.TotalTasks
	stats.CompletedTasks = totals.CompletedTasks
	stats.OverdueCount = totals.OverdueCount

	if stats.TotalTasks > 0 {
		stats.CompletionRate = float64(stats.CompletedTasks) / float64(stats.TotalTasks) * 100
	}

	avgDays, err := r.queries.GetAvgCompletionDays(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("GetInsightsStats: avg completion failed", logger.Error(err))
		return stats, err
	}
	stats.AvgCompletionDays = interfaceToFloat64(avgDays)

	velThis, err := r.queries.GetVelocityThisWeek(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("GetInsightsStats: velocity this week failed", logger.Error(err))
		return stats, err
	}
	stats.VelocityThisWeek = velThis

	velLast, err := r.queries.GetVelocityLastWeek(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("GetInsightsStats: velocity last week failed", logger.Error(err))
		return stats, err
	}
	stats.VelocityLastWeek = velLast

	statusRows, err := r.queries.GetTasksByStatus(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("GetInsightsStats: tasks by status failed", logger.Error(err))
		return stats, err
	}
	for _, row := range statusRows {
		stats.TasksByStatus[string(row.Status.FlowTasksStatus)] = row.Cnt
	}

	priorityRows, err := r.queries.GetTasksByPriority(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("GetInsightsStats: tasks by priority failed", logger.Error(err))
		return stats, err
	}
	for _, row := range priorityRows {
		stats.TasksByPriority[string(row.Priority.FlowTasksPriority)] = row.Cnt
	}

	return stats, nil
}

func (r *Repository) GetTeamWorkload(ctx context.Context, businessID int64) (domain.TeamWorkloadResult, error) {
	// cross-schema read from users — joins flow_tasks assignees to user display names
	rows, err := r.queries.GetTeamWorkload(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("GetTeamWorkload failed", logger.Error(err))
		return domain.TeamWorkloadResult{}, err
	}

	var result domain.TeamWorkloadResult
	for _, row := range rows {
		result.Members = append(result.Members, domain.TeamMemberWorkload{
			UserID:          row.UserID.Int64,
			Name:            row.Name.String,
			TotalTasks:      row.TotalTasks,
			CompletedTasks:  row.CompletedTasks,
			InProgressTasks: row.InProgressTasks,
			OverdueTasks:    row.OverdueTasks,
		})
	}
	if result.Members == nil {
		result.Members = []domain.TeamMemberWorkload{}
	}
	return result, nil
}

func (r *Repository) GetRoadmap(ctx context.Context, businessID int64) (domain.RoadmapResult, error) {
	projRows, err := r.queries.GetRoadmapProjects(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("GetRoadmap: projects query failed", logger.Error(err))
		return domain.RoadmapResult{}, err
	}

	var result domain.RoadmapResult
	for _, row := range projRows {
		p := domain.RoadmapProject{
			ID:             row.ID,
			Name:           row.Name,
			Status:         nullProjectStatus(row.Status),
			StartDate:      interfaceToString(row.StartDate),
			DueDate:        interfaceToString(row.DueDate),
			TotalTasks:     row.TotalTasks,
			CompletedTasks: row.DoneTasks,
		}
		if row.TotalTasks > 0 {
			p.Progress = float64(row.DoneTasks) / float64(row.TotalTasks) * 100
		}
		p.Milestones = []domain.RoadmapMilestone{}
		result.Projects = append(result.Projects, p)
	}

	if len(result.Projects) > 0 {
		projectIDs := make([]int64, len(result.Projects))
		for i := range result.Projects {
			projectIDs[i] = result.Projects[i].ID
		}
		byProject, err := r.getMilestonesByProjectIDs(ctx, businessID, projectIDs)
		if err != nil {
			logger.Ctx(ctx).Error("GetRoadmap: milestones query failed", logger.Error(err))
			return domain.RoadmapResult{}, err
		}
		for i := range result.Projects {
			if ms := byProject[result.Projects[i].ID]; ms != nil {
				result.Projects[i].Milestones = ms
			}
		}
	}

	if result.Projects == nil {
		result.Projects = []domain.RoadmapProject{}
	}
	return result, nil
}

func (r *Repository) getMilestonesByProjectIDs(ctx context.Context, businessID int64, projectIDs []int64) (map[int64][]domain.RoadmapMilestone, error) {
	if len(projectIDs) == 0 {
		return map[int64][]domain.RoadmapMilestone{}, nil
	}

	rows, err := r.queries.GetMilestonesByProjectIDs(ctx, sqlc.GetMilestonesByProjectIDsParams{
		BusinessID: businessID,
		ProjectIds: projectIDs,
	})
	if err != nil {
		return nil, err
	}

	byProject := make(map[int64][]domain.RoadmapMilestone, len(projectIDs))
	for _, row := range rows {
		byProject[row.ProjectID] = append(byProject[row.ProjectID], domain.RoadmapMilestone{
			ID:        row.ID,
			Name:      row.Name,
			DueDate:   interfaceToString(row.DueDate),
			TaskCount: row.TaskCount,
			DoneCount: row.DoneCount,
		})
	}
	return byProject, nil
}

func interfaceToString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case []byte:
		return string(val)
	default:
		return fmt.Sprint(v)
	}
}

func interfaceToFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int64:
		return float64(val)
	case []byte:
		f, err := strconv.ParseFloat(string(val), 64)
		if err != nil {
			logger.Global().Warn("failed to parse float value from database",
				logger.String("value", string(val)), logger.Error(err))
			return 0
		}
		return f
	case string:
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			logger.Global().Warn("failed to parse float value from database",
				logger.String("value", val), logger.Error(err))
			return 0
		}
		return f
	default:
		return 0
	}
}

// nullProjectStatus renders a nullable flow_projects.status ENUM as a string,
// empty when NULL. It deliberately does not substitute 'active': the schema
// defaults the column, so a NULL here means the row genuinely has no status and
// inventing one would report a lifecycle state the database does not hold.
func nullProjectStatus(v sqlc.NullFlowProjectsStatus) string {
	if !v.Valid {
		return ""
	}

	return string(v.FlowProjectsStatus)
}
