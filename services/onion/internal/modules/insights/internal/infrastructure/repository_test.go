package infrastructure

import (
	"testing"
	"time"
)

func TestMTDRangesPreserveElapsedTimeInPriorWindow(t *testing.T) {
	now := time.Date(2026, 8, 25, 15, 42, 7, 0, time.UTC)
	thisFrom, thisTo, priorFrom, priorTo := mtdRanges(now)

	if !thisFrom.Equal(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("thisFrom = %s", thisFrom)
	}
	if !thisTo.Equal(now) {
		t.Fatalf("thisTo = %s, want %s", thisTo, now)
	}
	if !priorFrom.Equal(time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("priorFrom = %s", priorFrom)
	}
	wantPriorTo := time.Date(2026, 7, 25, 15, 42, 7, 0, time.UTC)
	if !priorTo.Equal(wantPriorTo) {
		t.Fatalf("priorTo = %s, want %s", priorTo, wantPriorTo)
	}
	if priorTo.Sub(priorFrom) != thisTo.Sub(thisFrom) {
		t.Fatalf("prior window duration = %s, want %s", priorTo.Sub(priorFrom), thisTo.Sub(thisFrom))
	}
}
