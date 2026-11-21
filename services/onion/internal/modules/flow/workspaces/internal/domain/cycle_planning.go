package domain

import (
	"errors"
	"time"
)

const (
	CycleStatusUpcoming  = "upcoming"
	CycleStatusActive    = "active"
	CycleStatusCompleted = "completed"

	CycleDecisionMove    = "move"
	CycleDecisionBacklog = "backlog"
	CycleDecisionCancel  = "cancel"

	// CycleRankGap is the distance left between neighbouring ranks so a task
	// can be dropped between two others without renumbering the list.
	CycleRankGap int64 = 1024

	// MaxCycleTaskBatch caps how many tasks one add or complete call may name.
	MaxCycleTaskBatch = 500

	// DefaultCycleLengthDays is used when a cycle is started with no end date
	// and no stored length.
	DefaultCycleLengthDays = 14
)

var (
	ErrCycleNotFound       = errors.New("cycle not found")
	ErrCycleAlreadyActive  = errors.New("the project already has an active cycle")
	ErrCycleNotUpcoming    = errors.New("only an upcoming cycle can be started")
	ErrCycleCompleted      = errors.New("the cycle is completed")
	ErrCycleTaskNotFound   = errors.New("task not found in this project")
	ErrCycleRankAnchor     = errors.New("the anchor task is not in the target list")
	ErrInvalidCycleMove    = errors.New("give at most one of before_task_id and after_task_id")
	ErrInvalidCycleStatus  = errors.New("a cycle cannot return to upcoming")
	ErrInvalidCycleRequest = errors.New("invalid cycle request")
)

// CycleDecisionError reports a complete-cycle decision that cannot be applied.
// It carries the task so the client can point at the row.
type CycleDecisionError struct {
	TaskID int64
	Reason string
}

func (e *CycleDecisionError) Error() string { return e.Reason }

// CycleTotals counts a cycle's tasks by status bucket.
type CycleTotals struct {
	Backlog   int
	Active    int
	Done      int
	Cancelled int
}

// CycleCarryOver is the set of tasks a cycle received from one completed cycle.
type CycleCarryOver struct {
	FromCycleID   int64
	FromCycleName string
	TaskIDs       []int64
}

type UpdateCycleV2Input struct {
	ID         int64
	BusinessID int64
	Name       string
	StartDate  time.Time
	EndDate    time.Time
	Goal       *string
}

type StartCycleInput struct {
	ID         int64
	BusinessID int64
	// StartDate and EndDate are zero when the caller left them out.
	StartDate  time.Time
	EndDate    time.Time
	LengthDays int
	Goal       *string
	// Today is the default start date, supplied so the rule is testable.
	Today time.Time
}

type CycleTaskDecision struct {
	TaskID    int64
	Decision  string
	ToCycleID int64
}

type CompleteCycleInput struct {
	ID         int64
	BusinessID int64
	ActorID    int64
	Decisions  []CycleTaskDecision
}

// CompletedCycleTask is what completing a cycle did to one unfinished task.
type CompletedCycleTask struct {
	TaskID         int64
	Decision       string
	ToCycleID      int64
	PreviousStatus string
}

type MoveCycleTaskInput struct {
	BusinessID int64
	TaskID     int64
	// CycleID zero means the unplanned backlog.
	CycleID      int64
	BeforeTaskID int64
	AfterTaskID  int64
}

// MovedCycleTask is where a moved task landed.
type MovedCycleTask struct {
	ProjectID int64
	Rank      int64
}

// CycleMemberLoad is the estimate committed to one assignee in a cycle. UserID
// zero is the unassigned work.
type CycleMemberLoad struct {
	UserID    int64
	Minutes   int64
	TaskCount int
}

type CycleMemberCapacity struct {
	UserID           int64
	Name             string
	WorkingDays      int
	LeaveDays        int
	AvailableMinutes int
	CommittedMinutes int
	TaskCount        int
}

type CycleCapacity struct {
	Members           []CycleMemberCapacity
	LeaveApplied      bool
	MinutesPerDay     int
	UnassignedMinutes int
	UnassignedCount   int
}

// RankedTask is one entry of an ordered planning list. Ranked is false for a
// task that has never been placed (a backlog task with no rank yet).
type RankedTask struct {
	TaskID int64
	Rank   int64
	Ranked bool
}

// RankPlacement is where a moved task lands. Renumber is nil unless the
// neighbours left no room, in which case it holds a fresh rank for every task
// of the list, the moved one included.
type RankPlacement struct {
	Rank     int64
	Renumber map[int64]int64
}

// PlaceRank computes the rank for movedID in list, which is the target list in
// display order without the moved task. With beforeID or afterID the task goes
// next to that anchor; with neither it goes to the end, or to the top when
// atTop is set.
//
// The common case writes one row: the midpoint of the two neighbours. Only a
// list with unranked or touching neighbours is renumbered, once, after which
// it has gaps again.
func PlaceRank(list []RankedTask, movedID, beforeID, afterID int64, atTop bool) (RankPlacement, error) {
	index, err := rankIndex(list, beforeID, afterID, atTop)
	if err != nil {
		return RankPlacement{}, err
	}

	var prev, next *RankedTask
	if index > 0 {
		prev = &list[index-1]
	}
	if index < len(list) {
		next = &list[index]
	}
	if rank, ok := rankBetween(prev, next); ok {
		return RankPlacement{Rank: rank}, nil
	}

	renumber := make(map[int64]int64, len(list)+1)
	position := int64(0)
	assign := func(id int64) {
		position++
		renumber[id] = position * CycleRankGap
	}
	for i, item := range list {
		if i == index {
			assign(movedID)
		}
		assign(item.TaskID)
	}
	if index == len(list) {
		assign(movedID)
	}
	return RankPlacement{Rank: renumber[movedID], Renumber: renumber}, nil
}

func rankIndex(list []RankedTask, beforeID, afterID int64, atTop bool) (int, error) {
	if beforeID != 0 && afterID != 0 {
		return 0, ErrInvalidCycleMove
	}
	anchor := beforeID + afterID
	if anchor == 0 {
		if atTop {
			return 0, nil
		}
		return len(list), nil
	}
	for i, item := range list {
		if item.TaskID != anchor {
			continue
		}
		if afterID != 0 {
			return i + 1, nil
		}
		return i, nil
	}
	return 0, ErrCycleRankAnchor
}

func rankBetween(prev, next *RankedTask) (int64, bool) {
	switch {
	case prev == nil && next == nil:
		return CycleRankGap, true
	case prev == nil:
		return next.Rank - CycleRankGap, next.Ranked
	case next == nil:
		return prev.Rank + CycleRankGap, prev.Ranked
	}
	if !prev.Ranked || !next.Ranked || next.Rank-prev.Rank < 2 {
		return 0, false
	}
	return prev.Rank + (next.Rank-prev.Rank)/2, true
}

// ResolveCycleWindow applies the start rules: start today unless a date is
// given, and end on the given date, else after LengthDays, else after the
// cycle's stored length, else after DefaultCycleLengthDays.
func ResolveCycleWindow(in StartCycleInput, storedStart, storedEnd time.Time) (time.Time, time.Time, error) {
	start := in.StartDate
	if start.IsZero() {
		start = in.Today
	}
	end := in.EndDate
	if end.IsZero() {
		length := in.LengthDays
		if length <= 0 && !storedStart.IsZero() && storedEnd.After(storedStart) {
			length = int(storedEnd.Sub(storedStart).Hours() / 24)
		}
		if length <= 0 {
			length = DefaultCycleLengthDays
		}
		end = start.AddDate(0, 0, length)
	}
	if end.Before(start) {
		return time.Time{}, time.Time{}, ErrInvalidCycleRequest
	}
	return start, end, nil
}

// WorkingDaysBetween lists the Monday to Friday dates from start to end
// inclusive.
func WorkingDaysBetween(start, end time.Time) []time.Time {
	var days []time.Time
	for day := start; !day.After(end); day = day.AddDate(0, 0, 1) {
		if wd := day.Weekday(); wd != time.Saturday && wd != time.Sunday {
			days = append(days, day)
		}
	}
	return days
}
