package infrastructure

import (
	"errors"
	"testing"
	"time"
)

func TestTimesheetWeekRange_UsesHalfOpenCalendarWeek(t *testing.T) {
	start, end, err := timesheetWeekRange("2026-05-18")
	if err != nil {
		t.Fatalf("timesheetWeekRange: %v", err)
	}
	if got, want := start.Format("2006-01-02"), "2026-05-18"; got != want {
		t.Fatalf("start = %s, want %s", got, want)
	}
	if got, want := end.Format("2006-01-02"), "2026-05-25"; got != want {
		t.Fatalf("end = %s, want %s", got, want)
	}
	if !end.Equal(start.Add(7 * 24 * time.Hour)) {
		t.Fatalf("range duration = %s, want seven days", end.Sub(start))
	}
}

func TestTimesheetWeekRange_RejectsNonMondayWeekStart(t *testing.T) {
	if _, _, err := timesheetWeekRange("2026-05-19"); !errors.Is(err, errWeekStartNotMonday) {
		t.Fatalf("non-Monday error = %v, want %v", err, errWeekStartNotMonday)
	}

	if _, _, err := timesheetWeekRange("2026-02-30"); err == nil {
		t.Fatal("invalid date did not return an error")
	}
}
