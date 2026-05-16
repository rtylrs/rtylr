package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/pos_combos/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_combos/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	posCombosScope = "pos_combos"
	posCombosTTL   = 5 * time.Minute
)

type repository interface {
	List(ctx context.Context, ownerID int64) ([]domain.Combo, error)
	Create(ctx context.Context, ownerID int64, input domain.ComboInput) (int64, error)
	Update(ctx context.Context, ownerID, id int64, input domain.ComboInput) error
	Delete(ctx context.Context, ownerID, id int64) error
}

type Service struct {
	repo repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func invalidatePosCombosScope(ctx context.Context, ownerID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(posCombosScope, ownerID))
	}
}

func (s *Service) List(ctx context.Context, ownerID int64) ([]domain.Combo, error) {
	return cacher.Of[[]domain.Combo]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(posCombosScope, ownerID, "list"),
		func(ctx context.Context) ([]domain.Combo, error) {
			return s.repo.List(ctx, ownerID)
		},
		cacher.WithExpiry(ctx, posCombosTTL),
	)
}

func (s *Service) Create(ctx context.Context, ownerID int64, input domain.ComboInput) (int64, error) {
	id, err := s.repo.Create(ctx, ownerID, input)
	if err != nil {
		return 0, err
	}
	invalidatePosCombosScope(ctx, ownerID)
	return id, nil
}

func (s *Service) Update(ctx context.Context, ownerID, id int64, input domain.ComboInput) error {
	if err := s.repo.Update(ctx, ownerID, id, input); err != nil {
		return err
	}
	invalidatePosCombosScope(ctx, ownerID)
	return nil
}

func (s *Service) Delete(ctx context.Context, ownerID, id int64) error {
	if err := s.repo.Delete(ctx, ownerID, id); err != nil {
		return err
	}
	invalidatePosCombosScope(ctx, ownerID)
	return nil
}
