package application

import (
	"context"
	"database/sql"
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
	List(ctx context.Context, businessID int64) ([]domain.Combo, error)
	Create(ctx context.Context, ownerID, businessID int64, input domain.ComboInput) (int64, error)
	Update(ctx context.Context, businessID, id int64, input domain.ComboInput) error
	Delete(ctx context.Context, businessID, id int64) error
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func invalidatePosCombosScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(posCombosScope, businessID))
}

func (s *Service) List(ctx context.Context, businessID int64) ([]domain.Combo, error) {
	return cacher.Of[[]domain.Combo]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(posCombosScope, businessID, "list"),
		func(ctx context.Context) ([]domain.Combo, error) {
			return s.repo.List(ctx, businessID)
		},
		cacher.WithExpiry(ctx, posCombosTTL),
	)
}

func (s *Service) Create(ctx context.Context, ownerID, businessID int64, input domain.ComboInput) (int64, error) {
	if err := domain.ValidateComboInput(input); err != nil {
		return 0, err
	}
	id, err := s.repo.Create(ctx, ownerID, businessID, input)
	if err != nil {
		return 0, err
	}
	invalidatePosCombosScope(ctx, businessID)
	return id, nil
}

func (s *Service) Update(ctx context.Context, businessID, id int64, input domain.ComboInput) error {
	if err := s.repo.Update(ctx, businessID, id, input); err != nil {
		return err
	}
	invalidatePosCombosScope(ctx, businessID)
	return nil
}

func (s *Service) Delete(ctx context.Context, businessID, id int64) error {
	if err := s.repo.Delete(ctx, businessID, id); err != nil {
		return err
	}
	invalidatePosCombosScope(ctx, businessID)
	return nil
}
