package application

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/store_orders/internal/domain"
	"github.com/voxire/onion/internal/modules/store_orders/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
)

const (
	cacheTTL = 5 * time.Minute

	defaultListLimit int32 = 20
	maxListLimit     int32 = 100
	maxListOffset    int32 = 10000
)

type Service struct{ repo *infrastructure.Repository }

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func (s *Service) List(ctx context.Context, businessID int64) ([]domain.StoreOrder, error) {
	cacheKey := "store_orders:" + strconv.FormatInt(businessID, 10)
	return cacher.Of[[]domain.StoreOrder]().GetOrComputeSingle(ctx, cacheKey, func(ctx context.Context) ([]domain.StoreOrder, error) {
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
	if err := validateOrder(input); err != nil {
		return err
	}
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

func (s *Service) invalidateCache(ctx context.Context, businessID int64) {
	prefix := "store_orders:" + strconv.FormatInt(businessID, 10)
	cacher.Of[[]domain.StoreOrder]().Invalidate(ctx, []string{prefix})
}
