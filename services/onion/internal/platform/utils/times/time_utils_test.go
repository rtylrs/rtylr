package times

import (
	"database/sql"
	"errors"
	"testing"
	"time"
)

func TestGetTimeStringYMDHMS(t *testing.T) {
	in := time.Date(2026, 4, 28, 14, 30, 0, 0, time.UTC)
	if got := GetTimeStringYMDHMS(in); got != "2026-04-28 14:30:00" {
		t.Errorf("got %q", got)
	}
}

func TestParseWithDefault(t *testing.T) {
	def := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	parsed := ParseWithDefault(FormatWithDate, "2026-04-28", def)
	if parsed.Year() != 2026 || parsed.Month() != 4 || parsed.Day() != 28 {
		t.Errorf("expected 2026-04-28, got %v", parsed)
	}

	fallback := ParseWithDefault(FormatWithDate, "not-a-date", def)
	if !fallback.Equal(def) {
		t.Errorf("expected fallback to default, got %v", fallback)
	}
}

func TestDaysBetween(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 11, 0, 0, 0, 0, time.UTC)
	if got := DaysBetween(start, end); got != 10 {
		t.Errorf("expected 10, got %v", got)
	}
}

func TestUnixRoundTrip(t *testing.T) {
	now := time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC)
	if got := UnixFromTime(now); TimeFromUnix(got).Unix() != now.Unix() {
		t.Errorf("round trip failed for %v", now)
	}
	if got := UnixFromTime(time.Time{}); got != 0 {
		t.Errorf("zero time should give 0, got %d", got)
	}
}

func TestRangeHelpersAreOrdered(t *testing.T) {
	for _, fn := range []func() (time.Time, time.Time){
		GetDayRangeStartAndEnd,
		GetWeekRangeStartAndEnd,
		GetMonthRangeStartAndEnd,
	} {
		s, e := fn()
		if !s.Before(e) {
			t.Errorf("range start %v not before end %v", s, e)
		}
	}
}

func TestRangeHelpersUseUTCAndMondayWeekStart(t *testing.T) {
	dayStart, _ := GetDayRangeStartAndEnd()
	if dayStart.Location() != time.UTC {
		t.Fatalf("day range location = %v, want UTC", dayStart.Location())
	}
	weekStart, _ := GetWeekRangeStartAndEnd()
	if weekStart.Location() != time.UTC {
		t.Fatalf("week range location = %v, want UTC", weekStart.Location())
	}
	if weekStart.Weekday() != time.Monday {
		t.Fatalf("week starts on %v, want Monday", weekStart.Weekday())
	}
}

func TestGetDurationFromHHMMSS(t *testing.T) {
	cases := []struct {
		in   string
		want time.Duration
	}{
		{"30", 30 * time.Second},
		{"5:30", 5*time.Minute + 30*time.Second},
		{"1:30:45", time.Hour + 30*time.Minute + 45*time.Second},
	}
	for _, tc := range cases {
		got, err := GetDurationFromHHMMSS(tc.in)
		if err != nil {
			t.Errorf("unexpected err for %q: %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("GetDurationFromHHMMSS(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestGetDurationFromHHMMSS_RejectsEmptyAndOverflow(t *testing.T) {
	if _, err := GetDurationFromHHMMSS(""); !errors.Is(err, ErrDurationEmpty) {
		t.Errorf("expected ErrDurationEmpty for empty, got %v", err)
	}
	if _, err := GetDurationFromHHMMSS("   "); !errors.Is(err, ErrDurationEmpty) {
		t.Errorf("expected ErrDurationEmpty for whitespace, got %v", err)
	}
	if _, err := GetDurationFromHHMMSS("1:2:3:4"); !errors.Is(err, ErrDurationEmpty) {
		t.Errorf("expected ErrDurationEmpty for too many parts, got %v", err)
	}
}

func TestGetCountryTime(t *testing.T) {
	if _, err := GetCountryTime("LB"); err != nil {
		t.Errorf("expected LB to resolve, got %v", err)
	}
	if _, err := GetCountryTime("ZZ"); err != nil {
		t.Errorf("unknown country falls back to UTC, should not error: %v", err)
	}
}

func TestFromNullTimeAndFromTimestamp(t *testing.T) {
	now := time.Now()
	if FromNullTime(sql.NullTime{}) != nil {
		t.Error("invalid -> should return nil")
	}
	if got := FromNullTime(sql.NullTime{Time: now, Valid: true}); got == nil || !got.Equal(now) {
		t.Error("valid -> should return pointer to time")
	}
	if got := FromTimestamp(sql.NullTime{}); !got.IsZero() {
		t.Error("invalid timestamp should be zero")
	}
	if got := FromTimestamp(sql.NullTime{Time: now, Valid: true}); !got.Equal(now) {
		t.Error("valid timestamp should pass through")
	}
}
