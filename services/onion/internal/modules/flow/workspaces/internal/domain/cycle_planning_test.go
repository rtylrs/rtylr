package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
)

func ranked(pairs ...int64) []domain.RankedTask {
	out := make([]domain.RankedTask, 0, len(pairs)/2)
	for i := 0; i+1 < len(pairs); i += 2 {
		out = append(out, domain.RankedTask{TaskID: pairs[i], Rank: pairs[i+1], Ranked: true})
	}
	return out
}

func TestPlaceRankWritesOneRowWhenThereIsRoom(t *testing.T) {
	list := ranked(1, 1024, 2, 2048, 3, 3072)
	cases := []struct {
		name          string
		before, after int64
		atTop         bool
		want          int64
	}{
		{"end of a cycle", 0, 0, false, 4096},
		{"top of the backlog", 0, 0, true, 0},
		{"before the second", 2, 0, false, 1536},
		{"after the second", 0, 2, false, 2560},
		{"before the first", 1, 0, false, 0},
	}
	for _, tc := range cases {
		got, err := domain.PlaceRank(list, 9, tc.before, tc.after, tc.atTop)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if got.Rank != tc.want || got.Renumber != nil {
			t.Errorf("%s: rank %d renumber %v, want %d and no renumber", tc.name, got.Rank, got.Renumber, tc.want)
		}
	}
}

func TestPlaceRankRenumbersTouchingOrUnrankedNeighbours(t *testing.T) {
	got, err := domain.PlaceRank(ranked(1, 10, 2, 11), 9, 2, 0, false)
	if err != nil {
		t.Fatal(err)
	}
	want := map[int64]int64{1: 1024, 9: 2048, 2: 3072}
	if len(got.Renumber) != len(want) || got.Rank != 2048 {
		t.Fatalf("renumber = %v rank %d", got.Renumber, got.Rank)
	}
	for id, rank := range want {
		if got.Renumber[id] != rank {
			t.Errorf("task %d rank %d, want %d", id, got.Renumber[id], rank)
		}
	}

	unranked := []domain.RankedTask{{TaskID: 1, Rank: 1024, Ranked: true}, {TaskID: 2}}
	got, err = domain.PlaceRank(unranked, 9, 0, 0, false)
	if err != nil || got.Renumber == nil || got.Renumber[2] != 2048 || got.Rank != 3072 {
		t.Fatalf("placing after an unranked task = %+v, %v", got, err)
	}
}

func TestPlaceRankRejectsBadAnchors(t *testing.T) {
	list := ranked(1, 1024)
	if _, err := domain.PlaceRank(list, 9, 1, 1, false); !errors.Is(err, domain.ErrInvalidCycleMove) {
		t.Errorf("both anchors: err = %v", err)
	}
	if _, err := domain.PlaceRank(list, 9, 5, 0, false); !errors.Is(err, domain.ErrCycleRankAnchor) {
		t.Errorf("unknown anchor: err = %v", err)
	}
}

func TestResolveCycleWindow(t *testing.T) {
	day := func(s string) time.Time {
		v, _ := time.Parse("2006-01-02", s)
		return v
	}
	today := day("2026-12-07")
	start, end, err := domain.ResolveCycleWindow(domain.StartCycleInput{Today: today, LengthDays: 21}, time.Time{}, time.Time{})
	if err != nil || !start.Equal(today) || !end.Equal(day("2026-12-28")) {
		t.Fatalf("three weeks from today = %s..%s, %v", start, end, err)
	}
	_, end, _ = domain.ResolveCycleWindow(domain.StartCycleInput{Today: today}, day("2026-11-01"), day("2026-11-08"))
	if !end.Equal(day("2026-12-14")) {
		t.Fatalf("stored length kept: end = %s", end)
	}
	_, end, _ = domain.ResolveCycleWindow(domain.StartCycleInput{Today: today}, time.Time{}, time.Time{})
	if !end.Equal(day("2026-12-21")) {
		t.Fatalf("default two weeks: end = %s", end)
	}
}

func TestWorkingDaysBetweenSkipsWeekends(t *testing.T) {
	start, _ := time.Parse("2006-01-02", "2026-12-07")
	end, _ := time.Parse("2006-01-02", "2026-12-20")
	if got := len(domain.WorkingDaysBetween(start, end)); got != 10 {
		t.Fatalf("working days in two calendar weeks = %d, want 10", got)
	}
}
