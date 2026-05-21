package application

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/hr_dashboard/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_dashboard/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	hrDashboardScope = "hr_dashboard"
	hrDashboardTTL   = 5 * time.Minute
)

type repository interface {
	GetWidgets(ctx context.Context, ownerID int64) ([]*domain.DashboardWidget, error)
	UpsertWidgets(ctx context.Context, ownerID int64, widgets []*domain.DashboardWidget) error
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{
		repo: infrastructure.NewRepository(db),
	}
}

func invalidateHRDashboardScope(ctx context.Context, ownerID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(hrDashboardScope, ownerID))
	}
}

func (s *Service) GetWidgets(ctx context.Context, ownerID int64) ([]*domain.DashboardWidget, error) {
	return cacher.Of[[]*domain.DashboardWidget]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(hrDashboardScope, ownerID, "widgets"),
		func(ctx context.Context) ([]*domain.DashboardWidget, error) {
			return s.repo.GetWidgets(ctx, ownerID)
		},
		cacher.WithExpiry(ctx, hrDashboardTTL),
	)
}

func (s *Service) UpsertWidgets(ctx context.Context, ownerID int64, widgets []*domain.DashboardWidget) error {
	if err := s.repo.UpsertWidgets(ctx, ownerID, widgets); err != nil {
		return err
	}
	invalidateHRDashboardScope(ctx, ownerID)
	return nil
}
