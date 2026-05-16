package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/pos_modifiers/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_modifiers/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	posModifiersScope = "pos_modifiers"
	posModifiersTTL   = 5 * time.Minute
)

type repository interface {
	List(ctx context.Context, ownerID int64) ([]domain.Modifier, error)
	Create(ctx context.Context, ownerID int64, input domain.ModifierInput) (int64, error)
	Update(ctx context.Context, ownerID, id int64, input domain.ModifierInput) error
	Delete(ctx context.Context, ownerID, id int64) error
}

type Service struct {
	repo repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func invalidatePosModifiersScope(ctx context.Context, ownerID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(posModifiersScope, ownerID))
	}
}

func (s *Service) List(ctx context.Context, ownerID int64) ([]domain.Modifier, error) {
	return cacher.Of[[]domain.Modifier]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(posModifiersScope, ownerID, "list"),
		func(ctx context.Context) ([]domain.Modifier, error) {
			return s.repo.List(ctx, ownerID)
		},
		cacher.WithExpiry(ctx, posModifiersTTL),
	)
}

func (s *Service) Create(ctx context.Context, ownerID int64, input domain.ModifierInput) (int64, error) {
	id, err := s.repo.Create(ctx, ownerID, input)
	if err != nil {
		return 0, err
	}
	invalidatePosModifiersScope(ctx, ownerID)
	return id, nil
}

func (s *Service) Update(ctx context.Context, ownerID, id int64, input domain.ModifierInput) error {
	if err := s.repo.Update(ctx, ownerID, id, input); err != nil {
		return err
	}
	invalidatePosModifiersScope(ctx, ownerID)
	return nil
}

func (s *Service) Delete(ctx context.Context, ownerID, id int64) error {
	if err := s.repo.Delete(ctx, ownerID, id); err != nil {
		return err
	}
	invalidatePosModifiersScope(ctx, ownerID)
	return nil
}
