package domain

import (
	"slices"
	"testing"
	"time"
)

func TestNewSalesActivityLaysOutEveryDayAndTotals(t *testing.T) {
	first := time.Date(2026, 10, 18, 0, 0, 0, 0, time.UTC)
	activity := NewSalesActivity(first, 3, map[string]ActivityDay{
		"2026-10-18": {Calls: 4, Reached: 1},
		"2026-10-20": {Calls: 2, Reached: 2, LeadsAdded: 5, DealsWon: 1},
		"2026-10-21": {Calls: 9}, // after the window: not laid out
	}, []RepActivity{
		{RepID: 7, Calls: 2, LeadsQualified: 1},
		{RepID: 3, Calls: 4, DealsWon: 1, DealsLost: 2, WonAmount: 1500},
		{RepID: 0, LeadsQualified: 2},
		{RepID: 9, Calls: 4},
	})

	var dates []string
	for _, day := range activity.Days {
		dates = append(dates, day.Date.Format(time.DateOnly))
	}
	if want := []string{"2026-10-18", "2026-10-19", "2026-10-20"}; !slices.Equal(dates, want) {
		t.Fatalf("days = %v, want %v", dates, want)
	}
	if activity.Days[1].Calls != 0 || activity.Days[2].LeadsAdded != 5 {
		t.Errorf("quiet day or counts wrong: %+v", activity.Days)
	}
	if activity.Calls != 6 || activity.Reached != 3 || activity.LeadsAdded != 5 || activity.DealsWon != 1 {
		t.Errorf("day totals = %+v", activity)
	}
	if activity.LeadsQualified != 3 || activity.DealsLost != 2 || activity.WonAmount != 1500 {
		t.Errorf("rep totals = %+v", activity)
	}
	var order []int64
	for _, rep := range activity.Reps {
		order = append(order, rep.RepID)
	}
	// Most calls first, then deals won, then the lower id.
	if want := []int64{3, 9, 7, 0}; !slices.Equal(order, want) {
		t.Errorf("rep order = %v, want %v", order, want)
	}
}

func TestActivityWindowOffersThreeLengths(t *testing.T) {
	for in, want := range map[int]int{7: 7, 30: 30, 90: 90, 0: 30, 14: 30, 365: 30} {
		if got := ActivityWindow(in); got != want {
			t.Errorf("ActivityWindow(%d) = %d, want %d", in, got, want)
		}
	}
}
