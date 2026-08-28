package application

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/voxire/onion/internal/modules/pos_settings/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_settings/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	posSettingsScope = "pos_settings"
	posSettingsTTL   = 2 * time.Minute
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func (s *Service) GetSettings(ctx context.Context, businessID int64) (domain.Settings, error) {
	return cacher.Of[domain.Settings]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(posSettingsScope, businessID, "settings"),
		func(ctx context.Context) (domain.Settings, error) {
			settings, err := s.repo.GetSettings(ctx, businessID)
			if errors.Is(err, domain.ErrSettingsNotFound) {
				return domain.DefaultSettings(businessID), nil
			}
			return settings, err
		},
		cacher.WithExpiry(ctx, posSettingsTTL),
	)
}

func (s *Service) UpdateSettings(ctx context.Context, input domain.UpdateSettingsInput) (domain.Settings, error) {
	features := input.Features
	if !input.HasFeatures {
		features = domain.DefaultFeatures(domain.NormalizeProfile(input.Profile))
	}
	settings := domain.NormalizeSettings(domain.Settings{
		BusinessID:    input.BusinessID,
		Profile:       input.Profile,
		Theme:         input.Theme,
		Features:      features,
		ReceiptFooter: input.ReceiptFooter,
	})
	if err := s.repo.UpsertSettings(ctx, settings); err != nil {
		return domain.Settings{}, err
	}
	invalidatePOSSettings(ctx, input.BusinessID)
	return s.GetSettings(ctx, input.BusinessID)
}

func invalidatePOSSettings(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(posSettingsScope, businessID))
}
