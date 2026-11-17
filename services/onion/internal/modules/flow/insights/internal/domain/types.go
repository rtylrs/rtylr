package domain

// InsightsStats represents task velocity and completion statistics.
type InsightsStats struct {
	TotalTasks        int64
	CompletedTasks    int64
	CompletionRate    float64
	AvgCompletionDays float64
	VelocityThisWeek  int64
	VelocityLastWeek  int64
	OverdueCount      int64
	TasksByStatus     map[string]int64
	TasksByPriority   map[string]int64
}

// TeamMemberWorkload represents per-member task distribution.
type TeamMemberWorkload struct {
	UserID          int64
	TotalTasks      int64
	CompletedTasks  int64
	InProgressTasks int64
	OverdueTasks    int64
}

// TeamWorkloadResult is the full team workload response.
type TeamWorkloadResult struct {
	Members []TeamMemberWorkload
}

// RoadmapMilestone represents a milestone within a project roadmap.
type RoadmapMilestone struct {
	ID        int64
	Name      string
	DueDate   string
	TaskCount int64
	DoneCount int64
}

// RoadmapProject represents a project with milestone progress.
//
// TotalTasks and CompletedTasks are the counts Progress is derived from. They
// were computed by the query and discarded before; insights.dashboard.v1's
// RoadmapProject reports them separately, so they are carried through rather
// than recomputed.
type RoadmapProject struct {
	ID             int64
	Name           string
	Status         string
	Visibility     string
	StartDate      string
	DueDate        string
	Progress       float64
	TotalTasks     int64
	CompletedTasks int64
	Milestones     []RoadmapMilestone
}

// RoadmapResult is the full roadmap response.
type RoadmapResult struct {
	Projects []RoadmapProject
}

// ProjectLoggedTime is the time logged on one project's tasks over a window.
// Visibility and Status travel with it so a business-wide consumer can name
// only the projects it may show.
type ProjectLoggedTime struct {
	ProjectID  int64
	Name       string
	Visibility string
	Status     string
	Minutes    int64
}

// LoggedTime is the time logged in Flow over a window: the business total and
// its split per project, most time first.
type LoggedTime struct {
	TotalMinutes int64
	Projects     []ProjectLoggedTime
}
