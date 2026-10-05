package infrastructure

import (
	"testing"
	"time"
)

// The tool executor and the REST route both send plain YYYY-MM-DD dates. The
// repository used to accept only RFC3339, so every date-only range failed to
// parse and booking.get_calendar never returned a calendar.
func TestParseCalendarBoundary_DateOnly(t *testing.T) {
	got, err := parseCalendarBoundary("2026-06-01", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC); !got.Equal(want) {
		t.Errorf("lower bound = %v, want %v", got, want)
	}
}

// A date-only upper bound means the whole day, and the query window is
// half-open, so it has to resolve to the start of the next day.
func TestParseCalendarBoundary_DateOnlyUpperBoundIsStartOfNextDay(t *testing.T) {
	got, err := parseCalendarBoundary("2026-06-08", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := time.Date(2026, 6, 9, 0, 0, 0, 0, time.UTC); !got.Equal(want) {
		t.Errorf("upper bound = %v, want %v", got, want)
	}
}

func TestParseCalendarBoundary_RFC3339Unchanged(t *testing.T) {
	got, err := parseCalendarBoundary("2026-06-01T09:30:00Z", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := time.Date(2026, 6, 1, 9, 30, 0, 0, time.UTC); !got.Equal(want) {
		t.Errorf("timestamp = %v, want %v", got, want)
	}
}

func TestParseCalendarBoundary_RejectsGarbage(t *testing.T) {
	if _, err := parseCalendarBoundary("not-a-date", false); err == nil {
		t.Fatal("expected a parse error for an invalid value, got nil")
	}
}
