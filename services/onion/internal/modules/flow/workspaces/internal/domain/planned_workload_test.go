package domain

import (
	"testing"
	"time"
)

func dayPtr(s string) *time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return &t
}

func TestDistributePlanned(t *testing.T) {
	tasks := []PlannedTask{
		// Fri 27 Nov to Tue 1 Dec: 3 working days (Fri, Mon, Tue), 600 minutes.
		{ID: 1, AssigneeID: 7, StartDate: dayPtr("2026-11-27"), DueDate: dayPtr("2026-12-01"), EstimateMinutes: 600},
		// Due only, on Monday.
		{ID: 2, AssigneeID: 7, DueDate: dayPtr("2026-11-30"), EstimateMinutes: 90},
		// A weekend-only span keeps its estimate.
		{ID: 3, AssigneeID: 8, StartDate: dayPtr("2026-11-28"), DueDate: dayPtr("2026-11-29"), EstimateMinutes: 61},
		// No estimate, no assignee: left out.
		{ID: 4, AssigneeID: 7, DueDate: dayPtr("2026-11-30")},
		{ID: 5, DueDate: dayPtr("2026-11-30"), EstimateMinutes: 60},
	}
	got := DistributePlanned(tasks, *dayPtr("2026-11-23"), *dayPtr("2026-12-06"), nil)
	type row struct {
		user    int64
		date    string
		minutes int64
		tasks   int
	}
	var rows []row
	for _, d := range got {
		rows = append(rows, row{d.UserID, d.Date.Format("2006-01-02"), d.Minutes, len(d.TaskIDs)})
	}
	want := []row{
		{7, "2026-11-27", 200, 1},
		{7, "2026-11-30", 290, 2},
		{7, "2026-12-01", 200, 1},
		{8, "2026-11-28", 31, 1},
		{8, "2026-11-29", 30, 1},
	}
	if len(rows) != len(want) {
		t.Fatalf("got %v, want %v", rows, want)
	}
	for i := range want {
		if rows[i] != want[i] {
			t.Errorf("row %d = %v, want %v", i, rows[i], want[i])
		}
	}
}

func TestDistributePlannedHonoursTheWorkWeekAndRange(t *testing.T) {
	sunThu := []time.Weekday{time.Sunday, time.Monday, time.Tuesday, time.Wednesday, time.Thursday}
	tasks := []PlannedTask{{ID: 1, AssigneeID: 7, StartDate: dayPtr("2026-11-26"), DueDate: dayPtr("2026-11-29"), EstimateMinutes: 120}}
	got := DistributePlanned(tasks, *dayPtr("2026-11-29"), *dayPtr("2026-11-30"), sunThu)
	// Thu 26 and Sun 29 are working days on a Sunday week; only Sunday is in range.
	if len(got) != 1 || got[0].Date.Format("2006-01-02") != "2026-11-29" || got[0].Minutes != 60 {
		t.Fatalf("got %+v", got)
	}
}
