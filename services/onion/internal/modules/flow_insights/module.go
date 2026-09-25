package flow_insights

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/flow_insights/internal/application"
)

// FlowInsights is a data provider, not an HTTP surface. Its legacy chi routes
// were shadowed duplicates of the onekit flow.insights.v1 service that
// workspaces serves with pagination, and they are gone (#683). The module
// stays because insights.dashboard.v1 and echo consume the three readers
// below rather than growing their own cross-schema queries.
type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{
		service: application.NewService(ctx, db),
	}
}

// The three readers below exist so another module can ask Flow for its own
// delivery numbers without reaching into flow_* tables itself.
//
// insights.dashboard.v1 declares getInsights, getWorkload and getRoadmap as its
// own projections of this data, distinct from the series flow.insights.v1
// serves. Those handlers returned 501; they now delegate here rather than
// growing a second set of cross-schema queries over the same tables. Each takes
// the resolved businessID and passes it straight to a tenant-filtered query.
//
// The result types are declared here rather than reused from internal/domain
// because Go's internal rule correctly forbids another module importing that
// package, and because a consumer should depend on what Flow chooses to publish
// rather than on Flow's own storage shape. Same reason hr_core publishes
// EmployeeSummary.

// FlowDelivery is Flow's task velocity and completion summary.
type FlowDelivery struct {
	TotalTasks        int64
	CompletedTasks    int64
	InProgressTasks   int64
	OverdueTasks      int64
	AvgCompletionDays float64
	VelocityThisWeek  int64
	VelocityLastWeek  int64
}

// FlowWorkloadMember is one person's Flow task distribution.
type FlowWorkloadMember struct {
	UserID          int64
	TotalTasks      int64
	CompletedTasks  int64
	InProgressTasks int64
	OverdueTasks    int64
}

// FlowRoadmapProject is one Flow project with its progress.
type FlowRoadmapProject struct {
	ID             int64
	Name           string
	Status         string
	StartDate      string
	DueDate        string
	Progress       float64
	TotalTasks     int64
	CompletedTasks int64
}

// FlowInsights returns Flow's task velocity and completion statistics.
func (m *Module) FlowInsights(ctx context.Context, businessID int64) (FlowDelivery, error) {
	stats, err := m.service.GetInsights(ctx, businessID)
	if err != nil {
		return FlowDelivery{}, err
	}

	return FlowDelivery{
		TotalTasks:     stats.TotalTasks,
		CompletedTasks: stats.CompletedTasks,
		// TasksByStatus is keyed by the flow_tasks status enum. A status with no
		// tasks is absent from the map, and indexing a missing key yields 0,
		// which is the answer.
		InProgressTasks:   stats.TasksByStatus["in_progress"],
		OverdueTasks:      stats.OverdueCount,
		AvgCompletionDays: stats.AvgCompletionDays,
		VelocityThisWeek:  stats.VelocityThisWeek,
		VelocityLastWeek:  stats.VelocityLastWeek,
	}, nil
}

// FlowTeamWorkload returns per-member task distribution across Flow.
func (m *Module) FlowTeamWorkload(ctx context.Context, businessID int64) ([]FlowWorkloadMember, error) {
	result, err := m.service.GetTeamWorkload(ctx, businessID)
	if err != nil {
		return nil, err
	}

	members := make([]FlowWorkloadMember, 0, len(result.Members))
	for _, m := range result.Members {
		members = append(members, FlowWorkloadMember{
			UserID:          m.UserID,
			TotalTasks:      m.TotalTasks,
			CompletedTasks:  m.CompletedTasks,
			InProgressTasks: m.InProgressTasks,
			OverdueTasks:    m.OverdueTasks,
		})
	}

	return members, nil
}

// FlowRoadmap returns Flow projects with their milestone progress.
func (m *Module) FlowRoadmap(ctx context.Context, businessID int64) ([]FlowRoadmapProject, error) {
	result, err := m.service.GetRoadmap(ctx, businessID)
	if err != nil {
		return nil, err
	}

	projects := make([]FlowRoadmapProject, 0, len(result.Projects))
	for _, p := range result.Projects {
		projects = append(projects, FlowRoadmapProject{
			ID:             p.ID,
			Name:           p.Name,
			Status:         p.Status,
			StartDate:      p.StartDate,
			DueDate:        p.DueDate,
			Progress:       p.Progress,
			TotalTasks:     p.TotalTasks,
			CompletedTasks: p.CompletedTasks,
		})
	}

	return projects, nil
}
