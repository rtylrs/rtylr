package application

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/tax/internal/domain"
	"github.com/voxire/onion/internal/modules/tax/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	taxScope = "tax"
	taxTTL   = 10 * time.Minute
)

type repository interface {
	GetConfig(ctx context.Context, ownerID int64) (domain.TaxConfig, error)
	UpsertConfig(ctx context.Context, input domain.TaxConfigInput) error
	GetPLReport(ctx context.Context, ownerID int64) (domain.PLReport, error)
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func invalidateTaxScope(ctx context.Context, ownerID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(taxScope, ownerID))
	}
}

func (s *Service) GetConfig(ctx context.Context, ownerID int64) (domain.TaxConfig, error) {
	return cacher.Of[domain.TaxConfig]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(taxScope, ownerID, "config"),
		func(ctx context.Context) (domain.TaxConfig, error) {
			return s.repo.GetConfig(ctx, ownerID)
		},
		cacher.WithExpiry(ctx, taxTTL),
	)
}

func (s *Service) UpdateConfig(ctx context.Context, input domain.TaxConfigInput) error {
	if err := s.repo.UpsertConfig(ctx, input); err != nil {
		return err
	}
	invalidateTaxScope(ctx, input.OwnerID)
	return nil
}

// GetPLReport derives from orders/expenses; piggyback on order namespace so
// any order or expense mutation invalidates today's P&L.
func (s *Service) GetPLReport(ctx context.Context, ownerID int64) (domain.PLReport, error) {
	key := cacheKeys.OrdersList(ownerID) + "|pl_report"
	return cacher.Of[domain.PLReport]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (domain.PLReport, error) {
			return s.repo.GetPLReport(ctx, ownerID)
		},
		cacher.WithExpiry(ctx, 2*time.Minute),
	)
}
