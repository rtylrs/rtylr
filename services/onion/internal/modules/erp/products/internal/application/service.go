package application

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/erp/products/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheInvalidate "github.com/voxire/onion/internal/platform/cache/invalidate"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	productsCacheTTL = 5 * time.Minute
)

type Service struct {
	repo repository
}

type repository interface {
	ListProducts(context.Context, int64, string, int, int, int) ([]domain.Product, int64, bool, error)
	CreateProduct(context.Context, int64, domain.ProductInput) (int64, error)
	GetProduct(context.Context, int64, int64) (domain.Product, error)
	UpdateProduct(context.Context, int64, int64, domain.ProductInput) error
	DeleteProduct(context.Context, int64, int64) error
	ListProductAttributes(context.Context, int64, int64) ([]domain.ProductAttribute, error)
	ReplaceProductAttributes(context.Context, int64, int64, []domain.ProductAttributeInput) error
	GetProductByBarcode(context.Context, int64, string) (domain.Product, error)
	GetDashboard(context.Context, int64) (domain.DashboardStats, error)
	ListReorderSuggestions(context.Context, int64) (domain.ReorderSuggestionsResult, error)
	ListExpiringItems(context.Context, int64, int) ([]domain.ExpiryItem, error)
}

type productListPage struct {
	Items   []domain.Product `json:"items"`
	Total   int64            `json:"total"`
	HasMore bool             `json:"has_more"`
}

func NewService(repo repository) *Service {
	return &Service{repo: repo}
}

func invalidateProductScope(ctx context.Context, ownerID int64) {
	invalidationCtx := context.WithoutCancel(ctx)
	cacher.InvalidateScope(invalidationCtx,
		cacheKeys.ProductPrefix(ownerID),
		cacheKeys.MenuPrefix(ownerID),
	)

	owner := strconv.FormatInt(ownerID, 10)
	cacheInvalidate.InvalidateAsync(invalidationCtx, cacheKeys.InvalidateProduct, owner)
	cacheInvalidate.InvalidateAsync(invalidationCtx, cacheKeys.InvalidateMenu, owner)
}

func normalizeProductPagination(page, pageSize int64) (int64, int64, int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 50
	}
	if pageSize > 500 {
		pageSize = 500
	}
	return page, pageSize, int(pageSize), int((page - 1) * pageSize)
}

func paginateProducts(page, pageSize int64) (int, int) {
	_, _, limit, offset := normalizeProductPagination(page, pageSize)
	return limit, offset
}

func productListCacheKey(ownerID int64, search string, categoryID, page, pageSize int64) string {
	normalizedSearch := strings.ToLower(strings.TrimSpace(search))
	return fmt.Sprintf("%s|search=%s|category=%d|page=%d|page_size=%d|domain=v2",
		cacheKeys.ProductsList(ownerID),
		url.QueryEscape(normalizedSearch),
		categoryID,
		page,
		pageSize,
	)
}

func (s *Service) ListProducts(ctx context.Context, ownerID int64, search string, categoryID, page, pageSize int64) ([]domain.Product, int64, bool, error) {
	search = strings.TrimSpace(search)
	page, pageSize, limit, offset := normalizeProductPagination(page, pageSize)

	result, err := cacher.Of[productListPage]().GetOrComputeSingle(ctx,
		productListCacheKey(ownerID, search, categoryID, page, pageSize),
		func(ctx context.Context) (productListPage, error) {
			items, total, hasMore, err := s.repo.ListProducts(ctx, ownerID, search, int(categoryID), limit, offset)
			if err != nil {
				return productListPage{}, err
			}
			if items == nil {
				items = []domain.Product{}
			}
			return productListPage{Items: items, Total: total, HasMore: hasMore}, nil
		},
		cacher.WithExpiry(ctx, productsCacheTTL),
	)
	if err != nil {
		return nil, 0, false, err
	}
	if result.Items == nil {
		result.Items = []domain.Product{}
	}
	return result.Items, result.Total, result.HasMore, nil
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
		cacheKeys.Product(ownerID, id)+"|domain=v2",
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
	key := fmt.Sprintf("%s|barcode=%s|domain=v2", cacheKeys.ProductsList(ownerID), trimmed)
	return cacher.Of[domain.Product]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (domain.Product, error) {
			return s.repo.GetProductByBarcode(ctx, ownerID, trimmed)
		},
		cacher.WithExpiry(ctx, productsCacheTTL),
	)
}

func (s *Service) GetDashboard(ctx context.Context, ownerID int64) (domain.DashboardStats, error) {
	key := fmt.Sprintf("%s|dashboard|domain=v2", cacheKeys.ProductsList(ownerID))
	return cacher.Of[domain.DashboardStats]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (domain.DashboardStats, error) {
			return s.repo.GetDashboard(ctx, ownerID)
		},
		cacher.WithExpiry(ctx, productsCacheTTL),
	)
}

func (s *Service) ListReorderSuggestions(ctx context.Context, ownerID int64) (domain.ReorderSuggestionsResult, error) {
	key := fmt.Sprintf("%s|reorder|domain=v2", cacheKeys.ProductsList(ownerID))
	return cacher.Of[domain.ReorderSuggestionsResult]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (domain.ReorderSuggestionsResult, error) {
			return s.repo.ListReorderSuggestions(ctx, ownerID)
		},
		cacher.WithExpiry(ctx, productsCacheTTL),
	)
}

func (s *Service) ListExpiringItems(ctx context.Context, ownerID int64, daysAhead int) ([]domain.ExpiryItem, error) {
	if daysAhead <= 0 {
		daysAhead = 30
	}
	key := fmt.Sprintf("%s|expiring=%d|domain=v2", cacheKeys.ProductsList(ownerID), daysAhead)
	return cacher.Of[[]domain.ExpiryItem]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) ([]domain.ExpiryItem, error) {
			return s.repo.ListExpiringItems(ctx, ownerID, daysAhead)
		},
		cacher.WithExpiry(ctx, productsCacheTTL),
	)
}
