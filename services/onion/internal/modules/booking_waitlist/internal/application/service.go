package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/booking_waitlist/internal/domain"
	"github.com/voxire/onion/internal/modules/booking_waitlist/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
)

const cacheTTL = 5 * time.Minute

type Service struct{ repo *infrastructure.Repository }

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) List(ctx context.Context, ownerID int64) ([]domain.WaitlistEntry, error) {
	cacheKey := fmt.Sprintf("booking_waitlist:%d", ownerID)
	return cacher.Of[[]domain.WaitlistEntry]().GetOrComputeSingle(ctx, cacheKey, func(ctx context.Context) ([]domain.WaitlistEntry, error) {
		return s.repo.List(ctx, ownerID)
	}, cacher.WithExpiry(ctx, cacheTTL))
}

func (s *Service) Create(ctx context.Context, ownerID int64, input domain.WaitlistInput) (int64, error) {
	id, err := s.repo.Create(ctx, ownerID, input)
	if err != nil {
		return 0, err
	}
	s.invalidateCache(ctx, ownerID)
	return id, nil
}

func (s *Service) Update(ctx context.Context, ownerID, id int64, input domain.WaitlistInput) error {
	if err := s.repo.Update(ctx, ownerID, id, input); err != nil {
		return err
	}
	s.invalidateCache(ctx, ownerID)
	return nil
}

func (s *Service) invalidateCache(ctx context.Context, ownerID int64) {
	cacher.Of[[]domain.WaitlistEntry]().Invalidate(ctx, []string{fmt.Sprintf("booking_waitlist:%d", ownerID)})
}
