package application

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/kpis/internal/domain"
	"github.com/voxire/onion/internal/modules/kpis/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const kpiCacheTTL = 2 * time.Minute

type repository interface {
	GetKPIs(ctx context.Context, businessID int64) (domain.KPISummary, error)
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{
		repo: infrastructure.NewRepository(db),
	}
}

func (s *Service) GetKPIs(ctx context.Context, businessID int64) (domain.KPISummary, error) {
	return cacher.Of[domain.KPISummary]().GetOrComputeSingle(
		ctx,
		cacheKeys.BusinessKPIs(businessID),
		func(ctx context.Context) (domain.KPISummary, error) {
			return s.repo.GetKPIs(ctx, businessID)
		},
		cacher.WithExpiry(ctx, kpiCacheTTL),
	)
}
