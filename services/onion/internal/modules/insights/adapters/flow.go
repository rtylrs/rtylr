package insightsadapters

import (
	"context"
	"time"

	flowinsights "github.com/voxire/onion/internal/modules/flow/insights"
	"github.com/voxire/onion/internal/modules/insights"
)

type flowInsightsAdapter struct {
	module *flowinsights.Module
}

func (a flowInsightsAdapter) FlowInsights(ctx context.Context, businessID int64) (insights.FlowDelivery, error) {
	result, err := a.module.FlowInsights(ctx, businessID)
	return insights.FlowDelivery{
		TotalTasks:        result.TotalTasks,
		CompletedTasks:    result.CompletedTasks,
		InProgressTasks:   result.InProgressTasks,
		OverdueTasks:      result.OverdueTasks,
		AvgCompletionDays: result.AvgCompletionDays,
		VelocityThisWeek:  result.VelocityThisWeek,
		VelocityLastWeek:  result.VelocityLastWeek,
	}, err
}

func (a flowInsightsAdapter) FlowTeamWorkload(ctx context.Context, businessID int64) ([]insights.FlowWorkloadMember, error) {
	result, err := a.module.FlowTeamWorkload(ctx, businessID)
	if err != nil {
		return nil, err
	}
	members := make([]insights.FlowWorkloadMember, 0, len(result))
	for _, member := range result {
		members = append(members, insights.FlowWorkloadMember{
			UserID:          member.UserID,
			TotalTasks:      member.TotalTasks,
			CompletedTasks:  member.CompletedTasks,
			InProgressTasks: member.InProgressTasks,
			OverdueTasks:    member.OverdueTasks,
		})
	}
	return members, nil
}

func (a flowInsightsAdapter) FlowRoadmap(ctx context.Context, businessID int64) ([]insights.FlowRoadmapProject, error) {
	result, err := a.module.FlowRoadmap(ctx, businessID)
	if err != nil {
		return nil, err
	}
	projects := make([]insights.FlowRoadmapProject, 0, len(result))
	for _, project := range result {
		projects = append(projects, insights.FlowRoadmapProject{
			ID:             project.ID,
			Name:           project.Name,
			Status:         project.Status,
			Visibility:     project.Visibility,
			StartDate:      project.StartDate,
			DueDate:        project.DueDate,
			Progress:       project.Progress,
			TotalTasks:     project.TotalTasks,
			CompletedTasks: project.CompletedTasks,
		})
	}
	return projects, nil
}

func (a flowInsightsAdapter) FlowLoggedTime(ctx context.Context, businessID int64, from, to time.Time) (insights.FlowLoggedTime, error) {
	result, err := a.module.FlowLoggedTime(ctx, businessID, from, to)
	if err != nil {
		return insights.FlowLoggedTime{}, err
	}
	projects := make([]insights.FlowProjectTime, 0, len(result.Projects))
	for _, project := range result.Projects {
		projects = append(projects, insights.FlowProjectTime{
			ProjectID:  project.ProjectID,
			Name:       project.Name,
			Visibility: project.Visibility,
			Status:     project.Status,
			Minutes:    project.Minutes,
		})
	}
	return insights.FlowLoggedTime{TotalMinutes: result.TotalMinutes, Projects: projects}, nil
}

func (a flowInsightsAdapter) FlowDashboardProjects(ctx context.Context, businessID, userID int64) ([]insights.DashboardProject, error) {
	result, err := a.module.FlowDashboardProjects(ctx, businessID, userID)
	if err != nil {
		return nil, err
	}
	projects := make([]insights.DashboardProject, 0, len(result))
	for _, project := range result {
		projects = append(projects, insights.DashboardProject(project))
	}
	return projects, nil
}

func (a flowInsightsAdapter) FlowDashboardProject(ctx context.Context, businessID, userID, projectID int64) (insights.DashboardProject, bool, error) {
	project, found, err := a.module.FlowDashboardProject(ctx, businessID, userID, projectID)
	if err != nil || !found {
		return insights.DashboardProject{}, false, err
	}
	return insights.DashboardProject(project), true, nil
}

func (a flowInsightsAdapter) FlowProjectProgress(ctx context.Context, businessID, projectID int64) (insights.FlowProjectProgress, error) {
	result, err := a.module.FlowProjectProgress(ctx, businessID, projectID)
	if err != nil {
		return insights.FlowProjectProgress{}, err
	}
	overdue := make([]insights.ProjectOverdueTask, 0, len(result.Overdue))
	for _, task := range result.Overdue {
		overdue = append(overdue, insights.ProjectOverdueTask(task))
	}
	return insights.FlowProjectProgress{
		TotalTasks:     result.TotalTasks,
		CompletedTasks: result.CompletedTasks,
		OverdueTasks:   result.OverdueTasks,
		Overdue:        overdue,
	}, nil
}

func (a flowInsightsAdapter) FlowProjectTime(ctx context.Context, businessID, projectID int64, thisFrom, thisTo, priorFrom, priorTo time.Time) (insights.FlowProjectTimeTotals, error) {
	result, err := a.module.FlowProjectTime(ctx, businessID, projectID, thisFrom, thisTo, priorFrom, priorTo)
	if err != nil {
		return insights.FlowProjectTimeTotals{}, err
	}
	return insights.FlowProjectTimeTotals(result), nil
}

func NewFlowReader(module *flowinsights.Module) insights.FlowReader {
	return flowInsightsAdapter{module: module}
}
