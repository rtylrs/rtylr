package application

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/bills/internal/domain"
	"github.com/voxire/onion/internal/modules/bills/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	billsScope = "bills"
	billsTTL   = 5 * time.Minute
)

type repository interface {
	List(ctx context.Context, businessID int64) ([]domain.Bill, error)
	Get(ctx context.Context, businessID, id int64) (domain.Bill, error)
	Delete(ctx context.Context, businessID, id int64) error
	Create(ctx context.Context, businessID int64, input domain.BillInput) (int64, error)
	Update(ctx context.Context, businessID, id int64, input domain.BillInput) error
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func invalidateBillsScope(ctx context.Context, businessID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(billsScope, businessID))
	}
}

func (s *Service) List(ctx context.Context, businessID int64) ([]domain.Bill, error) {
	return cacher.Of[[]domain.Bill]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(billsScope, businessID, "list"),
		func(ctx context.Context) ([]domain.Bill, error) {
			return s.repo.List(ctx, businessID)
		},
		cacher.WithExpiry(ctx, billsTTL),
	)
}

func (s *Service) GetBill(ctx context.Context, businessID, id int64) (domain.Bill, error) {
	key := cacheKeys.Scoped(billsScope, businessID, "id="+strconv.FormatInt(id, 10))
	return cacher.Of[domain.Bill]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (domain.Bill, error) {
			return s.repo.Get(ctx, businessID, id)
		},
		cacher.WithExpiry(ctx, billsTTL),
	)
}

func (s *Service) DeleteBill(ctx context.Context, businessID, id int64) error {
	if err := s.repo.Delete(ctx, businessID, id); err != nil {
		return err
	}
	invalidateBillsScope(ctx, businessID)
	return nil
}

func (s *Service) Create(ctx context.Context, businessID int64, input domain.BillInput) (int64, error) {
	id, err := s.repo.Create(ctx, businessID, input)
	if err != nil {
		return 0, err
	}
	invalidateBillsScope(ctx, businessID)
	return id, nil
}

func (s *Service) Update(ctx context.Context, businessID, id int64, input domain.BillInput) error {
	if err := s.repo.Update(ctx, businessID, id, input); err != nil {
		return err
	}
	invalidateBillsScope(ctx, businessID)
	return nil
}
