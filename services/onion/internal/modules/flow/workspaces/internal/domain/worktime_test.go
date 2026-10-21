package domain

import "testing"

func TestClockedMinutes(t *testing.T) {
	cases := []struct {
		name         string
		start        string
		end          string
		breakMinutes int32
		want         int32
	}{
		{"a plain day less its break", "09:00", "17:30", 30, 480},
		{"no break", "09:00", "17:00", 0, 480},
		{"a day that recorded only its lines has no clock", "", "", 0, 0},
		{"a start with no end is not half a day", "09:00", "", 0, 0},
		{"an end with no start is not half a day", "", "17:00", 0, 0},
		// An end before the start is a night shift, not a typo. Treating it as an
		// error would refuse to record hours that were genuinely worked.
		{"crossing midnight", "22:00", "06:00", 0, 480},
		{"crossing midnight with a break", "22:00", "06:30", 30, 480},
		// A break longer than the shift is nonsense, but clamping to zero is
		// better than reporting negative hours into payroll.
		{"a break longer than the day clamps to zero", "09:00", "10:00", 120, 0},
		{"malformed input is not a clock", "9:00", "17:00", 0, 0},
		{"an impossible hour is rejected", "25:00", "26:00", 0, 0},
		{"an impossible minute is rejected", "09:60", "17:00", 0, 0},
		{"non-numeric is rejected", "ab:cd", "17:00", 0, 0},
		{"midnight to midnight is a zero-length day, not 24 hours", "00:00", "00:00", 0, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ClockedMinutes(tc.start, tc.end, tc.breakMinutes)
			if got != tc.want {
				t.Fatalf("ClockedMinutes(%q, %q, %d) = %d, want %d",
					tc.start, tc.end, tc.breakMinutes, got, tc.want)
			}
		})
	}
}

func TestValidWorkMode(t *testing.T) {
	for _, mode := range []string{WorkModeOnsite, WorkModeRemote, WorkModeHybrid} {
		if !ValidWorkMode(mode) {
			t.Fatalf("ValidWorkMode(%q) = false, want true", mode)
		}
	}
	// Rejected rather than stored: the value drives reporting, so a typo would
	// quietly become its own category.
	for _, mode := range []string{"", "Onsite", "office", "wfh", "remote "} {
		if ValidWorkMode(mode) {
			t.Fatalf("ValidWorkMode(%q) = true, want false", mode)
		}
	}
}
