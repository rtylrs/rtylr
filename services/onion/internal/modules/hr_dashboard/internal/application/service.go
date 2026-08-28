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
	GetWidgets(ctx context.Context, businessID int64) ([]*domain.DashboardWidget, error)
	UpsertWidgets(ctx context.Context, businessID int64, widgets []*domain.DashboardWidget) error
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{
		repo: infrastructure.NewRepository(db),
	}
}

func invalidateHRDashboardScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(hrDashboardScope, businessID))
}

func (s *Service) GetWidgets(ctx context.Context, businessID int64) ([]*domain.DashboardWidget, error) {
	return cacher.Of[[]*domain.DashboardWidget]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(hrDashboardScope, businessID, "widgets"),
		func(ctx context.Context) ([]*domain.DashboardWidget, error) {
			return s.repo.GetWidgets(ctx, businessID)
		},
		cacher.WithExpiry(ctx, hrDashboardTTL),
	)
}

func (s *Service) UpsertWidgets(ctx context.Context, businessID int64, widgets []*domain.DashboardWidget) error {
	if err := s.repo.UpsertWidgets(ctx, businessID, widgets); err != nil {
		return err
	}
	invalidateHRDashboardScope(ctx, businessID)
	return nil
}
