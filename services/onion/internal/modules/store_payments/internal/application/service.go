package application

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/store_payments/internal/domain"
	"github.com/voxire/onion/internal/modules/store_payments/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
)

const cacheTTL = 5 * time.Minute

type Service struct{ repo *infrastructure.Repository }

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func (s *Service) List(ctx context.Context, businessID int64) ([]domain.PaymentMethod, error) {
	cacheKey := "store_payments:" + strconv.FormatInt(businessID, 10)
	return cacher.Of[[]domain.PaymentMethod]().GetOrComputeSingle(ctx, cacheKey, func(ctx context.Context) ([]domain.PaymentMethod, error) {
		return s.repo.List(ctx, businessID)
	}, cacher.WithExpiry(ctx, cacheTTL))
}

func (s *Service) Create(ctx context.Context, ownerID, businessID int64, input domain.PaymentMethodInput) (int64, error) {
	id, err := s.repo.Create(ctx, ownerID, businessID, input)
	if err != nil {
		return 0, err
	}
	s.invalidateCache(ctx, businessID)
	return id, nil
}

func (s *Service) Update(ctx context.Context, businessID, id int64, input domain.PaymentMethodInput) error {
	if err := s.repo.Update(ctx, businessID, id, input); err != nil {
		return err
	}
	s.invalidateCache(ctx, businessID)
	return nil
}

func (s *Service) invalidateCache(ctx context.Context, businessID int64) {
	cacher.Of[[]domain.PaymentMethod]().Invalidate(ctx, []string{"store_payments:" + strconv.FormatInt(businessID, 10)})
}
