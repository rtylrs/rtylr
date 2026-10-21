package domain

import (
	"cmp"
	"slices"
	"time"

	"github.com/voxire/onion/internal/platform/money"
)

// DefaultActivityDays is the dashboard's window when the caller asks for none
// of the offered ones.
const DefaultActivityDays = 30

// ActivityWindow is the number of days to report: 7, 30 or 90.
func ActivityWindow(days int) int {
	switch days {
	case 7, 30, 90:
		return days
	}
	return DefaultActivityDays
}

// ActivityDay is one of the caller's days of sales work.
type ActivityDay struct {
	Date       time.Time
	Calls      int64
	Reached    int64
	LeadsAdded int64
	DealsWon   int64
}

// RepActivity is one person's share of a window: the calls they logged and
// the outcomes of the leads and deals assigned to them. RepID 0 is unassigned
// work.
type RepActivity struct {
	RepID          int64
	Calls          int64
	Reached        int64
	LeadsQualified int64
	DealsWon       int64
	DealsLost      int64
	WonAmount      money.Amount
}

// SalesActivity is a window of sales work: every day in order, the window's
// totals, and each rep's share, busiest first.
type SalesActivity struct {
	Days           []ActivityDay
	Calls          int64
	Reached        int64
	LeadsAdded     int64
	LeadsQualified int64
	DealsWon       int64
	DealsLost      int64
	WonAmount      money.Amount
	Reps           []RepActivity
}

// NewSalesActivity lays out the `days` days starting at first, taking each
// day's counts from byDay (keyed YYYY-MM-DD; quiet days are absent), and
// totals the window.
func NewSalesActivity(first time.Time, days int, byDay map[string]ActivityDay, reps []RepActivity) SalesActivity {
	activity := SalesActivity{Days: make([]ActivityDay, days), Reps: reps}
	for i := range activity.Days {
		date := first.AddDate(0, 0, i)
		day := byDay[date.Format(time.DateOnly)]
		day.Date = date
		activity.Days[i] = day
		activity.Calls += day.Calls
		activity.Reached += day.Reached
		activity.LeadsAdded += day.LeadsAdded
		activity.DealsWon += day.DealsWon
	}
	for _, rep := range reps {
		activity.LeadsQualified += rep.LeadsQualified
		activity.DealsLost += rep.DealsLost
		activity.WonAmount += rep.WonAmount
	}
	slices.SortFunc(activity.Reps, func(a, b RepActivity) int {
		return cmp.Or(
			cmp.Compare(b.Calls, a.Calls),
			cmp.Compare(b.DealsWon, a.DealsWon),
			cmp.Compare(b.WonAmount, a.WonAmount),
			cmp.Compare(a.RepID, b.RepID),
		)
	})
	return activity
}
