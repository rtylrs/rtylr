package application

import (
	"context"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/erp/store_products/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
)

const cacheTTL = 5 * time.Minute

type Service struct {
	repo repository
}

func NewService(_ context.Context, repo repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, businessID int64) ([]domain.StoreProduct, error) {
	cacheKey := "store_products:" + strconv.FormatInt(businessID, 10) + "|domain=v2"
	return cacher.Of[[]domain.StoreProduct]().GetOrComputeSingle(ctx, cacheKey, func(ctx context.Context) ([]domain.StoreProduct, error) {
		return s.repo.List(ctx, businessID)
	}, cacher.WithExpiry(ctx, cacheTTL))
}

func (s *Service) Create(ctx context.Context, ownerID, businessID int64, input domain.StoreProductInput) (int64, error) {
	if err := validateProduct(input); err != nil {
		return 0, err
	}
	id, err := s.repo.Create(ctx, ownerID, businessID, input)
	if err != nil {
		return 0, err
	}
	s.invalidateCache(ctx, businessID)
	return id, nil
}

func (s *Service) Update(ctx context.Context, businessID, id int64, input domain.StoreProductInput) error {
	if err := validateProduct(input); err != nil {
		return err
	}
	if err := s.repo.Update(ctx, businessID, id, input); err != nil {
		return err
	}
	s.invalidateCache(ctx, businessID)
	return nil
}

func validateProduct(input domain.StoreProductInput) error {
	if input.Price < 0 || input.CompareAtPrice < 0 {
		return domain.ErrInvalidPrice
	}
	return nil
}

func (s *Service) Delete(ctx context.Context, businessID, id int64) error {
	if err := s.repo.Delete(ctx, businessID, id); err != nil {
		return err
	}
	s.invalidateCache(ctx, businessID)
	return nil
}

func (s *Service) invalidateCache(ctx context.Context, businessID int64) {
	prefix := "store_products:" + strconv.FormatInt(businessID, 10)
	cacher.Of[[]domain.StoreProduct]().Invalidate(ctx, []string{prefix})
}

type repository interface {
	Create(ctx context.Context, ownerID, businessID int64, input domain.StoreProductInput) (int64, error)
	Delete(ctx context.Context, businessID, id int64) error
	List(ctx context.Context, businessID int64) ([]domain.StoreProduct, error)
	Update(ctx context.Context, businessID, id int64, input domain.StoreProductInput) error
}
