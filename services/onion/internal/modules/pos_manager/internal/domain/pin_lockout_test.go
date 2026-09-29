package domain

import (
	"testing"
	"time"
)

// The schedule must be 2s -> 4s -> 8s, then the full lockout. The task
// description of the sibling module named that curve explicitly, and a change
// here silently weakens the manager PIN brake.
func TestPinBackoff_Schedule(t *testing.T) {
	cases := map[int]time.Duration{
		0:  0,
		1:  0,
		2:  2 * time.Second,
		3:  4 * time.Second,
		4:  8 * time.Second,
		5:  PinLockoutDuration,
		6:  PinLockoutDuration,
		99: PinLockoutDuration,
	}
	for failures, want := range cases {
		if got := PinBackoff(failures); got != want {
			t.Errorf("PinBackoff(%d) = %v, want %v", failures, got, want)
		}
	}
}

// The curve must be monotonic. An attacker who paces attempts to sit inside a
// cheaper band must not be able to buy extra guesses.
func TestPinBackoff_IsMonotonic(t *testing.T) {
	prev := time.Duration(-1)
	for failures := 0; failures <= 12; failures++ {
		got := PinBackoff(failures)
		if got < prev {
			t.Fatalf("PinBackoff(%d) = %v dropped from %v; the curve is not monotonic", failures, got, prev)
		}
		prev = got
	}
}

func TestPinRetryAt(t *testing.T) {
	last := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	if got := PinRetryAt(0, last); !got.IsZero() {
		t.Errorf("PinRetryAt(0) = %v, want the zero time (not backing off)", got)
	}
	if got := PinRetryAt(1, last); !got.IsZero() {
		t.Errorf("PinRetryAt(1) = %v, want the zero time (a first mistake must stay cheap)", got)
	}
	if got, want := PinRetryAt(3, last), last.Add(4*time.Second); !got.Equal(want) {
		t.Errorf("PinRetryAt(3) = %v, want %v", got, want)
	}
	if got, want := PinRetryAt(5, last), last.Add(PinLockoutDuration); !got.Equal(want) {
		t.Errorf("PinRetryAt(5) = %v, want %v", got, want)
	}
	// A missing timestamp must not produce a retry instant in the past that
	// silently expires the wait.
	if got := PinRetryAt(5, time.Time{}); !got.IsZero() {
		t.Errorf("PinRetryAt(5, zero) = %v, want the zero time", got)
	}
}
