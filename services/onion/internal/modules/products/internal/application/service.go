package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/products/internal/domain"
	"github.com/voxire/onion/internal/modules/products/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
)

const productsCacheTTL = 5 * time.Minute

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	repo := infrastructure.NewRepository(ctx)

	return &Service{
		repo: repo,
	}
}

func (s *Service) ListProducts(ctx context.Context, ownerID int64, search string, categoryID, limit, offset int) ([]domain.Product, error) {
	cacheKey := fmt.Sprintf("products:%d:%s:%d:%d:%d", ownerID, search, categoryID, limit, offset)

	return cacher.Of[[]domain.Product]().GetOrComputeSingle(ctx, cacheKey, func(ctx context.Context) ([]domain.Product, error) {
		return s.repo.ListProducts(ctx, ownerID, search, categoryID, limit, offset)
	}, cacher.WithExpiry(ctx, productsCacheTTL))
}

func (s *Service) CreateProduct(ctx context.Context, ownerID int64, input domain.ProductInput) (int64, error) {
	id, err := s.repo.CreateProduct(ctx, ownerID, input)
	if err != nil {
		return 0, err
	}

	s.invalidateCache(ctx, ownerID)

	return id, nil
}

func (s *Service) UpdateProduct(ctx context.Context, ownerID, id int64, input domain.ProductInput) error {
	err := s.repo.UpdateProduct(ctx, ownerID, id, input)
	if err != nil {
		return err
	}

	s.invalidateCache(ctx, ownerID)

	return nil
}

func (s *Service) DeleteProduct(ctx context.Context, ownerID, id int64) error {
	err := s.repo.DeleteProduct(ctx, ownerID, id)
	if err != nil {
		return err
	}

	s.invalidateCache(ctx, ownerID)

	return nil
}

func (s *Service) GetProduct(ctx context.Context, ownerID, id int64) (domain.Product, error) {
	return s.repo.GetProduct(ctx, ownerID, id)
}

// GetProductByBarcode powers the POS + ERP scan-to-find flow. Empty
// barcode is treated as a miss so handlers can reuse the standard
// not-found error path.
func (s *Service) GetProductByBarcode(ctx context.Context, ownerID int64, barcode string) (domain.Product, error) {
	trimmed := strings.TrimSpace(barcode)
	if trimmed == "" {
		return domain.Product{}, domain.ErrProductNotFound
	}
	return s.repo.GetProductByBarcode(ctx, ownerID, trimmed)
}

func (s *Service) GetDashboard(ctx context.Context, ownerID int64) (domain.DashboardStats, error) {
	return s.repo.GetDashboard(ctx, ownerID)
}

func (s *Service) ListReorderSuggestions(ctx context.Context, ownerID int64) ([]domain.ReorderSuggestion, error) {
	return s.repo.ListReorderSuggestions(ctx, ownerID)
}

func (s *Service) ListExpiringItems(ctx context.Context, ownerID int64, daysAhead int) ([]domain.ExpiryItem, error) {
	if daysAhead <= 0 {
		daysAhead = 30
	}
	return s.repo.ListExpiringItems(ctx, ownerID, daysAhead)
}

func (s *Service) invalidateCache(ctx context.Context, ownerID int64) {
	prefix := fmt.Sprintf("products:%d:", ownerID)
	cacher.Of[[]domain.Product]().Invalidate(ctx, []string{prefix})
}
