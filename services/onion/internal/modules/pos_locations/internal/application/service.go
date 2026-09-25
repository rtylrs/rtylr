package application

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/pos_locations/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_locations/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	posLocationsScope = "pos_locations"
	posLocationsTTL   = 5 * time.Minute
)

type repository interface {
	List(ctx context.Context, businessID int64) ([]domain.Location, error)
	Create(ctx context.Context, input domain.LocationInput) (int64, error)
	Update(ctx context.Context, input domain.LocationUpdateInput) error
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func invalidatePosLocationsScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(posLocationsScope, businessID))
}

func (s *Service) List(ctx context.Context, businessID int64) ([]domain.Location, error) {
	return cacher.Of[[]domain.Location]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(posLocationsScope, businessID, "list"),
		func(ctx context.Context) ([]domain.Location, error) {
			return s.repo.List(ctx, businessID)
		},
		cacher.WithExpiry(ctx, posLocationsTTL),
	)
}

func (s *Service) Create(ctx context.Context, input domain.LocationInput) (int64, error) {
	if err := domain.ValidateLocationInput(input); err != nil {
		return 0, err
	}
	id, err := s.repo.Create(ctx, input)
	if err != nil {
		return 0, err
	}
	invalidatePosLocationsScope(ctx, input.BusinessID)
	return id, nil
}

func (s *Service) Update(ctx context.Context, input domain.LocationUpdateInput) error {
	if err := domain.ValidateLocationUpdateInput(input); err != nil {
		return err
	}
	if err := s.repo.Update(ctx, input); err != nil {
		return err
	}
	invalidatePosLocationsScope(ctx, input.BusinessID)
	return nil
}
