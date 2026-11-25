package domain

import (
	"errors"
	"strings"
)

// Flow's hierarchy is epic > task > subtask, and no deeper.
//
// The two links are stored separately. A subtask points at its parent task
// through parent_task_id, exactly as before epics existed. A task points at its
// epic through epic_id. A subtask stores no epic of its own: it reads its
// parent's. Keeping the epic link off parent_task_id is what lets every
// existing "top-level task" read (boards, counts, insights) stay untouched.

const (
	TaskTypeTask    = "task"
	TaskTypeBug     = "bug"
	TaskTypeFeature = "feature"
	TaskTypeEpic    = "epic"
	TaskTypeChore   = "chore"
	TaskTypeSpike   = "spike"

	maxEpicColorLength = 20
)

var (
	ErrInvalidTaskType        = errors.New("task_type must be one of task, bug, feature, chore, spike, epic")
	ErrInvalidEpicColor       = errors.New("epic_color must be at most 20 characters")
	ErrEpicColorOnNonEpic     = errors.New("epic_color can only be set on an epic")
	ErrTaskParentNotFound     = errors.New("parent task not found")
	ErrTaskParentOtherProject = errors.New("parent must be in the same project")
	ErrTaskParentCycle        = errors.New("a task cannot be its own parent or ancestor")
	ErrTaskParentTooDeep      = errors.New("the hierarchy is limited to epic, task and subtask")
	ErrEpicCannotHaveParent   = errors.New("an epic cannot have a parent")
	ErrEpicHasSubtasks        = errors.New("a task that has subtasks cannot become an epic")
	ErrEpicHasChildren        = errors.New("an epic that has tasks cannot change type")
)

// HierarchyRuleError returns the rule sentinel err matches, when err is a rule
// violation the caller can fix by changing the request, and nil for a missing
// row or a fault.
func HierarchyRuleError(err error) error {
	for _, target := range []error{
		ErrInvalidTaskType, ErrInvalidEpicColor, ErrEpicColorOnNonEpic,
		ErrTaskParentOtherProject, ErrTaskParentCycle, ErrTaskParentTooDeep,
		ErrEpicCannotHaveParent, ErrEpicHasSubtasks, ErrEpicHasChildren,
	} {
		if errors.Is(err, target) {
			return target
		}
	}
	return nil
}

func IsValidTaskType(taskType string) bool {
	switch taskType {
	case TaskTypeTask, TaskTypeBug, TaskTypeFeature, TaskTypeChore, TaskTypeSpike, TaskTypeEpic:
		return true
	}
	return false
}

// NormalizeEpicColor trims a colour and rejects one the column cannot hold.
func NormalizeEpicColor(color string) (string, error) {
	color = strings.TrimSpace(color)
	if len(color) > maxEpicColorLength {
		return "", ErrInvalidEpicColor
	}
	return color, nil
}

// HierarchyNode is the part of a task the hierarchy rules read.
type HierarchyNode struct {
	ID           int64
	ProjectID    int64
	BusinessID   int64
	TaskType     string
	ParentTaskID *int64
	EpicID       *int64
	// Subtasks counts rows whose parent_task_id is this task. EpicChildren
	// counts rows whose epic_id is this task.
	Subtasks     int64
	EpicChildren int64
}

// TaskPlacement is where a task sits once a parent has been resolved. At most
// one side is set; both nil means a top-level task with no epic.
type TaskPlacement struct {
	ParentTaskID *int64
	EpicID       *int64
}

// PlaceUnderParent decides how child attaches to parent. An epic parent makes
// child one of the epic's tasks; any other parent makes child a subtask.
//
// child.ID is zero for a task that does not exist yet. The depth limit is what
// rules out cycles: a parent must be top-level and a task that has subtasks
// cannot become one, so the only loops left are a task under itself or under
// its own subtask, and both are named here.
func PlaceUnderParent(child, parent HierarchyNode) (TaskPlacement, error) {
	if parent.ID == 0 || parent.BusinessID != child.BusinessID {
		return TaskPlacement{}, ErrTaskParentNotFound
	}
	if child.TaskType == TaskTypeEpic {
		return TaskPlacement{}, ErrEpicCannotHaveParent
	}
	if child.ID != 0 && parent.ID == child.ID {
		return TaskPlacement{}, ErrTaskParentCycle
	}
	if parent.ProjectID != child.ProjectID {
		return TaskPlacement{}, ErrTaskParentOtherProject
	}
	parentID := parent.ID
	if parent.TaskType == TaskTypeEpic {
		return TaskPlacement{EpicID: &parentID}, nil
	}
	if parent.ParentTaskID != nil {
		if child.ID != 0 && *parent.ParentTaskID == child.ID {
			return TaskPlacement{}, ErrTaskParentCycle
		}
		return TaskPlacement{}, ErrTaskParentTooDeep
	}
	if child.Subtasks > 0 {
		return TaskPlacement{}, ErrTaskParentTooDeep
	}
	return TaskPlacement{ParentTaskID: &parentID}, nil
}

// ValidateTaskTypeChange guards the two conversions that would leave the
// hierarchy inconsistent: into an epic while attached to something, and out of
// an epic while tasks still point at it.
func ValidateTaskTypeChange(task HierarchyNode, newType string) error {
	if !IsValidTaskType(newType) {
		return ErrInvalidTaskType
	}
	if newType == task.TaskType {
		return nil
	}
	if newType == TaskTypeEpic {
		if task.ParentTaskID != nil || task.EpicID != nil {
			return ErrEpicCannotHaveParent
		}
		if task.Subtasks > 0 {
			return ErrEpicHasSubtasks
		}
	}
	if task.TaskType == TaskTypeEpic && task.EpicChildren > 0 {
		return ErrEpicHasChildren
	}
	return nil
}

// EpicRollup summarises one epic's direct tasks.
//
// Counts and estimates cover the epic's direct, unarchived tasks only: a
// subtask is a breakdown of its parent, so counting it would count the same
// work twice. LoggedMinutes also includes time logged on those tasks'
// subtasks, because that time is not recorded anywhere on the parent.
type EpicRollup struct {
	EpicID              int64
	Identifier          string
	Title               string
	Color               string
	Status              string
	TotalTasks          int64
	BacklogTasks        int64
	ActiveTasks         int64
	DoneTasks           int64
	CancelledTasks      int64
	EstimateMinutes     int64
	DoneEstimateMinutes int64
	LoggedMinutes       int64
}
