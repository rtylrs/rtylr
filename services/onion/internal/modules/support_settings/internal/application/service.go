package application

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/support_settings/internal/domain"
	"github.com/voxire/onion/internal/modules/support_settings/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	supportSettingsScope = "support_settings"
	supportSettingsTTL   = 10 * time.Minute
)

type repository interface {
	Get(ctx context.Context, businessID int64) (*domain.Settings, error)
	Upsert(ctx context.Context, input domain.SettingsInput) error
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{
		repo: infrastructure.NewRepository(db),
	}
}

func invalidateSupportSettingsScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(supportSettingsScope, businessID))
}

func (s *Service) Get(ctx context.Context, businessID int64) (*domain.Settings, error) {
	return cacher.Of[*domain.Settings]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(supportSettingsScope, businessID, "settings"),
		func(ctx context.Context) (*domain.Settings, error) {
			return s.repo.Get(ctx, businessID)
		},
		cacher.WithExpiry(ctx, supportSettingsTTL),
	)
}

func (s *Service) Upsert(ctx context.Context, input domain.SettingsInput) error {
	if input.SLAHours <= 0 || input.SLAHours > 1000 {
		return domain.ErrInvalidSLAHours
	}
	if err := s.repo.Upsert(ctx, input); err != nil {
		return err
	}
	invalidateSupportSettingsScope(ctx, input.BusinessID)
	return nil
}
