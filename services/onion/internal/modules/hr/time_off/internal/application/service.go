package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/hr/time_off/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	timeOffScope = "time_off"
	timeOffTTL   = 2 * time.Minute
)

type repository interface {
	ListTimeOffRequests(ctx context.Context, businessID int64) ([]domain.TimeOffRequest, error)
	ListTimeOffRequestsForUser(ctx context.Context, businessID, userID int64) ([]domain.TimeOffRequest, error)
	StaffBelongsToUser(ctx context.Context, businessID, userID, staffID int64) (bool, error)
	CreateTimeOffRequest(ctx context.Context, input domain.TimeOffInput) (int64, error)
	ApproveTimeOffRequest(ctx context.Context, businessID int64, input domain.TimeOffApprovalInput) error
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, repo repository) *Service {
	return &Service{repo: repo}
}

func invalidateTimeOffScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(timeOffScope, businessID))
}

func (s *Service) ListTimeOffRequests(ctx context.Context, businessID int64) ([]domain.TimeOffRequest, error) {
	return cacher.Of[[]domain.TimeOffRequest]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(timeOffScope, businessID, "list")+"|domain=v2",
		func(ctx context.Context) ([]domain.TimeOffRequest, error) {
			return s.repo.ListTimeOffRequests(ctx, businessID)
		},
		cacher.WithExpiry(ctx, timeOffTTL),
	)
}

// ListTimeOffRequestsForUser lists only the requests of the user's own staff
// records, for callers who may not see colleagues' leave.
func (s *Service) ListTimeOffRequestsForUser(ctx context.Context, businessID, userID int64) ([]domain.TimeOffRequest, error) {
	return cacher.Of[[]domain.TimeOffRequest]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(timeOffScope, businessID, fmt.Sprintf("list|user=%d", userID))+"|domain=v2",
		func(ctx context.Context) ([]domain.TimeOffRequest, error) {
			return s.repo.ListTimeOffRequestsForUser(ctx, businessID, userID)
		},
		cacher.WithExpiry(ctx, timeOffTTL),
	)
}

// StaffBelongsToUser reports whether staffID is one of the user's own staff records.
func (s *Service) StaffBelongsToUser(ctx context.Context, businessID, userID, staffID int64) (bool, error) {
	return s.repo.StaffBelongsToUser(ctx, businessID, userID, staffID)
}

func (s *Service) SubmitTimeOffRequest(ctx context.Context, input domain.TimeOffInput) (int64, error) {
	id, err := s.repo.CreateTimeOffRequest(ctx, input)
	if err != nil {
		return 0, err
	}
	invalidateTimeOffScope(ctx, input.BusinessID)
	return id, nil
}

func (s *Service) ApproveTimeOffRequest(ctx context.Context, businessID int64, input domain.TimeOffApprovalInput) error {
	if err := s.repo.ApproveTimeOffRequest(ctx, businessID, input); err != nil {
		return err
	}
	invalidateTimeOffScope(ctx, businessID)
	return nil
}
