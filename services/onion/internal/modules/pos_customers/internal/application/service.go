package application

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/pos_customers/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_customers/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const posCustomersCacheTTL = 2 * time.Minute

type repository interface {
	Search(ctx context.Context, businessID int64, query string) ([]domain.Customer, error)
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

// Search piggybacks on the canonical customer namespace so any upstream
// customer mutation wipes these results too.
func (s *Service) Search(ctx context.Context, businessID int64, query string) ([]domain.Customer, error) {
	key := fmt.Sprintf("%s|pos_search=%s", cacheKeys.CustomersList(businessID), query)
	return cacher.Of[[]domain.Customer]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) ([]domain.Customer, error) {
			return s.repo.Search(ctx, businessID, query)
		},
		cacher.WithExpiry(ctx, posCustomersCacheTTL),
	)
}
