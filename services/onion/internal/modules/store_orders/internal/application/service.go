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

const cacheTTL = 5 * time.Minute

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
	for _, item := range input.Items {
		if item.Quantity <= 0 || item.Price < 0 || item.Total < 0 {
			return domain.ErrInvalidAmount
		}
	}
	return nil
}

func (s *Service) invalidateCache(ctx context.Context, businessID int64) {
	prefix := "store_orders:" + strconv.FormatInt(businessID, 10)
	cacher.Of[[]domain.StoreOrder]().Invalidate(ctx, []string{prefix})
}
