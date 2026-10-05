package application

import (
	"context"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/erp/brands/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
)

const (
	brandsCacheTTL = 5 * time.Minute

	defaultListLimit int32 = 20
	maxListLimit     int32 = 100
)

type BrandPage struct {
	Brands []domain.Brand
	Total  int64
	Limit  int64
	Offset int64
}

type repository interface {
	ListBrands(ctx context.Context, businessID int64) ([]domain.Brand, error)
	ListBrandsPaginated(ctx context.Context, businessID int64, limit, offset int32) ([]domain.Brand, error)
	CountBrands(ctx context.Context, businessID int64) (int64, error)
	CreateBrand(ctx context.Context, businessID int64, input domain.BrandInput) (int64, error)
	UpdateBrand(ctx context.Context, businessID, id int64, input domain.BrandInput) error
	DeleteBrand(ctx context.Context, businessID, id int64) error
}

type Service struct {
	repo repository
}

func NewService(repo repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListBrands(ctx context.Context, businessID int64) ([]domain.Brand, error) {
	cacheKey := "brands:" + strconv.FormatInt(businessID, 10) + "|domain=v2"

	return cacher.Of[[]domain.Brand]().GetOrComputeSingle(ctx, cacheKey, func(ctx context.Context) ([]domain.Brand, error) {
		return s.repo.ListBrands(ctx, businessID)
	}, cacher.WithExpiry(ctx, brandsCacheTTL))
}

// ListBrandsPaginated returns a single page of brands. A non-positive limit
// defaults to 20 and limit is capped at 100; a negative offset is clamped to
// zero.
func (s *Service) ListBrandsPaginated(ctx context.Context, businessID int64, limit, offset int32) ([]domain.Brand, error) {
	normalizedLimit, normalizedOffset := normalizeListBounds(int64(limit), int64(offset))
	return s.repo.ListBrandsPaginated(ctx, businessID, normalizedLimit, normalizedOffset)
}

func (s *Service) ListBrandsPage(ctx context.Context, businessID int64, limit, offset int64) (BrandPage, error) {
	normalizedLimit, normalizedOffset := normalizeListBounds(limit, offset)
	brands, err := s.repo.ListBrandsPaginated(ctx, businessID, normalizedLimit, normalizedOffset)
	if err != nil {
		return BrandPage{}, err
	}
	total, err := s.repo.CountBrands(ctx, businessID)
	if err != nil {
		return BrandPage{}, err
	}
	return BrandPage{
		Brands: brands,
		Total:  total,
		Limit:  int64(normalizedLimit),
		Offset: int64(normalizedOffset),
	}, nil
}

func normalizeListBounds(limit, offset int64) (int32, int32) {
	if limit <= 0 {
		limit = int64(defaultListLimit)
	}
	if limit > int64(maxListLimit) {
		limit = int64(maxListLimit)
	}
	if offset < 0 {
		offset = 0
	}
	const maxInt32 = int64(1<<31 - 1)
	if offset > maxInt32 {
		offset = maxInt32
	}
	return int32(limit), int32(offset)
}

func (s *Service) CreateBrand(ctx context.Context, businessID int64, input domain.BrandInput) (int64, error) {
	id, err := s.repo.CreateBrand(ctx, businessID, input)
	if err != nil {
		return 0, err
	}

	s.invalidateCache(ctx, businessID)

	return id, nil
}

func (s *Service) UpdateBrand(ctx context.Context, businessID, id int64, input domain.BrandInput) error {
	err := s.repo.UpdateBrand(ctx, businessID, id, input)
	if err != nil {
		return err
	}

	s.invalidateCache(ctx, businessID)

	return nil
}

func (s *Service) DeleteBrand(ctx context.Context, businessID, id int64) error {
	err := s.repo.DeleteBrand(ctx, businessID, id)
	if err != nil {
		return err
	}

	s.invalidateCache(ctx, businessID)

	return nil
}

func (s *Service) invalidateCache(ctx context.Context, businessID int64) {
	prefix := "brands:" + strconv.FormatInt(businessID, 10) + "|domain=v2"
	cacher.Of[[]domain.Brand]().Invalidate(ctx, []string{prefix})
}
