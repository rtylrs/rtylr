package application

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/pos_menu/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_menu/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const menuCacheTTL = 5 * time.Minute

type repository interface {
	List(ctx context.Context, businessID int64) ([]domain.MenuItem, error)
	IsOnlineOrderingEnabled(ctx context.Context, businessID int64) (bool, error)
	QuickCreate(ctx context.Context, ownerID, businessID int64, input domain.MenuItemInput) (int64, error)
	Update(ctx context.Context, businessID, id int64, input domain.MenuItemUpdateInput) error
	BusinessOwnerID(ctx context.Context, businessID int64) (int64, error)
}

func (s *Service) ListPublic(ctx context.Context, businessID int64) ([]domain.MenuItem, error) {
	enabled, err := s.repo.IsOnlineOrderingEnabled(ctx, businessID)
	if err != nil {
		return nil, err
	}
	if !enabled {
		return []domain.MenuItem{}, nil
	}
	return s.List(ctx, businessID)
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func invalidateMenuScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.MenuPrefix(businessID))
}

func (s *Service) List(ctx context.Context, businessID int64) ([]domain.MenuItem, error) {
	return cacher.Of[[]domain.MenuItem]().GetOrComputeSingle(ctx,
		cacheKeys.MenusList(businessID),
		func(ctx context.Context) ([]domain.MenuItem, error) {
			return s.repo.List(ctx, businessID)
		},
		cacher.WithExpiry(ctx, menuCacheTTL),
	)
}

func (s *Service) QuickCreate(ctx context.Context, ownerID, businessID int64, input domain.MenuItemInput) (int64, error) {
	id, err := s.repo.QuickCreate(ctx, ownerID, businessID, input)
	if err != nil {
		return 0, err
	}
	invalidateMenuScope(ctx, businessID)
	return id, nil
}

func (s *Service) Update(ctx context.Context, businessID, id int64, input domain.MenuItemUpdateInput) error {
	if err := s.repo.Update(ctx, businessID, id, input); err != nil {
		return err
	}
	invalidateMenuScope(ctx, businessID)
	return nil
}

func (s *Service) BusinessOwnerID(ctx context.Context, fallbackOwnerID, businessID int64) (int64, error) {
	if businessID == 0 {
		return fallbackOwnerID, nil
	}
	return s.repo.BusinessOwnerID(ctx, businessID)
}
