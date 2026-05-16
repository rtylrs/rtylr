package application

import (
	"context"
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
	List(ctx context.Context, ownerID int64) ([]domain.Automation, error)
	Create(ctx context.Context, input domain.AutomationInput) (int64, error)
	Update(ctx context.Context, input domain.AutomationUpdateInput) error
	Delete(ctx context.Context, ownerID, automationID int64) error
}

type Service struct {
	repo repository
}

func NewService(ctx context.Context) *Service {
	return &Service{
		repo: infrastructure.NewRepository(ctx),
	}
}

func invalidateAutomationsScope(ctx context.Context, ownerID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(automationsScope, ownerID))
	}
}

func (s *Service) List(ctx context.Context, ownerID int64) ([]domain.Automation, error) {
	return cacher.Of[[]domain.Automation]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(automationsScope, ownerID, "list"),
		func(ctx context.Context) ([]domain.Automation, error) {
			return s.repo.List(ctx, ownerID)
		},
		cacher.WithExpiry(ctx, automationsTTL),
	)
}

func (s *Service) Create(ctx context.Context, input domain.AutomationInput) (int64, error) {
	id, err := s.repo.Create(ctx, input)
	if err != nil {
		return 0, err
	}
	invalidateAutomationsScope(ctx, input.OwnerID)
	return id, nil
}

func (s *Service) Update(ctx context.Context, input domain.AutomationUpdateInput) error {
	if err := s.repo.Update(ctx, input); err != nil {
		return err
	}
	invalidateAutomationsScope(ctx, input.OwnerID)
	return nil
}

func (s *Service) Delete(ctx context.Context, ownerID, automationID int64) error {
	if err := s.repo.Delete(ctx, ownerID, automationID); err != nil {
		return err
	}
	invalidateAutomationsScope(ctx, ownerID)
	return nil
}
