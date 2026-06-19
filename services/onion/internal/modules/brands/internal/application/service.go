package application

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/brands/internal/domain"
	"github.com/voxire/onion/internal/modules/brands/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
)

const brandsCacheTTL = 5 * time.Minute

type repository interface {
	ListBrands(ctx context.Context, businessID int64) ([]domain.Brand, error)
	CreateBrand(ctx context.Context, businessID int64, input domain.BrandInput) (int64, error)
	UpdateBrand(ctx context.Context, businessID, id int64, input domain.BrandInput) error
	DeleteBrand(ctx context.Context, businessID, id int64) error
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	repo := infrastructure.NewRepository(db)

	return &Service{
		repo: repo,
	}
}

func (s *Service) ListBrands(ctx context.Context, businessID int64) ([]domain.Brand, error) {
	cacheKey := "brands:" + strconv.FormatInt(businessID, 10)

	return cacher.Of[[]domain.Brand]().GetOrComputeSingle(ctx, cacheKey, func(ctx context.Context) ([]domain.Brand, error) {
		return s.repo.ListBrands(ctx, businessID)
	}, cacher.WithExpiry(ctx, brandsCacheTTL))
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
	prefix := "brands:" + strconv.FormatInt(businessID, 10)
	cacher.Of[[]domain.Brand]().Invalidate(ctx, []string{prefix})
}
