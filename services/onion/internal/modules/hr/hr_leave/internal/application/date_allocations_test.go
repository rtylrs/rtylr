package application

import (
	"testing"
	"time"
)

func TestLeaveDayAllocationsSplitsCalendarYears(t *testing.T) {
	start := time.Date(2026, time.December, 30, 0, 0, 0, 0, time.UTC)
	end := time.Date(2027, time.January, 2, 0, 0, 0, 0, time.UTC)

	allocations := leaveDayAllocations(start, end)
	if len(allocations) != 2 {
		t.Fatalf("allocation count = %d, want 2", len(allocations))
	}
	if allocations[0].Year != 2026 || allocations[0].Days != 2 {
		t.Fatalf("first allocation = %+v, want 2026/2", allocations[0])
	}
	if allocations[1].Year != 2027 || allocations[1].Days != 2 {
		t.Fatalf("second allocation = %+v, want 2027/2", allocations[1])
	}
	if totalAllocatedDays(allocations) != 4 {
		t.Fatalf("total allocated days = %v, want 4", totalAllocatedDays(allocations))
	}
}

func TestLeaveDayAllocationsKeepsSingleYearRequestsTogether(t *testing.T) {
	start := time.Date(2027, time.March, 10, 0, 0, 0, 0, time.UTC)
	end := time.Date(2027, time.March, 12, 0, 0, 0, 0, time.UTC)

	allocations := leaveDayAllocations(start, end)
	if len(allocations) != 1 || allocations[0].Year != 2027 || allocations[0].Days != 3 {
		t.Fatalf("allocations = %+v, want one 2027 allocation for 3 days", allocations)
	}
}
