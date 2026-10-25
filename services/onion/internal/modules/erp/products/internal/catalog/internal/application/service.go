package application

import (
	"context"
	"errors"
	"time"

	"github.com/voxire/onion/internal/modules/erp/products/internal/catalog/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const menuCacheTTL = 5 * time.Minute

type repository interface {
	List(ctx context.Context, businessID int64) ([]domain.MenuItem, error)
	QuickCreate(ctx context.Context, businessID int64, input domain.MenuItemInput) (int64, error)
	Update(ctx context.Context, businessID, id int64, input domain.MenuItemUpdateInput) error
}

func (s *Service) ListPublic(ctx context.Context, businessID int64) ([]domain.MenuItem, error) {
	if s.onlineOrdering == nil {
		return nil, errors.New("online ordering resolver is unavailable")
	}
	enabled, err := s.onlineOrdering(ctx, businessID)
	if err != nil {
		return nil, err
	}
	if !enabled {
		return []domain.MenuItem{}, nil
	}
	return s.List(ctx, businessID)
}

type Service struct {
	repo           repository
	onlineOrdering func(context.Context, int64) (bool, error)
}

func (s *Service) SetOnlineOrdering(onlineOrdering func(context.Context, int64) (bool, error)) {
	s.onlineOrdering = onlineOrdering
}

func NewService(_ context.Context, repo repository) *Service {
	return &Service{repo: repo}
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

func (s *Service) QuickCreate(ctx context.Context, businessID int64, input domain.MenuItemInput) (int64, error) {
	id, err := s.repo.QuickCreate(ctx, businessID, input)
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
