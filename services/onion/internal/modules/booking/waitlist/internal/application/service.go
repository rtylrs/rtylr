package application

import (
	"context"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/booking/waitlist/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
)

const cacheTTL = 5 * time.Minute

type Service struct{ repo repository }

func NewService(_ context.Context, repo repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, businessID int64) ([]domain.WaitlistEntry, error) {
	cacheKey := "booking_waitlist:" + strconv.FormatInt(businessID, 10) + "|domain=v2"
	return cacher.Of[[]domain.WaitlistEntry]().GetOrComputeSingle(ctx, cacheKey, func(ctx context.Context) ([]domain.WaitlistEntry, error) {
		return s.repo.List(ctx, businessID)
	}, cacher.WithExpiry(ctx, cacheTTL))
}

func (s *Service) Create(ctx context.Context, businessID int64, input domain.WaitlistInput) (int64, error) {
	if input.ServiceID != 0 {
		if err := s.repo.ServiceBelongsToOwner(ctx, businessID, input.ServiceID); err != nil {
			return 0, err
		}
	}
	id, err := s.repo.Create(ctx, businessID, input)
	if err != nil {
		return 0, err
	}
	s.invalidateCache(ctx, businessID)
	return id, nil
}

func (s *Service) Update(ctx context.Context, businessID, id int64, input domain.WaitlistInput) error {
	if input.ServiceID != 0 {
		if err := s.repo.ServiceBelongsToOwner(ctx, businessID, input.ServiceID); err != nil {
			return err
		}
	}
	if err := s.repo.Update(ctx, businessID, id, input); err != nil {
		return err
	}
	s.invalidateCache(ctx, businessID)
	return nil
}

func (s *Service) invalidateCache(ctx context.Context, businessID int64) {
	cacher.Of[[]domain.WaitlistEntry]().Invalidate(ctx, []string{"booking_waitlist:" + strconv.FormatInt(businessID, 10) + "|domain=v2"})
}

type repository interface {
	Create(ctx context.Context, businessID int64, input domain.WaitlistInput) (int64, error)
	List(ctx context.Context, businessID int64) ([]domain.WaitlistEntry, error)
	ServiceBelongsToOwner(ctx context.Context, businessID, serviceID int64) error
	Update(ctx context.Context, businessID, id int64, input domain.WaitlistInput) error
}
