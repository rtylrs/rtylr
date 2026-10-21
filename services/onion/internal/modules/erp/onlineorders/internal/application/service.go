package application

import (
	"context"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/erp/onlineorders/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	cacheTTL = 5 * time.Minute

	defaultListLimit int32 = 20
	maxListLimit     int32 = 100
	maxListOffset    int32 = 10000
)

type Service struct{ repo repository }

func NewService(_ context.Context, repo repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, businessID int64) ([]domain.StoreOrder, error) {
	return cacher.Of[[]domain.StoreOrder]().GetOrComputeSingle(ctx, storeOrdersCacheKey(businessID), func(ctx context.Context) ([]domain.StoreOrder, error) {
		return s.repo.List(ctx, businessID)
	}, cacher.WithExpiry(ctx, cacheTTL))
}

// ListPaginated returns a single page of store orders plus the tenant's total
// count. A non-positive limit defaults to 20 and limit is capped at 100; a
// negative offset is clamped to zero.
func (s *Service) ListPaginated(ctx context.Context, businessID int64, limit, offset int32) (domain.StoreOrderPage, error) {
	limit, offset = normalizeListPagination(limit, offset)

	orders, err := s.repo.ListPaginated(ctx, businessID, limit, offset)
	if err != nil {
		return domain.StoreOrderPage{}, err
	}

	total, err := s.repo.Count(ctx, businessID)
	if err != nil {
		return domain.StoreOrderPage{}, err
	}

	return domain.StoreOrderPage{
		Orders: orders,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

func normalizeListPagination(limit, offset int32) (int32, int32) {
	if limit <= 0 {
		limit = defaultListLimit
	}
	if limit > maxListLimit {
		limit = maxListLimit
	}
	if offset < 0 {
		offset = 0
	}
	if offset > maxListOffset {
		offset = maxListOffset
	}
	return limit, offset
}

func (s *Service) GetDetail(ctx context.Context, businessID, orderID int64) (domain.StoreOrderDetail, error) {
	return s.repo.GetDetail(ctx, businessID, orderID)
}

func (s *Service) Create(ctx context.Context, ownerID, businessID int64, input domain.StoreOrderInput) (int64, error) {
	if err := validateOrder(input); err != nil {
		return 0, err
	}
	id, err := s.repo.Create(ctx, ownerID, businessID, input)
	if err != nil {
		return 0, err
	}
	s.invalidateCache(ctx, businessID)
	return id, nil
}

func (s *Service) Update(ctx context.Context, businessID, id int64, input domain.StoreOrderInput) error {
	// Updates change only order metadata and status. Items and monetary totals
	// are create-time, catalog-derived values and are not persisted by Update,
	// so applying create-order amount validation here rejects valid partial
	// updates (for example, an update that echoes the existing total).
	if err := s.repo.Update(ctx, businessID, id, input); err != nil {
		return err
	}
	s.invalidateCache(ctx, businessID)
	return nil
}

func validateOrder(input domain.StoreOrderInput) error {
	if input.Subtotal < 0 || input.Tax < 0 || input.ShippingCost < 0 || input.Total < 0 {
		return domain.ErrInvalidAmount
	}
	if len(input.Items) == 0 && (input.Subtotal != 0 || input.Tax != 0 || input.ShippingCost != 0 || input.Total != 0) {
		return domain.ErrInvalidAmount
	}
	for _, item := range input.Items {
		if item.ProductID <= 0 || item.Quantity <= 0 || item.Price < 0 || item.Total < 0 {
			return domain.ErrInvalidAmount
		}
	}
	return nil
}

func storeOrdersCacheKey(businessID int64) string {
	return "store_orders:" + strconv.FormatInt(businessID, 10) + "|domain=v2"
}

func (s *Service) invalidateCache(ctx context.Context, businessID int64) {
	// Invalidate deletes exact keys, so it must be the key List caches under.
	cacher.Of[[]domain.StoreOrder]().Invalidate(ctx, []string{storeOrdersCacheKey(businessID)})
	// Confirming, cancelling or refunding an order moves stock.
	cacher.InvalidateScope(ctx, cacheKeys.ProductPrefix(businessID))
}

type repository interface {
	Count(ctx context.Context, businessID int64) (int64, error)
	Create(ctx context.Context, ownerID, businessID int64, input domain.StoreOrderInput) (int64, error)
	GetDetail(ctx context.Context, businessID, orderID int64) (domain.StoreOrderDetail, error)
	List(ctx context.Context, businessID int64) ([]domain.StoreOrder, error)
	ListPaginated(ctx context.Context, businessID int64, limit, offset int32) ([]domain.StoreOrder, error)
	Update(ctx context.Context, businessID, id int64, input domain.StoreOrderInput) error
}
