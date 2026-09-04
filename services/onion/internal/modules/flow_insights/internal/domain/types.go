package domain

// InsightsStats represents task velocity and completion statistics.
type InsightsStats struct {
	TotalTasks        int64            `json:"total_tasks"`
	CompletedTasks    int64            `json:"completed_tasks"`
	CompletionRate    float64          `json:"completion_rate"`
	AvgCompletionDays float64          `json:"avg_completion_days"`
	VelocityThisWeek  int64            `json:"velocity_this_week"`
	VelocityLastWeek  int64            `json:"velocity_last_week"`
	OverdueCount      int64            `json:"overdue_count"`
	TasksByStatus     map[string]int64 `json:"tasks_by_status"`
	TasksByPriority   map[string]int64 `json:"tasks_by_priority"`
}

// TeamMemberWorkload represents per-member task distribution.
type TeamMemberWorkload struct {
	UserID          int64  `json:"user_id"`
	Name            string `json:"name"`
	TotalTasks      int64  `json:"total_tasks"`
	CompletedTasks  int64  `json:"completed_tasks"`
	InProgressTasks int64  `json:"in_progress_tasks"`
	OverdueTasks    int64  `json:"overdue_tasks"`
}

// TeamWorkloadResult is the full team workload response.
type TeamWorkloadResult struct {
	Members []TeamMemberWorkload `json:"members"`
}

// RoadmapMilestone represents a milestone within a project roadmap.
type RoadmapMilestone struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	DueDate   string `json:"due_date"`
	TaskCount int64  `json:"task_count"`
	DoneCount int64  `json:"done_count"`
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
	StartDate      string             `json:"start_date"`
	DueDate        string             `json:"due_date"`
	Progress       float64            `json:"progress"`
	TotalTasks     int64              `json:"total_tasks"`
	CompletedTasks int64              `json:"completed_tasks"`
	Milestones     []RoadmapMilestone `json:"milestones"`
}

// RoadmapResult is the full roadmap response.
type RoadmapResult struct {
	Projects []RoadmapProject `json:"projects"`
}
