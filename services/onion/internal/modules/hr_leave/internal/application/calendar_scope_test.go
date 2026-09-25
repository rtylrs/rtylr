package application

import (
	"testing"

	"github.com/voxire/onion/internal/modules/hr_leave/internal/domain"
)

func TestLeaveCalendarCacheKeySeparatesLocations(t *testing.T) {
	base := domain.LeaveCalendarParams{DateFrom: "2026-01-01", DateTo: "2026-01-31"}
	one := base
	one.LocationID = 10
	two := base
	two.LocationID = 20
	if leaveCalendarCacheKey(1001, one) == leaveCalendarCacheKey(1001, two) {
		t.Fatal("different location filters must not share a calendar cache entry")
	}
	if leaveCalendarCacheKey(1001, one) == leaveCalendarCacheKey(2002, one) {
		t.Fatal("different businesses must not share a calendar cache entry")
	}
}

func TestNormalizeLeaveCalendarParamsRejectsInvalidRangeAndLocation(t *testing.T) {
	if _, err := normalizeLeaveCalendarParams(domain.LeaveCalendarParams{LocationID: -1, DateFrom: "2026-01-01", DateTo: "2026-01-02"}); err == nil {
		t.Fatal("negative location should be rejected")
	}
	if _, err := normalizeLeaveCalendarParams(domain.LeaveCalendarParams{DateFrom: "2026-02-01", DateTo: "2026-01-01"}); err == nil {
		t.Fatal("reversed date range should be rejected")
	}
}
