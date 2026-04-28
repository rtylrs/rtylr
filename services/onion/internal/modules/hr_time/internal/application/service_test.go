package application

import (
	"context"
	"errors"
	"testing"

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
