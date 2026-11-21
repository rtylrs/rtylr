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
	WaitingOn    string
	DecisionNote string
	ID           int64
	ProjectID    int64
	WorkspaceID  int64
	OwnerID      int64
	BusinessID   int64
	ParentTaskID *int64
	SectionID    *int64
	CycleID      *int64
	// CycleName and PlanRank come with CycleID from flow_cycle_tasks. PlanRank
	// is the rank inside the cycle, or inside the unplanned backlog when
	// CycleID is nil; nil for a backlog task that was never placed.
	CycleName      string
	PlanRank       *int64
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
	Identifier        string // e.g. "T-42"
	ProjectName       string
	ProjectColor      string
	AssigneeName      string
	AssigneeInitials  string
	SubtaskCount      int32
	CompletedSubtasks int32
	CommentCount      int32
	AttachmentCount   int32
	StatusName        string
	StatusColor       string
	Dependencies      []int64
	// FieldValues holds the task's custom field values keyed by field id.
	FieldValues map[int64]FieldValue

	// TaskType is task, bug, feature or epic. EpicColor is the epic's own
	// colour on an epic and the colour of the epic a task sits under otherwise.
	// EpicID and EpicTitle name that epic; a subtask reports its parent's.
	TaskType  string
	EpicID    *int64
	EpicTitle string
	EpicColor string
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
	Status         string
	EffortMinutes  *int32
	Priority       string
	AssigneeID     *int64
	ReporterID     int64
	DueDate        *time.Time
	StartDate      *time.Time
	EstimatedHours *float64
	Labels         []int64
	// FieldValues are custom field values to store with the new task.
	FieldValues []FieldValueInput
	// ParentTaskID may name an epic or a task on the way in. The service
	// resolves it, so the repository only ever sees a task parent here and an
	// epic in EpicID.
	TaskType  string
	EpicID    *int64
	EpicColor string
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
	// V2 carries explicit clears separately from omitted legacy fields.
	ClearDescription bool
	ClearDueDate     bool
	ClearStartDate   bool
	EffortMinutes    *int32
	EstimatedHours   *float64
	WaitingOn        *string
	DecisionNote     *string
	// FocusPin is per-user, so the repository writes flow_task_focus_pins
	// keyed on the caller rather than a column on flow_tasks.
	FocusPin *bool
	// TaskType empty leaves the type alone. EpicColor nil leaves the colour
	// alone and an empty string clears it.
	TaskType  string
	EpicColor *string
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
	// Custom field clauses, already validated against the project's fields.
	FieldFilters []TaskFieldFilter
	FieldSort    *TaskFieldSort
	// TaskTypes narrows to the listed types. EpicID narrows to one epic's
	// tasks; NoEpic to tasks that sit under no epic and are not epics.
	TaskTypes []string
	EpicID    *int64
	NoEpic    bool
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
