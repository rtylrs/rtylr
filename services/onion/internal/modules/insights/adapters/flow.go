package insightsadapters

import (
	"context"

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
			StartDate:      project.StartDate,
			DueDate:        project.DueDate,
			Progress:       project.Progress,
			TotalTasks:     project.TotalTasks,
			CompletedTasks: project.CompletedTasks,
		})
	}
	return projects, nil
}

func NewFlowReader(module *flowinsights.Module) insights.FlowReader {
	return flowInsightsAdapter{module: module}
}
