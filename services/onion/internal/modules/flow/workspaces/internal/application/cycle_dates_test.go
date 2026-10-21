package application_test

import (
	"errors"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/application"
)

// The v1 cycle handler grew these guards in the security/infra audit (4514d877).
// v2 had none, and its time.Parse discarded the error, so an unparseable or empty
// date became the zero time and a year-0001 row was written. Deleting the v1
// handler without porting them would have quietly undone that fix on the only
// remaining path.

func TestNormalizeCycleDates_RejectsEndBeforeStart(t *testing.T) {
	_, _, err := application.NormalizeCycleDates("2026-08-20", "2026-08-01")

	if !errors.Is(err, application.ErrInvalidCycleDates) {
		t.Fatalf("end before start: want ErrInvalidCycleDates, got %v", err)
	}
}

// Rejected rather than silently zeroed. Treating "not-a-date" as the zero time
// is exactly how the year-0001 rows were written.
//
// The end dates here are deliberately FAR IN THE FUTURE. An earlier version of
// this test used a past end date, and it passed even with the parse error
// swallowed again: a zeroed start defaults to today, today was after that end,
// and the end-before-start check rejected it for the wrong reason. Mutation
// probing caught that; the dates are chosen so only the parse guard can fail it.
func TestNormalizeCycleDates_RejectsUnparseable(t *testing.T) {
	for _, tc := range []struct{ name, start, end string }{
		{"bad start", "not-a-date", "2099-12-01"},
		{"bad end", "2099-01-01", "31-12-2099"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, err := application.NormalizeCycleDates(tc.start, tc.end); !errors.Is(err, application.ErrInvalidCycleDates) {
				t.Fatalf("want ErrInvalidCycleDates, got %v", err)
			}
		})
	}
}

// An omitted start is defaulted, not rejected, matching v1: a caller supplying
// only an end date gets today.
func TestNormalizeCycleDates_DefaultsOmittedStartToToday(t *testing.T) {
	start, end, err := application.NormalizeCycleDates("", "2026-12-01")
	if err != nil {
		t.Fatalf("omitted start should default, got %v", err)
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)
	if !start.Equal(today) {
		t.Fatalf("start: got %v want %v", start, today)
	}
	if end.Year() != 2026 || end.Month() != time.December {
		t.Fatalf("end should be the supplied date, got %v", end)
	}
}

// An omitted end becomes start + 14 days, which is v1's behaviour.
func TestNormalizeCycleDates_DefaultsOmittedEndToTwoWeeks(t *testing.T) {
	start, end, err := application.NormalizeCycleDates("2026-08-01", "")
	if err != nil {
		t.Fatalf("omitted end should default, got %v", err)
	}

	if want := start.AddDate(0, 0, 14); !end.Equal(want) {
		t.Fatalf("end: got %v want %v", end, want)
	}
}

// Both omitted must still produce a usable, ordered pair rather than two zero
// times, which is the case that wrote year-0001 rows.
func TestNormalizeCycleDates_BothOmittedIsUsable(t *testing.T) {
	start, end, err := application.NormalizeCycleDates("", "")
	if err != nil {
		t.Fatalf("both omitted should default, got %v", err)
	}

	if start.IsZero() || end.IsZero() {
		t.Fatalf("neither date may be the zero time: start=%v end=%v", start, end)
	}
	if !end.After(start) {
		t.Fatalf("end must follow start: start=%v end=%v", start, end)
	}
}
