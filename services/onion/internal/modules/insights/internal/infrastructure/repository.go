package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/insights/internal/infrastructure/persistence/sqlc"
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

type InsightsData struct {
	TotalTasks       int64
	CompletedTasks   int64
	InProgressTasks  int64
	OverdueTasks     int64
	AvgCycleTimeDays float64
	TasksThisWeek    int64
	TasksLastWeek    int64
}

func (r *Repository) GetInsights(ctx context.Context, ownerID int64, workspaceID *int64) (InsightsData, error) {
	var data InsightsData

	if workspaceID != nil {
		row, err := r.queries.GetInsightsByWorkspace(ctx, sqlc.GetInsightsByWorkspaceParams{
			OwnerID:     ownerID,
			WorkspaceID: *workspaceID,
		})
		if err != nil {
			logger.Ctx(ctx).Error("GetInsights failed", logger.Error(err))
			return InsightsData{}, err
		}
		data.TotalTasks = row.TotalTasks
		data.CompletedTasks = row.CompletedTasks
		data.InProgressTasks = row.InProgressTasks
		data.OverdueTasks = row.OverdueTasks
		data.AvgCycleTimeDays = interfaceToFloat(row.AvgCycleTimeDays)
	} else {
		row, err := r.queries.GetInsightsByOwner(ctx, ownerID)
		if err != nil {
			logger.Ctx(ctx).Error("GetInsights failed", logger.Error(err))
			return InsightsData{}, err
		}
		data.TotalTasks = row.TotalTasks
		data.CompletedTasks = row.CompletedTasks
		data.InProgressTasks = row.InProgressTasks
		data.OverdueTasks = row.OverdueTasks
		data.AvgCycleTimeDays = interfaceToFloat(row.AvgCycleTimeDays)
	}

	weekStart := time.Now().AddDate(0, 0, -int(time.Now().Weekday()))
	lastWeekStart := weekStart.AddDate(0, 0, -7)

	if workspaceID != nil {
		tp, err := r.queries.GetThroughputByWorkspace(ctx, sqlc.GetThroughputByWorkspaceParams{
			CompletedAt:   sql.NullTime{Time: weekStart, Valid: true},
			CompletedAt_2: sql.NullTime{Time: lastWeekStart, Valid: true},
			CompletedAt_3: sql.NullTime{Time: weekStart, Valid: true},
			OwnerID:       ownerID,
			WorkspaceID:   *workspaceID,
		})
		if err != nil {
			logger.Ctx(ctx).Error("GetInsights throughput failed", logger.Error(err))
		} else {
			data.TasksThisWeek = tp.TasksThisWeek
			data.TasksLastWeek = tp.TasksLastWeek
		}
	} else {
		tp, err := r.queries.GetThroughputByOwner(ctx, sqlc.GetThroughputByOwnerParams{
			CompletedAt:   sql.NullTime{Time: weekStart, Valid: true},
			CompletedAt_2: sql.NullTime{Time: lastWeekStart, Valid: true},
			CompletedAt_3: sql.NullTime{Time: weekStart, Valid: true},
			OwnerID:       ownerID,
		})
		if err != nil {
			logger.Ctx(ctx).Error("GetInsights throughput failed", logger.Error(err))
		} else {
			data.TasksThisWeek = tp.TasksThisWeek
			data.TasksLastWeek = tp.TasksLastWeek
		}
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
	if workspaceID != nil {
		rows, err := r.queries.GetTeamWorkloadByWorkspace(ctx, sqlc.GetTeamWorkloadByWorkspaceParams{
			OwnerID:     ownerID,
			WorkspaceID: *workspaceID,
		})
		if err != nil {
			logger.Ctx(ctx).Error("GetTeamWorkload failed", logger.Error(err))
			return nil, err
		}
		var entries []WorkloadEntry
		for _, row := range rows {
			entries = append(entries, WorkloadEntry{
				UserID:          row.AssigneeID.Int64,
				OpenTasks:       row.OpenTasks,
				InProgressTasks: row.InProgressTasks,
				CompletedTasks:  row.CompletedTasks,
				OverdueTasks:    row.OverdueTasks,
			})
		}
		return entries, nil
	}

	rows, err := r.queries.GetTeamWorkloadByOwner(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("GetTeamWorkload failed", logger.Error(err))
		return nil, err
	}
	var entries []WorkloadEntry
	for _, row := range rows {
		entries = append(entries, WorkloadEntry{
			UserID:          row.AssigneeID.Int64,
			OpenTasks:       row.OpenTasks,
			InProgressTasks: row.InProgressTasks,
			CompletedTasks:  row.CompletedTasks,
			OverdueTasks:    row.OverdueTasks,
		})
	}
	return entries, nil
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
	if workspaceID != nil {
		rows, err := r.queries.GetRoadmapByWorkspace(ctx, sqlc.GetRoadmapByWorkspaceParams{
			OwnerID:     ownerID,
			WorkspaceID: *workspaceID,
		})
		if err != nil {
			logger.Ctx(ctx).Error("GetRoadmap failed", logger.Error(err))
			return nil, err
		}
		return roadmapFromWorkspaceRows(rows), nil
	}

	rows, err := r.queries.GetRoadmapByOwner(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("GetRoadmap failed", logger.Error(err))
		return nil, err
	}
	return roadmapFromOwnerRows(rows), nil
}

func roadmapFromOwnerRows(rows []sqlc.GetRoadmapByOwnerRow) []RoadmapProject {
	var projects []RoadmapProject
	for _, row := range rows {
		rp := RoadmapProject{
			ProjectID:      row.ID,
			ProjectName:    row.Name,
			Status:         string(row.Status.FlowProjectsStatus),
			TotalTasks:     row.TotalTasks,
			CompletedTasks: row.CompletedTasks,
		}
		if row.StartDate.Valid {
			rp.StartDate = &row.StartDate.Time
		}
		if row.DueDate.Valid {
			rp.DueDate = &row.DueDate.Time
		}
		projects = append(projects, rp)
	}
	return projects
}

func roadmapFromWorkspaceRows(rows []sqlc.GetRoadmapByWorkspaceRow) []RoadmapProject {
	var projects []RoadmapProject
	for _, row := range rows {
		rp := RoadmapProject{
			ProjectID:      row.ID,
			ProjectName:    row.Name,
			Status:         string(row.Status.FlowProjectsStatus),
			TotalTasks:     row.TotalTasks,
			CompletedTasks: row.CompletedTasks,
		}
		if row.StartDate.Valid {
			rp.StartDate = &row.StartDate.Time
		}
		if row.DueDate.Valid {
			rp.DueDate = &row.DueDate.Time
		}
		projects = append(projects, rp)
	}
	return projects
}

func interfaceToFloat(v interface{}) float64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case []byte:
		f, _ := strconv.ParseFloat(string(val), 64)
		return f
	case string:
		f, _ := strconv.ParseFloat(val, 64)
		return f
	case float64:
		return val
	default:
		f, _ := strconv.ParseFloat(fmt.Sprintf("%v", val), 64)
		return f
	}
}
