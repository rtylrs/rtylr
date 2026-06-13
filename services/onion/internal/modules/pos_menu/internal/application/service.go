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
	List(ctx context.Context, ownerID int64) ([]domain.MenuItem, error)
	QuickCreate(ctx context.Context, ownerID int64, input domain.MenuItemInput) (int64, error)
	BusinessOwnerID(ctx context.Context, businessID int64) (int64, error)
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func invalidateMenuScope(ctx context.Context, ownerID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.MenuPrefix(ownerID))
	}
}

func (s *Service) List(ctx context.Context, ownerID int64) ([]domain.MenuItem, error) {
	return cacher.Of[[]domain.MenuItem]().GetOrComputeSingle(ctx,
		cacheKeys.MenusList(ownerID),
		func(ctx context.Context) ([]domain.MenuItem, error) {
			return s.repo.List(ctx, ownerID)
		},
		cacher.WithExpiry(ctx, menuCacheTTL),
	)
}

func (s *Service) QuickCreate(ctx context.Context, ownerID int64, input domain.MenuItemInput) (int64, error) {
	id, err := s.repo.QuickCreate(ctx, ownerID, input)
	if err != nil {
		return 0, err
	}
	invalidateMenuScope(ctx, ownerID)
	return id, nil
}

func (s *Service) BusinessOwnerID(ctx context.Context, fallbackOwnerID, businessID int64) (int64, error) {
	if businessID == 0 {
		return fallbackOwnerID, nil
	}
	return s.repo.BusinessOwnerID(ctx, businessID)
}
