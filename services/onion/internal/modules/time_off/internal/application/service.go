package application

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/time_off/internal/domain"
	"github.com/voxire/onion/internal/modules/time_off/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	timeOffScope = "time_off"
	timeOffTTL   = 2 * time.Minute
)

type repository interface {
	ListTimeOffRequests(ctx context.Context, ownerID int64) ([]domain.TimeOffRequest, error)
	CreateTimeOffRequest(ctx context.Context, input domain.TimeOffInput) (int64, error)
	ApproveTimeOffRequest(ctx context.Context, ownerID int64, input domain.TimeOffApprovalInput) error
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func invalidateTimeOffScope(ctx context.Context, ownerID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(timeOffScope, ownerID))
	}
}

func (s *Service) ListTimeOffRequests(ctx context.Context, ownerID int64) ([]domain.TimeOffRequest, error) {
	return cacher.Of[[]domain.TimeOffRequest]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(timeOffScope, ownerID, "list"),
		func(ctx context.Context) ([]domain.TimeOffRequest, error) {
			return s.repo.ListTimeOffRequests(ctx, ownerID)
		},
		cacher.WithExpiry(ctx, timeOffTTL),
	)
}

func (s *Service) SubmitTimeOffRequest(ctx context.Context, input domain.TimeOffInput) (int64, error) {
	id, err := s.repo.CreateTimeOffRequest(ctx, input)
	if err != nil {
		return 0, err
	}
	invalidateTimeOffScope(ctx, input.OwnerID)
	return id, nil
}

func (s *Service) ApproveTimeOffRequest(ctx context.Context, ownerID int64, input domain.TimeOffApprovalInput) error {
	if err := s.repo.ApproveTimeOffRequest(ctx, ownerID, input); err != nil {
		return err
	}
	invalidateTimeOffScope(ctx, ownerID)
	return nil
}
