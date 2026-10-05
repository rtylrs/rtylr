package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/erp/variants/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
)

const variantsCacheTTL = 5 * time.Minute

type VariantPage struct {
	Items   []domain.Variant
	HasMore bool
	Total   int64
	Limit   int
	Offset  int
}

type repository interface {
	ListVariantsPage(ctx context.Context, businessID, productID int64, limit, offset int) ([]domain.Variant, bool, int64, error)
	CreateVariant(ctx context.Context, businessID int64, input domain.VariantInput) (int64, error)
	UpdateVariant(ctx context.Context, businessID, id int64, input domain.VariantInput) error
	DeleteVariant(ctx context.Context, businessID, id int64) error
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, repo repository) *Service {
	return &Service{repo: repo}
}

func paginateVariants(offset, limit int64) (int, int) {
	if offset < 0 {
		offset = 0
	}
	if limit < 1 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}
	return int(limit), int(offset)
}

func (s *Service) ListVariantsPage(ctx context.Context, businessID, productID, limit, offset int64) (VariantPage, error) {
	pageSize, pageOffset := paginateVariants(offset, limit)
	cacheKey := fmt.Sprintf("variants:%d:p=%d:limit=%d:offset=%d", businessID, productID, pageSize, pageOffset) + "|domain=v2"
	page, err := cacher.Of[VariantPage]().GetOrComputeSingle(ctx, cacheKey, func(ctx context.Context) (VariantPage, error) {
		items, hasMore, total, listErr := s.repo.ListVariantsPage(ctx, businessID, productID, pageSize, pageOffset)
		return VariantPage{Items: items, HasMore: hasMore, Total: total, Limit: pageSize, Offset: pageOffset}, listErr
	}, cacher.WithExpiry(ctx, variantsCacheTTL))
	if err != nil {
		return VariantPage{}, err
	}
	if page.Items == nil {
		page.Items = []domain.Variant{}
	}
	return page, nil
}

func (s *Service) ListVariants(ctx context.Context, businessID, productID int64) ([]domain.Variant, error) {
	page, err := s.ListVariantsPage(ctx, businessID, productID, 100, 0)
	return page.Items, err
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
	if err := s.repo.UpdateVariant(ctx, businessID, id, input); err != nil {
		return err
	}
	s.invalidateCache(ctx, businessID)
	return nil
}

func (s *Service) DeleteVariant(ctx context.Context, businessID, id int64) error {
	if err := s.repo.DeleteVariant(ctx, businessID, id); err != nil {
		return err
	}
	s.invalidateCache(ctx, businessID)
	return nil
}

func (s *Service) invalidateCache(ctx context.Context, businessID int64) {
	cacher.InvalidateTypedScope[VariantPage](ctx, fmt.Sprintf("variants:%d:*", businessID))
}
