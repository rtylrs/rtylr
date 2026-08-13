package infrastructure

import (
	"testing"
	"time"
)

func TestCustomerMonthBounds(t *testing.T) {
	start, end := customerMonthBounds(time.Date(2024, time.December, 31, 23, 59, 59, 0, time.FixedZone("UTC+3", 3*60*60)))

	if want := time.Date(2024, time.December, 1, 0, 0, 0, 0, time.UTC); !start.Equal(want) {
		t.Fatalf("start = %s, want %s", start, want)
	}
	if want := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC); !end.Equal(want) {
		t.Fatalf("end = %s, want %s", end, want)
	}
}
