package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/store_categories/internal/domain"
	"github.com/voxire/onion/internal/modules/store_categories/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
)

const cacheTTL = 5 * time.Minute

type Service struct{ repo *infrastructure.Repository }

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) List(ctx context.Context, ownerID int64) ([]domain.StoreCategory, error) {
	cacheKey := fmt.Sprintf("store_categories:%d", ownerID)
	return cacher.Of[[]domain.StoreCategory]().GetOrComputeSingle(ctx, cacheKey, func(ctx context.Context) ([]domain.StoreCategory, error) {
		return s.repo.List(ctx, ownerID)
	}, cacher.WithExpiry(ctx, cacheTTL))
}

func (s *Service) Create(ctx context.Context, ownerID int64, input domain.StoreCategoryInput) (int64, error) {
	id, err := s.repo.Create(ctx, ownerID, input)
	if err != nil {
		return 0, err
	}
	s.invalidateCache(ctx, ownerID)
	return id, nil
}

func (s *Service) Update(ctx context.Context, ownerID, id int64, input domain.StoreCategoryInput) error {
	if err := s.repo.Update(ctx, ownerID, id, input); err != nil {
		return err
	}
	s.invalidateCache(ctx, ownerID)
	return nil
}

func (s *Service) Delete(ctx context.Context, ownerID, id int64) error {
	if err := s.repo.Delete(ctx, ownerID, id); err != nil {
		return err
	}
	s.invalidateCache(ctx, ownerID)
	return nil
}

func (s *Service) invalidateCache(ctx context.Context, ownerID int64) {
	prefix := fmt.Sprintf("store_categories:%d", ownerID)
	cacher.Of[[]domain.StoreCategory]().Invalidate(ctx, []string{prefix})
}
