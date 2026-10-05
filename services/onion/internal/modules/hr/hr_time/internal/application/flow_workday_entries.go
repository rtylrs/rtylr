package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/hr/hr_time/internal/domain"
)

func (s *Service) ListFlowWorkdayEntries(ctx context.Context, ownerID, businessID, workDayID int64) ([]domain.FlowWorkdayEntry, int32, error) {
	return s.repo.ListFlowWorkdayEntries(ctx, ownerID, businessID, workDayID)
}

func (s *Service) SumFlowWorkdayEntryMinutesByDayIDs(ctx context.Context, businessID int64, workDayIDs []int64) (map[int64]int32, error) {
	return s.repo.SumFlowWorkdayEntryMinutesByDayIDs(ctx, businessID, workDayIDs)
}

func (s *Service) SumFlowWorkdayEntryMinutesByTaskIDs(ctx context.Context, businessID int64, taskIDs []int64) (map[int64]int64, error) {
	return s.repo.SumFlowWorkdayEntryMinutesByTaskIDs(ctx, businessID, taskIDs)
}

func (s *Service) SumFlowWorkdayEntryMinutesByMemberAndTaskIDs(ctx context.Context, businessID int64, taskIDs []int64) ([]domain.FlowWorkdayMemberMinutes, error) {
	return s.repo.SumFlowWorkdayEntryMinutesByMemberAndTaskIDs(ctx, businessID, taskIDs)
}

func (s *Service) CreateFlowWorkdayEntry(ctx context.Context, input domain.FlowWorkdayEntryInput) (*domain.FlowWorkdayEntry, error) {
	if input.DurationMinutes <= 0 {
		return nil, domain.ErrInvalidDuration
	}
	if _, err := time.Parse(time.DateOnly, input.Date); err != nil {
		return nil, err
	}
	if input.OwnerID <= 0 || input.BusinessID <= 0 || input.EmployeeID <= 0 || input.WorkDayID <= 0 {
		return nil, domain.ErrTimeEntryNotFound
	}
	entry, err := s.repo.CreateFlowWorkdayEntry(ctx, input)
	if err == nil {
		invalidateHRTimeScope(ctx, input.BusinessID)
	}
	return entry, err
}

func (s *Service) UpdateFlowWorkdayEntry(ctx context.Context, input domain.FlowWorkdayEntryInput) (*domain.FlowWorkdayEntry, error) {
	if input.ID <= 0 || input.OwnerID <= 0 || input.BusinessID <= 0 {
		return nil, domain.ErrTimeEntryNotFound
	}
	if input.DurationMinutes <= 0 {
		return nil, domain.ErrInvalidDuration
	}
	entry, err := s.repo.UpdateFlowWorkdayEntry(ctx, input)
	if err == nil {
		invalidateHRTimeScope(ctx, input.BusinessID)
	}
	return entry, err
}

func (s *Service) FlowWorkdayEntryOwner(ctx context.Context, id, ownerID, businessID int64) (domain.FlowWorkdayEntryOwner, bool, error) {
	return s.repo.FlowWorkdayEntryOwner(ctx, id, ownerID, businessID)
}

func (s *Service) DeleteFlowWorkdayEntry(ctx context.Context, id, ownerID, businessID int64) (bool, error) {
	deleted, err := s.repo.DeleteFlowWorkdayEntry(ctx, id, ownerID, businessID)
	if err == nil && deleted {
		invalidateHRTimeScope(ctx, businessID)
	}
	return deleted, err
}
