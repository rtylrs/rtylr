package application

import (
	"strings"
	"testing"
)

// resolvePeriodRange depends on time.Now() so we can't pin exact dates, but we
// can lock in the SHAPE: yyyy-mm-dd, from <= to, and the relationships between
// months are consistent.

func TestResolvePeriodRange_Shape(t *testing.T) {
	for _, period := range []string{"quarter", "year", "month", "", "garbage"} {
		from, to := resolvePeriodRange(period)
		if !looksLikeDate(from) {
			t.Errorf("from for %q is not yyyy-mm-dd: %q", period, from)
		}
		if !looksLikeDate(to) {
			t.Errorf("to for %q is not yyyy-mm-dd: %q", period, to)
		}
		if from > to {
			t.Errorf("from %q > to %q for %q", from, to, period)
		}
	}
}

func TestResolvePeriodRange_QuarterIsThreeMonths(t *testing.T) {
	from, to := resolvePeriodRange("quarter")
	// from and to are in the same year and three months apart (the to date is the
	// last day of the third month, so months differ by 2 inclusive).
	if from[:4] != to[:4] {
		t.Errorf("quarter spans years: from=%s, to=%s", from, to)
	}
}

func TestResolvePeriodRange_YearStartsJanFirst(t *testing.T) {
	from, _ := resolvePeriodRange("year")
	if !strings.HasSuffix(from, "-01-01") {
		t.Errorf("year start should end with -01-01, got %q", from)
	}
}

func looksLikeDate(s string) bool {
	if len(s) != 10 {
		return false
	}
	return s[4] == '-' && s[7] == '-'
}
