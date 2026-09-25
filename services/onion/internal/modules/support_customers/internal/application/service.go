package application

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/support_customers/internal/domain"
	"github.com/voxire/onion/internal/modules/support_customers/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	supportCustomersScope = "support_customers"
	supportCustomersTTL   = 2 * time.Minute
)

type repository interface {
	List(ctx context.Context, businessID int64, page, pageSize int) ([]domain.Customer, int, error)
	GetByID(ctx context.Context, businessID, id int64) (*domain.Customer, error)
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{
		repo: infrastructure.NewRepository(db),
	}
}

type supportCustomersListResult struct {
	Items []domain.Customer
	Total int
}

func (s *Service) List(ctx context.Context, businessID int64, page, pageSize int) ([]domain.Customer, int, error) {
	key := cacheKeys.Scoped(supportCustomersScope, businessID, fmt.Sprintf("list|p=%d|ps=%d", page, pageSize))
	res, err := cacher.Of[supportCustomersListResult]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (supportCustomersListResult, error) {
			items, total, err := s.repo.List(ctx, businessID, page, pageSize)
			return supportCustomersListResult{Items: items, Total: total}, err
		},
		cacher.WithExpiry(ctx, supportCustomersTTL),
	)
	return res.Items, res.Total, err
}

func (s *Service) GetByID(ctx context.Context, businessID, id int64) (*domain.Customer, error) {
	key := cacheKeys.Scoped(supportCustomersScope, businessID, fmt.Sprintf("id=%d", id))
	return cacher.Of[*domain.Customer]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (*domain.Customer, error) {
			return s.repo.GetByID(ctx, businessID, id)
		},
		cacher.WithExpiry(ctx, supportCustomersTTL),
	)
}
