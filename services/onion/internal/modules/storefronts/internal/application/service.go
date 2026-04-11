package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/storefronts/internal/domain"
	"github.com/voxire/onion/internal/modules/storefronts/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
)

const cacheTTL = 5 * time.Minute

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) List(ctx context.Context, ownerID int64) ([]domain.Storefront, error) {
	cacheKey := fmt.Sprintf("storefronts:%d", ownerID)
	return cacher.Of[[]domain.Storefront]().GetOrComputeSingle(ctx, cacheKey, func(ctx context.Context) ([]domain.Storefront, error) {
		return s.repo.List(ctx, ownerID)
	}, cacher.WithExpiry(ctx, cacheTTL))
}

func (s *Service) Create(ctx context.Context, ownerID int64, input domain.StorefrontInput) (int64, error) {
	id, err := s.repo.Create(ctx, ownerID, input)
	if err != nil {
		return 0, err
	}
	s.invalidateCache(ctx, ownerID)
	return id, nil
}

func (s *Service) Update(ctx context.Context, ownerID, id int64, input domain.StorefrontInput) error {
	err := s.repo.Update(ctx, ownerID, id, input)
	if err != nil {
		return err
	}
	s.invalidateCache(ctx, ownerID)
	return nil
}

func (s *Service) invalidateCache(ctx context.Context, ownerID int64) {
	prefix := fmt.Sprintf("storefronts:%d", ownerID)
	cacher.Of[[]domain.Storefront]().Invalidate(ctx, []string{prefix})
}
