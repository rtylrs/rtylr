package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/hr_time/internal/domain"
)

func TestCreateTimeEntry_RejectsNonPositiveDuration(t *testing.T) {
	s := &Service{}
	for _, dur := range []int32{0, -1, -60, -1000} {
		_, err := s.CreateTimeEntry(context.Background(), 1, domain.TimeEntryRecord{
			DurationMinutes: dur,
			EmployeeID:      1,
			Date:            "2026-04-01",
		})
		if !errors.Is(err, domain.ErrInvalidDuration) {
			t.Errorf("expected ErrInvalidDuration for %d, got %v", dur, err)
		}
	}
}

func TestUpdateTimeEntry_RejectsNonPositiveDuration(t *testing.T) {
	s := &Service{}
	for _, dur := range []int32{0, -5} {
		_, err := s.UpdateTimeEntry(context.Background(), 1, domain.TimeEntryRecord{
			DurationMinutes: dur,
		})
		if !errors.Is(err, domain.ErrInvalidDuration) {
			t.Errorf("expected ErrInvalidDuration for %d, got %v", dur, err)
		}
	}
}

func TestPreviousWeekRange(t *testing.T) {
	start, end := previousWeekRange(time.Date(2026, time.June, 23, 10, 15, 0, 0, time.UTC))

	if start != "2026-06-15" {
		t.Fatalf("expected start 2026-06-15, got %s", start)
	}
	if end != "2026-06-21" {
		t.Fatalf("expected end 2026-06-21, got %s", end)
	}
}

func TestNormalizeListTimeEntriesParams(t *testing.T) {
	got := normalizeListTimeEntriesParams(domain.ListTimeEntriesParams{Limit: 0, Offset: -5})
	if got.Limit != 20 || got.Offset != 0 {
		t.Errorf("got limit=%d offset=%d, want 20,0", got.Limit, got.Offset)
	}
	got = normalizeListTimeEntriesParams(domain.ListTimeEntriesParams{Limit: 500, Offset: 10})
	if got.Limit != 100 || got.Offset != 10 {
		t.Errorf("got limit=%d offset=%d, want 100,10", got.Limit, got.Offset)
	}
	got = normalizeListTimeEntriesParams(domain.ListTimeEntriesParams{Limit: 50, Offset: 0})
	if got.Limit != 50 || got.Offset != 0 {
		t.Errorf("got limit=%d offset=%d, want 50,0", got.Limit, got.Offset)
	}
}
