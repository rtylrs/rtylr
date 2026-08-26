package application

import (
	"context"
	"database/sql"
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

func NewService(_ context.Context, db *sql.DB) *Service {
	repo := infrastructure.NewRepository(db)

	return &Service{
		repo: repo,
	}
}

func (s *Service) ListVariants(ctx context.Context, businessID, productID int64) ([]domain.Variant, error) {
	cacheKey := fmt.Sprintf("variants:%d:%d", businessID, productID)

	return cacher.Of[[]domain.Variant]().GetOrComputeSingle(ctx, cacheKey, func(ctx context.Context) ([]domain.Variant, error) {
		return s.repo.ListVariants(ctx, businessID, productID)
	}, cacher.WithExpiry(ctx, variantsCacheTTL))
}

func (s *Service) CreateVariant(ctx context.Context, businessID int64, input domain.VariantInput) (int64, error) {
	id, err := s.repo.CreateVariant(ctx, businessID, input)
	if err != nil {
		return 0, err
	}

	s.invalidateCache(ctx, businessID)

	return id, nil
}

func (s *Service) UpdateVariant(ctx context.Context, businessID, id int64, input domain.VariantInput) error {
	err := s.repo.UpdateVariant(ctx, businessID, id, input)
	if err != nil {
		return err
	}

	s.invalidateCache(ctx, businessID)

	return nil
}

func (s *Service) DeleteVariant(ctx context.Context, businessID, id int64) error {
	err := s.repo.DeleteVariant(ctx, businessID, id)
	if err != nil {
		return err
	}

	s.invalidateCache(ctx, businessID)

	return nil
}

func (s *Service) invalidateCache(ctx context.Context, businessID int64) {
	cacher.InvalidateTypedScope[[]domain.Variant](ctx, fmt.Sprintf("variants:%d:*", businessID))
}
