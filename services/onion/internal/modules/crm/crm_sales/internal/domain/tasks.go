package domain

import "time"

// TaskType is what a follow-up task asks for.
type TaskType string

const (
	TaskCall    TaskType = "call"
	TaskEmail   TaskType = "email"
	TaskMeeting TaskType = "meeting"
	TaskTodo    TaskType = "todo"
)

func (t TaskType) IsValid() bool {
	return t == TaskCall || t == TaskEmail || t == TaskMeeting || t == TaskTodo
}

// Task is a follow-up due on a date. A lead's or deal's next follow-up is its
// earliest open task.
type Task struct {
	ID          int64
	LeadID      *int64
	DealID      *int64
	Title       string
	Type        TaskType
	DueDate     string
	AssignedTo  *int64
	CompletedAt *time.Time
	CreatedBy   int64
	CreatedAt   time.Time
	// What the task is about, for worklists.
	LeadName    string
	CompanyName string
	Phone       string
	DealTitle   string
}

// TaskInput creates or edits a task.
type TaskInput struct {
	LeadID     *int64
	DealID     *int64
	Title      string
	Type       TaskType
	DueDate    string
	AssignedTo *int64
}

func (in *TaskInput) Normalize() error {
	if in.Type == "" {
		in.Type = TaskTodo
	}
	if !in.Type.IsValid() {
		return invalid("unknown task type %q", in.Type)
	}
	in.Title = trimTo(in.Title, 255)
	if in.Title == "" {
		in.Title = map[TaskType]string{
			TaskCall: "Call", TaskEmail: "Email", TaskMeeting: "Meeting", TaskTodo: "Follow up",
		}[in.Type]
	}
	if !IsDate(in.DueDate) {
		return invalid("a task needs a due date (YYYY-MM-DD)")
	}
	if in.AssignedTo != nil && *in.AssignedTo <= 0 {
		in.AssignedTo = nil
	}
	return nil
}

// TaskScope is which tasks a list shows.
type TaskScope string

const (
	TasksOpen      TaskScope = "open"
	TasksCompleted TaskScope = "completed"
)

// TaskFilter selects tasks. AssignedTo 0 means unassigned.
type TaskFilter struct {
	Scope      TaskScope
	AssignedTo *int64
	Due        FollowUpWindow
	LeadID     *int64
	DealID     *int64
	Cursor     string
	Limit      int
}

// TaskPage is one page of tasks.
type TaskPage struct {
	Tasks      []Task
	NextCursor string
}

// TaskCounts is the open work for a worklist header.
type TaskCounts struct {
	Overdue  int64
	DueToday int64
	Upcoming int64
}
