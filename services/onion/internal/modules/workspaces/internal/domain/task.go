package domain

import "time"

const (
	TaskStatusBacklog    = "backlog"
	TaskStatusTodo       = "todo"
	TaskStatusInProgress = "in_progress"
	TaskStatusInReview   = "in_review"
	TaskStatusDone       = "done"
	TaskStatusCancelled  = "cancelled"
)

// IsValidTaskStatus keeps writes aligned with the flow_tasks.status enum. The
// database reports invalid enum values as a warning/error after coercion, so
// validate at the API boundary and return a client error instead.
func IsValidTaskStatus(status string) bool {
	switch status {
	case TaskStatusBacklog, TaskStatusTodo, TaskStatusInProgress,
		TaskStatusInReview, TaskStatusDone, TaskStatusCancelled:
		return true
	default:
		return false
	}
}

type Task struct {
	ID             int64
	ProjectID      int64
	WorkspaceID    int64
	OwnerID        int64
	BusinessID     int64
	ParentTaskID   *int64
	SectionID      *int64
	CycleID        *int64
	MilestoneID    *int64
	Title          string
	Description    string
	Status         string
	StatusID       *int64
	Priority       string
	AssigneeID     *int64
	ReporterID     int64
	DueDate        *time.Time
	StartDate      *time.Time
	CompletedAt    *time.Time
	EstimatedHours *float64
	ActualHours    *float64
	EffortMinutes  *int32
	FocusPin       bool
	SortOrder      int
	Labels         []int64
	Archived       bool
	CreatedAt      time.Time
	UpdatedAt      time.Time

	// Display projections populated by list/detail reads via JOINs.
	// Safe to ignore on writes — populated server-side from related tables.
	Identifier       string // e.g. "T-42"
	ProjectName      string
	ProjectColor     string
	AssigneeName     string
	AssigneeInitials string
	SubtaskCount     int32
	StatusName       string
	StatusColor      string
	Dependencies     []int64
}

type ProjectColumn struct {
	ID          int64
	ProjectID   int64
	WorkspaceID int64
	Name        string
	Color       string
	SortOrder   int
	CreatedBy   int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ProjectColumnInput struct {
	ProjectID   int64
	WorkspaceID int64
	Name        string
	Color       string
	SortOrder   int
	CreatedBy   int64
}

type UpdateProjectColumnInput struct {
	ID        int64
	OwnerID   int64
	ActorID   int64
	Name      string
	Color     string
	SortOrder int
}

type TaskComment struct {
	ID            int64
	TaskID        int64
	UserID        int64
	ActorName     string
	ActorInitials string
	Content       string
	IsEdited      bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type TaskActivity struct {
	ID            int64
	TaskID        int64
	UserID        int64
	ActorName     string
	ActorInitials string
	Action        string
	OldValue      string
	NewValue      string
	CreatedAt     time.Time
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
	BusinessID     int64
	ParentTaskID   *int64
	SectionID      *int64
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
	BusinessID  int64
	SectionID   *int64
	Title       string
	Description string
	Status      string
	Priority    string
	AssigneeID  *int64
	DueDate     *time.Time
	StartDate   *time.Time
	Labels      []int64
	Archived    *bool
	// The fields below use pointers rather than the zero-value-means-absent
	// convention above, so they can express "clear this value" as well as
	// "leave it alone". Title and friends cannot, which is why a description
	// can never be emptied once set.
	EffortMinutes  *int32
	EstimatedHours *float64
	WaitingOn      *string
	DecisionNote   *string
	// FocusPin is per-user, so the repository writes flow_task_focus_pins
	// keyed on the caller rather than a column on flow_tasks.
	FocusPin *bool
}

type AddTaskDependencyInput struct {
	TaskID          int64
	DependsOnTaskID int64
	BusinessID      int64
}

type RemoveTaskDependencyInput struct {
	TaskID          int64
	DependsOnTaskID int64
	BusinessID      int64
}

type UpdateTaskStatusInput struct {
	ID         int64
	OwnerID    int64
	BusinessID int64
	Status     string
}

type ListTasksFilter struct {
	ProjectID   int64
	BusinessID  int64
	WorkspaceID int64
	Status      string
	AssigneeID  *int64
	Priority    string
	ActiveOnly  bool
	Compact     bool
	Limit       int64
	Offset      int64
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
	BusinessID  int64
	Name        string
	Color       string
}

type DashboardStats struct {
	OpenTasks         int64
	InProgressTasks   int64
	CompletedThisWeek int64
	OverdueTasks      int64
	MyTasks           []Task
}
