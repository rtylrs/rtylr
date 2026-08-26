package application

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/automations/internal/domain"
	"github.com/voxire/onion/internal/modules/automations/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	automationsScope = "automations"
	automationsTTL   = 5 * time.Minute
)

type repository interface {
	List(ctx context.Context, businessID int64) ([]domain.Automation, error)
	Create(ctx context.Context, input domain.AutomationInput) (int64, error)
	Update(ctx context.Context, input domain.AutomationUpdateInput) error
	Delete(ctx context.Context, businessID, automationID int64) error
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{
		repo: infrastructure.NewRepository(db),
	}
}

func invalidateAutomationsScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(automationsScope, businessID))
}

func (s *Service) List(ctx context.Context, businessID int64) ([]domain.Automation, error) {
	return cacher.Of[[]domain.Automation]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(automationsScope, businessID, "list"),
		func(ctx context.Context) ([]domain.Automation, error) {
			return s.repo.List(ctx, businessID)
		},
		cacher.WithExpiry(ctx, automationsTTL),
	)
}

func (s *Service) Create(ctx context.Context, input domain.AutomationInput) (int64, error) {
	id, err := s.repo.Create(ctx, input)
	if err != nil {
		return 0, err
	}
	invalidateAutomationsScope(ctx, input.BusinessID)
	return id, nil
}

func (s *Service) Update(ctx context.Context, input domain.AutomationUpdateInput) error {
	if err := s.repo.Update(ctx, input); err != nil {
		return err
	}
	invalidateAutomationsScope(ctx, input.BusinessID)
	return nil
}

func (s *Service) Delete(ctx context.Context, businessID, automationID int64) error {
	if err := s.repo.Delete(ctx, businessID, automationID); err != nil {
		return err
	}
	invalidateAutomationsScope(ctx, businessID)
	return nil
}
