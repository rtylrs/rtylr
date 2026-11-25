package domain

import (
	"errors"
	"sort"
	"time"
)

// Planned workload (T-341): each open, assigned, estimated task spreads its
// estimate evenly over the working days from its start date to its due date
// (or the one date it has), giving planned minutes per person per day.

// MaxPlannedWorkloadDays caps a request's date range.
const MaxPlannedWorkloadDays = 62

// ErrInvalidPlannedRange rejects a missing, reversed or too long range.
var ErrInvalidPlannedRange = errors.New("from_date and to_date must be dates no more than 62 days apart")

// PlannedTask is a task that contributes to someone's plan.
type PlannedTask struct {
	ID              int64
	Title           string
	ProjectID       int64
	AssigneeID      int64
	StartDate       *time.Time
	DueDate         *time.Time
	EstimateMinutes int64
}

// PlannedDay is one person's planned minutes on one day, with the tasks that
// make it up.
type PlannedDay struct {
	UserID  int64
	Date    time.Time
	Minutes int64
	TaskIDs []int64
}

// DefaultWorkingDays is Monday to Friday.
var DefaultWorkingDays = []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday}

func dateOnly(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// DistributePlanned spreads each task's estimate over the working days of its
// span and returns the days that fall in [from, to], ordered by user then
// date. A span with no working day (a weekend task) uses every day of it, so
// no estimate is dropped. Remainder minutes go to the earliest days.
func DistributePlanned(tasks []PlannedTask, from, to time.Time, workingDays []time.Weekday) []PlannedDay {
	if len(workingDays) == 0 {
		workingDays = DefaultWorkingDays
	}
	working := map[time.Weekday]bool{}
	for _, d := range workingDays {
		working[d] = true
	}
	from, to = dateOnly(from), dateOnly(to)
	type key struct {
		user int64
		day  time.Time
	}
	byKey := map[key]*PlannedDay{}
	for _, task := range tasks {
		if task.EstimateMinutes <= 0 || task.AssigneeID == 0 {
			continue
		}
		start, end := task.StartDate, task.DueDate
		if start == nil {
			start = end
		}
		if end == nil {
			end = start
		}
		if start == nil {
			continue
		}
		first, last := dateOnly(*start), dateOnly(*end)
		if last.Before(first) {
			first, last = last, first
		}
		var days []time.Time
		for d := first; !d.After(last); d = d.AddDate(0, 0, 1) {
			if working[d.Weekday()] {
				days = append(days, d)
			}
		}
		if len(days) == 0 {
			for d := first; !d.After(last); d = d.AddDate(0, 0, 1) {
				days = append(days, d)
			}
		}
		share := task.EstimateMinutes / int64(len(days))
		extra := task.EstimateMinutes % int64(len(days))
		for i, d := range days {
			minutes := share
			if int64(i) < extra {
				minutes++
			}
			if minutes == 0 || d.Before(from) || d.After(to) {
				continue
			}
			k := key{task.AssigneeID, d}
			if byKey[k] == nil {
				byKey[k] = &PlannedDay{UserID: task.AssigneeID, Date: d}
			}
			byKey[k].Minutes += minutes
			byKey[k].TaskIDs = append(byKey[k].TaskIDs, task.ID)
		}
	}
	out := make([]PlannedDay, 0, len(byKey))
	for _, day := range byKey {
		out = append(out, *day)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].UserID != out[j].UserID {
			return out[i].UserID < out[j].UserID
		}
		return out[i].Date.Before(out[j].Date)
	})
	return out
}
