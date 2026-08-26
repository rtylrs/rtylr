package application

import (
	"context"
	"database/sql"
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
	List(ctx context.Context, businessID int64) ([]domain.Modifier, error)
	Create(ctx context.Context, ownerID, businessID int64, input domain.ModifierInput) (int64, error)
	Update(ctx context.Context, businessID, id int64, input domain.ModifierInput) error
	Delete(ctx context.Context, businessID, id int64) error
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func invalidatePosModifiersScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(posModifiersScope, businessID))
}

func (s *Service) List(ctx context.Context, businessID int64) ([]domain.Modifier, error) {
	return cacher.Of[[]domain.Modifier]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(posModifiersScope, businessID, "list"),
		func(ctx context.Context) ([]domain.Modifier, error) {
			return s.repo.List(ctx, businessID)
		},
		cacher.WithExpiry(ctx, posModifiersTTL),
	)
}

func (s *Service) Create(ctx context.Context, ownerID, businessID int64, input domain.ModifierInput) (int64, error) {
	if err := domain.ValidateModifierInput(input); err != nil {
		return 0, err
	}
	id, err := s.repo.Create(ctx, ownerID, businessID, input)
	if err != nil {
		return 0, err
	}
	invalidatePosModifiersScope(ctx, businessID)
	return id, nil
}

func (s *Service) Update(ctx context.Context, businessID, id int64, input domain.ModifierInput) error {
	if err := s.repo.Update(ctx, businessID, id, input); err != nil {
		return err
	}
	invalidatePosModifiersScope(ctx, businessID)
	return nil
}

func (s *Service) Delete(ctx context.Context, businessID, id int64) error {
	if err := s.repo.Delete(ctx, businessID, id); err != nil {
		return err
	}
	invalidatePosModifiersScope(ctx, businessID)
	return nil
}
