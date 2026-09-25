package application

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/modules/erp_settings/internal/domain"
	"github.com/voxire/onion/internal/modules/erp_settings/internal/infrastructure"
)

type repository interface {
	Get(ctx context.Context, businessID int64) (domain.Settings, error)
	Upsert(ctx context.Context, input domain.SettingsInput) error
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func (s *Service) GetSettings(ctx context.Context, businessID int64) (domain.Settings, error) {
	if businessID <= 0 {
		return domain.Settings{}, domain.ValidateInput(domain.SettingsInput{BusinessID: businessID})
	}

	settings, err := s.repo.Get(ctx, businessID)
	if errors.Is(err, domain.ErrSettingsNotFound) {
		return domain.DefaultSettings(businessID), nil
	}
	return settings, err
}

func (s *Service) UpdateSettings(ctx context.Context, input domain.SettingsInput) (domain.Settings, error) {
	if err := domain.ValidateInput(input); err != nil {
		return domain.Settings{}, err
	}
	if err := s.repo.Upsert(ctx, input); err != nil {
		return domain.Settings{}, err
	}
	return s.GetSettings(ctx, input.BusinessID)
}
