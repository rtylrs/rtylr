package infrastructure

import (
	"testing"
	"time"
)

func TestCustomerMonthBounds(t *testing.T) {
	location := time.FixedZone("UTC+3", 3*60*60)
	start, end := customerMonthBounds(time.Date(2024, time.December, 31, 23, 59, 59, 0, location), location)

	if want := time.Date(2024, time.November, 30, 21, 0, 0, 0, time.UTC); !start.Equal(want) {
		t.Fatalf("start = %s, want %s", start, want)
	}
	if want := time.Date(2024, time.December, 31, 21, 0, 0, 0, time.UTC); !end.Equal(want) {
		t.Fatalf("end = %s, want %s", end, want)
	}
}

func TestCustomerMonthBounds_UsesInstantInBusinessLocation(t *testing.T) {
	location, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Fatalf("load location: %v", err)
	}

	// This instant is still February in Los Angeles, but already March in UTC.
	start, end := customerMonthBounds(time.Date(2024, time.March, 1, 1, 0, 0, 0, time.UTC), location)

	if want := time.Date(2024, time.February, 1, 8, 0, 0, 0, time.UTC); !start.Equal(want) {
		t.Fatalf("start = %s, want %s", start, want)
	}
	if want := time.Date(2024, time.March, 1, 8, 0, 0, 0, time.UTC); !end.Equal(want) {
		t.Fatalf("end = %s, want %s", end, want)
	}
}
