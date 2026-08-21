package domain

// InsightsStats represents task velocity and completion statistics.
type InsightsStats struct {
	TotalTasks        int64            `json:"totalTasks"`
	CompletedTasks    int64            `json:"completedTasks"`
	CompletionRate    float64          `json:"completionRate"`
	AvgCompletionDays float64          `json:"avgCompletionDays"`
	VelocityThisWeek  int64            `json:"velocityThisWeek"`
	VelocityLastWeek  int64            `json:"velocityLastWeek"`
	OverdueCount      int64            `json:"overdueCount"`
	TasksByStatus     map[string]int64 `json:"tasksByStatus"`
	TasksByPriority   map[string]int64 `json:"tasksByPriority"`
}

// TeamMemberWorkload represents per-member task distribution.
type TeamMemberWorkload struct {
	UserID          int64  `json:"userId"`
	Name            string `json:"name"`
	TotalTasks      int64  `json:"totalTasks"`
	CompletedTasks  int64  `json:"completedTasks"`
	InProgressTasks int64  `json:"inProgressTasks"`
	OverdueTasks    int64  `json:"overdueTasks"`
}

// TeamWorkloadResult is the full team workload response.
type TeamWorkloadResult struct {
	Members []TeamMemberWorkload `json:"members"`
}

// RoadmapMilestone represents a milestone within a project roadmap.
type RoadmapMilestone struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	DueDate   string `json:"dueDate"`
	TaskCount int64  `json:"taskCount"`
	DoneCount int64  `json:"doneCount"`
}

// RoadmapProject represents a project with milestone progress.
//
// TotalTasks and CompletedTasks are the counts Progress is derived from. They
// were computed by the query and discarded before; insights.dashboard.v1's
// RoadmapProject reports them separately, so they are carried through rather
// than recomputed.
type RoadmapProject struct {
	ID             int64              `json:"id"`
	Name           string             `json:"name"`
	Status         string             `json:"status"`
	StartDate      string             `json:"startDate"`
	DueDate        string             `json:"dueDate"`
	Progress       float64            `json:"progress"`
	TotalTasks     int64              `json:"totalTasks"`
	CompletedTasks int64              `json:"completedTasks"`
	Milestones     []RoadmapMilestone `json:"milestones"`
}

// RoadmapResult is the full roadmap response.
type RoadmapResult struct {
	Projects []RoadmapProject `json:"projects"`
}
