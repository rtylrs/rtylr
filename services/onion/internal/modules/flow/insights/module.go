package flow_insights

import (
	"context"
	"database/sql"
	"time"

	moduleinfra "github.com/voxire/onion/internal/modules/flow/insights/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/flow/insights/internal/application"
)

// FlowInsights is a data provider, not an HTTP surface. Its legacy chi routes
// were shadowed duplicates of the onekit flow.insights.v1 service that
// workspaces serves with pagination, and they are gone (#683). The module
// stays because insights.dashboard.v1 and echo consume the readers
// below rather than growing their own cross-schema queries.
type Module struct {
	service       *application.Service
	assigneeNames application.AssigneeNames
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{
		service: application.NewService(ctx, moduleinfra.NewRepository(db)),
	}
}

// The readers below exist so another module can ask Flow for its own
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
	Visibility     string
	StartDate      string
	DueDate        string
	Progress       float64
	TotalTasks     int64
	CompletedTasks int64
}

// FlowProjectTime is the time logged on one project's tasks. Visibility and
// Status are carried so a business-wide view can keep private projects
// unnamed.
type FlowProjectTime struct {
	ProjectID  int64
	Name       string
	Visibility string
	Status     string
	Minutes    int64
}

// FlowLoggedTime is the time logged in Flow over a window, in minutes: the
// business total and its split per project, most time first.
type FlowLoggedTime struct {
	TotalMinutes int64
	Projects     []FlowProjectTime
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
			Visibility:     p.Visibility,
			StartDate:      p.StartDate,
			DueDate:        p.DueDate,
			Progress:       p.Progress,
			TotalTasks:     p.TotalTasks,
			CompletedTasks: p.CompletedTasks,
		})
	}

	return projects, nil
}

// FlowLoggedTime returns the minutes logged on Flow tasks with logged_at in
// [from, to]. Only Flow's own time logs count; rows linked to an HR time entry
// belong to HR.
func (m *Module) FlowLoggedTime(ctx context.Context, businessID int64, from, to time.Time) (FlowLoggedTime, error) {
	result, err := m.service.GetLoggedTime(ctx, businessID, from, to)
	if err != nil {
		return FlowLoggedTime{}, err
	}

	projects := make([]FlowProjectTime, 0, len(result.Projects))
	for _, p := range result.Projects {
		projects = append(projects, FlowProjectTime{
			ProjectID:  p.ProjectID,
			Name:       p.Name,
			Visibility: p.Visibility,
			Status:     p.Status,
			Minutes:    p.Minutes,
		})
	}

	return FlowLoggedTime{TotalMinutes: result.TotalMinutes, Projects: projects}, nil
}

// FlowDashboardProject is a project a user may open a project dashboard for.
type FlowDashboardProject struct {
	ID         int64
	Name       string
	Status     string
	Visibility string
}

// FlowOverdueTask is an open task past its due date. AssigneeID is zero when
// the task is unassigned. AssigneeName is the name Flow's task list shows for
// the assignee, empty when the task is unassigned or the name is unknown.
type FlowOverdueTask struct {
	ID           int64
	Title        string
	DueDate      string
	AssigneeID   int64
	AssigneeName string
}

// AssigneeNames reads the display name Flow's task list shows for each of
// userIDs. The workspaces module serves it.
type AssigneeNames = application.AssigneeNames

// SetAssigneeNames wires the name lookup FlowProjectProgress uses to name
// overdue tasks' assignees. Without it the names stay empty.
func (m *Module) SetAssigneeNames(names AssigneeNames) {
	m.assigneeNames = names
}

// FlowProjectProgress is one project's task completion and overdue work.
// TotalTasks and CompletedTasks count top-level tasks, like FlowRoadmap;
// OverdueTasks counts every open task past its due date, like FlowInsights.
// Overdue lists the five oldest.
type FlowProjectProgress struct {
	TotalTasks     int64
	CompletedTasks int64
	OverdueTasks   int64
	Overdue        []FlowOverdueTask
}

// FlowProjectTimeTotals is the minutes logged on one project's tasks, all time and
// in two windows. Like FlowLoggedTime, only Flow's own time logs count.
type FlowProjectTimeTotals struct {
	AllTimeMinutes int64
	ThisMinutes    int64
	PriorMinutes   int64
}

// FlowDashboardProjects returns the projects userID may see in businessID:
// every public project and the private ones they own, created or are a
// project member of. Personal and archived projects are left out.
func (m *Module) FlowDashboardProjects(ctx context.Context, businessID, userID int64) ([]FlowDashboardProject, error) {
	result, err := m.service.ListDashboardProjects(ctx, businessID, userID)
	if err != nil {
		return nil, err
	}
	projects := make([]FlowDashboardProject, 0, len(result))
	for _, p := range result {
		projects = append(projects, FlowDashboardProject(p))
	}
	return projects, nil
}

// FlowDashboardProject returns one project under the FlowDashboardProjects
// rule. found is false for a missing project, another business's project and
// a private project userID is not a member of alike.
func (m *Module) FlowDashboardProject(ctx context.Context, businessID, userID, projectID int64) (FlowDashboardProject, bool, error) {
	p, found, err := m.service.GetDashboardProject(ctx, businessID, userID, projectID)
	if err != nil || !found {
		return FlowDashboardProject{}, false, err
	}
	return FlowDashboardProject(p), true, nil
}

// FlowProjectProgress returns one project's task completion and overdue work,
// with each overdue task's assignee named the way Flow's task list names them.
// It does not check visibility: call FlowDashboardProject first.
func (m *Module) FlowProjectProgress(ctx context.Context, businessID, projectID int64) (FlowProjectProgress, error) {
	result, err := m.service.GetProjectTaskProgress(ctx, businessID, projectID)
	if err != nil {
		return FlowProjectProgress{}, err
	}
	if err := application.NameOverdueAssignees(ctx, m.assigneeNames, businessID, result.Overdue); err != nil {
		return FlowProjectProgress{}, err
	}
	overdue := make([]FlowOverdueTask, 0, len(result.Overdue))
	for _, t := range result.Overdue {
		overdue = append(overdue, FlowOverdueTask(t))
	}
	return FlowProjectProgress{
		TotalTasks:     result.TotalTasks,
		CompletedTasks: result.CompletedTasks,
		OverdueTasks:   result.OverdueTasks,
		Overdue:        overdue,
	}, nil
}

// FlowProjectTime returns the minutes logged on one project's tasks, all time
// and with logged_at in [thisFrom, thisTo] and [priorFrom, priorTo]. It does
// not check visibility: call FlowDashboardProject first.
func (m *Module) FlowProjectTime(ctx context.Context, businessID, projectID int64, thisFrom, thisTo, priorFrom, priorTo time.Time) (FlowProjectTimeTotals, error) {
	result, err := m.service.GetProjectTimeTotals(ctx, businessID, projectID, thisFrom, thisTo, priorFrom, priorTo)
	if err != nil {
		return FlowProjectTimeTotals{}, err
	}
	return FlowProjectTimeTotals(result), nil
}
