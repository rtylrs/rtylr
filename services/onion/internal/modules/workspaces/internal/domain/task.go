package domain

import "time"

type Task struct {
	ID             int64
	ProjectID      int64
	WorkspaceID    int64
	OwnerID        int64
	ParentTaskID   *int64
	Title          string
	Description    string
	Status         string
	Priority       string
	AssigneeID     *int64
	ReporterID     int64
	DueDate        *time.Time
	StartDate      *time.Time
	CompletedAt    *time.Time
	EstimatedHours *float64
	ActualHours    *float64
	SortOrder      int
	Labels         []int64
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type TaskComment struct {
	ID        int64
	TaskID    int64
	UserID    int64
	Content   string
	IsEdited  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type TaskActivity struct {
	ID        int64
	TaskID    int64
	UserID    int64
	Action    string
	OldValue  string
	NewValue  string
	CreatedAt time.Time
}

type TaskDetail struct {
	Task     Task
	Subtasks []Task
	Comments []TaskComment
	Activity []TaskActivity
}

type TaskInput struct {
	ProjectID      int64
	WorkspaceID    int64
	OwnerID        int64
	ParentTaskID   *int64
	Title          string
	Description    string
	Priority       string
	AssigneeID     *int64
	ReporterID     int64
	DueDate        *time.Time
	StartDate      *time.Time
	EstimatedHours *float64
	Labels         []int64
}

type UpdateTaskInput struct {
	ID          int64
	OwnerID     int64
	Title       string
	Description string
	Status      string
	Priority    string
	AssigneeID  *int64
	DueDate     *time.Time
	StartDate   *time.Time
	Labels      []int64
}

type UpdateTaskStatusInput struct {
	ID      int64
	OwnerID int64
	Status  string
}

type ListTasksFilter struct {
	ProjectID   int64
	WorkspaceID int64
	Status      string
	AssigneeID  *int64
	Priority    string
}

type Label struct {
	ID          int64
	WorkspaceID int64
	Name        string
	Color       string
	CreatedAt   time.Time
}

type LabelInput struct {
	WorkspaceID int64
	Name        string
	Color       string
}

type DashboardStats struct {
	OpenTasks           int64
	InProgressTasks     int64
	CompletedThisWeek   int64
	OverdueTasks        int64
	MyTasks             []Task
}
