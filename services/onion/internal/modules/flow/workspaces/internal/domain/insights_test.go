package domain

import (
	"testing"
	"time"
)

func day(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return t
}

func at(s string) *time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return &t
}

func TestCycleBurndownStopsAtTodayAndCountsCompletionDays(t *testing.T) {
	completions := []*time.Time{
		at("2026-11-15T09:00:00Z"), // before the start: done from day 0
		at("2026-11-17T23:59:00Z"), // end of day 1
		nil,
		nil,
	}
	got := CycleBurndown(day("2026-11-16"), day("2026-11-20"), day("2026-11-18"), completions)
	want := []InsightsBurndownPoint{
		{Day: 0, Ideal: 4, Actual: 3},
		{Day: 1, Ideal: 3, Actual: 2},
		{Day: 2, Ideal: 2, Actual: 2},
	}
	if len(got) != len(want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("point %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestCycleBurndownEdges(t *testing.T) {
	one := []*time.Time{nil}
	if got := CycleBurndown(day("2026-11-16"), day("2026-11-20"), day("2026-11-16"), one); len(got) != 2 {
		t.Fatalf("first day = %+v, want two points so a line can be drawn", got)
	}
	if got := CycleBurndown(day("2026-11-16"), day("2026-11-20"), day("2026-12-01"), one); len(got) != 5 || got[4].Ideal != 0 {
		t.Fatalf("after the end = %+v, want the five cycle days ending at ideal 0", got)
	}
	if got := CycleBurndown(day("2026-11-16"), day("2026-11-20"), day("2026-11-10"), one); len(got) != 0 {
		t.Fatalf("not begun = %+v, want none", got)
	}
	if got := CycleBurndown(day("2026-11-16"), day("2026-11-20"), day("2026-11-18"), nil); len(got) != 0 {
		t.Fatalf("empty scope = %+v, want none", got)
	}
}

func TestWeeklyThroughputKeepsEmptyWeeksOldestFirst(t *testing.T) {
	// Wednesday 25 November 2026; its week starts Monday the 23rd.
	got := WeeklyThroughput(map[string]int{
		"2026-11-25": 2, "2026-11-23": 1, "2026-11-12": 4, "2026-10-01": 9,
	}, day("2026-11-25"), 3)
	want := []InsightsThroughput{{Week: "2026-11-09", Completed: 4}, {Week: "2026-11-16"}, {Week: "2026-11-23", Completed: 3}}
	if len(got) != len(want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("week %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}
