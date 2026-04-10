package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/variants/internal/domain"
	"github.com/voxire/onion/internal/modules/variants/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
)

const variantsCacheTTL = 5 * time.Minute

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	repo := infrastructure.NewRepository(ctx)

	return &Service{
		repo: repo,
	}
}

func (s *Service) ListVariants(ctx context.Context, ownerID, productID int64) ([]domain.Variant, error) {
	cacheKey := fmt.Sprintf("variants:%d:%d", ownerID, productID)

	return cacher.Of[[]domain.Variant]().GetOrComputeSingle(ctx, cacheKey, func(ctx context.Context) ([]domain.Variant, error) {
		return s.repo.ListVariants(ctx, ownerID, productID)
	}, cacher.WithExpiry(ctx, variantsCacheTTL))
}

func (s *Service) CreateVariant(ctx context.Context, ownerID int64, input domain.VariantInput) (int64, error) {
	id, err := s.repo.CreateVariant(ctx, ownerID, input)
	if err != nil {
		return 0, err
	}

	s.invalidateCache(ctx, ownerID)

	return id, nil
}

func (s *Service) UpdateVariant(ctx context.Context, ownerID, id int64, input domain.VariantInput) error {
	err := s.repo.UpdateVariant(ctx, ownerID, id, input)
	if err != nil {
		return err
	}

	s.invalidateCache(ctx, ownerID)

	return nil
}

func (s *Service) DeleteVariant(ctx context.Context, ownerID, id int64) error {
	err := s.repo.DeleteVariant(ctx, ownerID, id)
	if err != nil {
		return err
	}

	s.invalidateCache(ctx, ownerID)

	return nil
}

func (s *Service) invalidateCache(ctx context.Context, ownerID int64) {
	prefix := fmt.Sprintf("variants:%d:", ownerID)
	cacher.Of[[]domain.Variant]().Invalidate(ctx, []string{prefix})
}
