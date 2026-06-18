package application

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/products/internal/domain"
	"github.com/voxire/onion/internal/modules/products/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	productsCacheTTL         = 5 * time.Minute
	productsListCacheVersion = "v3"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	repo := infrastructure.NewRepository(db)

	return &Service{
		repo: repo,
	}
}

func invalidateProductScope(ctx context.Context, ownerID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ProductPrefix(ownerID))
		cache.DeleteByPattern(ctx, cacheKeys.MenuPrefix(ownerID))
	}
}

func paginateProducts(page, pageSize int64) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 50
	}
	if pageSize > 500 {
		pageSize = 500
	}
	return int(pageSize), int((page - 1) * pageSize)
}

func (s *Service) ListProducts(ctx context.Context, ownerID int64, search string, categoryID, page, pageSize int64) ([]domain.Product, bool, error) {
	search = strings.TrimSpace(search)
	limit, offset := paginateProducts(page, pageSize)
	key := fmt.Sprintf("%s|%s|s=%s|c=%d|p=%d|ps=%d",
		cacheKeys.ProductsList(ownerID), productsListCacheVersion, search, categoryID, page, pageSize)

	type pageResult struct {
		items   []domain.Product
		hasMore bool
	}

	res, err := cacher.Of[pageResult]().GetOrComputeSingle(ctx, key, func(ctx context.Context) (pageResult, error) {
		items, hasMore, err := s.repo.ListProducts(ctx, ownerID, search, int(categoryID), limit, offset)
		if items == nil {
			items = []domain.Product{}
		}
		return pageResult{items: items, hasMore: hasMore}, err
	}, cacher.WithExpiry(ctx, productsCacheTTL))
	if err != nil {
		return nil, false, err
	}
	if res.items == nil {
		res.items = []domain.Product{}
	}
	return res.items, res.hasMore, nil
}

func (s *Service) CreateProduct(ctx context.Context, ownerID int64, input domain.ProductInput) (int64, error) {
	id, err := s.repo.CreateProduct(ctx, ownerID, input)
	if err != nil {
		return 0, err
	}
	invalidateProductScope(ctx, ownerID)
	return id, nil
}

func (s *Service) UpdateProduct(ctx context.Context, ownerID, id int64, input domain.ProductInput) error {
	if err := s.repo.UpdateProduct(ctx, ownerID, id, input); err != nil {
		return err
	}
	invalidateProductScope(ctx, ownerID)
	return nil
}

func (s *Service) ListProductAttributes(ctx context.Context, ownerID, productID int64) ([]domain.ProductAttribute, error) {
	return s.repo.ListProductAttributes(ctx, ownerID, productID)
}

func (s *Service) ReplaceProductAttributes(ctx context.Context, ownerID, productID int64, attrs []domain.ProductAttributeInput) error {
	if err := s.repo.ReplaceProductAttributes(ctx, ownerID, productID, attrs); err != nil {
		return err
	}
	invalidateProductScope(ctx, ownerID)
	return nil
}

func (s *Service) DeleteProduct(ctx context.Context, ownerID, id int64) error {
	if err := s.repo.DeleteProduct(ctx, ownerID, id); err != nil {
		return err
	}
	invalidateProductScope(ctx, ownerID)
	return nil
}

func (s *Service) GetProduct(ctx context.Context, ownerID, id int64) (domain.Product, error) {
	return cacher.Of[domain.Product]().GetOrComputeSingle(ctx,
		cacheKeys.Product(ownerID, id),
		func(ctx context.Context) (domain.Product, error) {
			return s.repo.GetProduct(ctx, ownerID, id)
		},
		cacher.WithExpiry(ctx, productsCacheTTL),
	)
}

// GetProductByBarcode powers the POS + ERP scan-to-find flow. Empty
// barcode is treated as a miss so handlers can reuse the standard
// not-found error path.
func (s *Service) GetProductByBarcode(ctx context.Context, ownerID int64, barcode string) (domain.Product, error) {
	trimmed := strings.TrimSpace(barcode)
	if trimmed == "" {
		return domain.Product{}, domain.ErrProductNotFound
	}
	key := fmt.Sprintf("%s|barcode=%s", cacheKeys.ProductsList(ownerID), trimmed)
	return cacher.Of[domain.Product]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (domain.Product, error) {
			return s.repo.GetProductByBarcode(ctx, ownerID, trimmed)
		},
		cacher.WithExpiry(ctx, productsCacheTTL),
	)
}

func (s *Service) GetDashboard(ctx context.Context, ownerID int64) (domain.DashboardStats, error) {
	key := fmt.Sprintf("%s|dashboard", cacheKeys.ProductsList(ownerID))
	return cacher.Of[domain.DashboardStats]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (domain.DashboardStats, error) {
			return s.repo.GetDashboard(ctx, ownerID)
		},
		cacher.WithExpiry(ctx, productsCacheTTL),
	)
}

func (s *Service) ListReorderSuggestions(ctx context.Context, ownerID int64) ([]domain.ReorderSuggestion, error) {
	key := fmt.Sprintf("%s|reorder", cacheKeys.ProductsList(ownerID))
	return cacher.Of[[]domain.ReorderSuggestion]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) ([]domain.ReorderSuggestion, error) {
			return s.repo.ListReorderSuggestions(ctx, ownerID)
		},
		cacher.WithExpiry(ctx, productsCacheTTL),
	)
}

func (s *Service) ListExpiringItems(ctx context.Context, ownerID int64, daysAhead int) ([]domain.ExpiryItem, error) {
	if daysAhead <= 0 {
		daysAhead = 30
	}
	key := fmt.Sprintf("%s|expiring=%d", cacheKeys.ProductsList(ownerID), daysAhead)
	return cacher.Of[[]domain.ExpiryItem]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) ([]domain.ExpiryItem, error) {
			return s.repo.ListExpiringItems(ctx, ownerID, daysAhead)
		},
		cacher.WithExpiry(ctx, productsCacheTTL),
	)
}
