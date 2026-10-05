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
	QuickCreate(ctx context.Context, ownerID, businessID int64, input domain.MenuItemInput) (int64, error)
	Update(ctx context.Context, ownerID, businessID, id int64, input domain.MenuItemUpdateInput) error
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
	owner          func(context.Context, int64) (int64, error)
	onlineOrdering func(context.Context, int64) (bool, error)
}

func (s *Service) SetBusinessContext(owner func(context.Context, int64) (int64, error), onlineOrdering func(context.Context, int64) (bool, error)) {
	s.owner, s.onlineOrdering = owner, onlineOrdering
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

func (s *Service) QuickCreate(ctx context.Context, ownerID, businessID int64, input domain.MenuItemInput) (int64, error) {
	ownerID, err := s.BusinessOwnerID(ctx, ownerID, businessID)
	if err != nil {
		return 0, err
	}
	id, err := s.repo.QuickCreate(ctx, ownerID, businessID, input)
	if err != nil {
		return 0, err
	}
	invalidateMenuScope(ctx, businessID)
	return id, nil
}

func (s *Service) Update(ctx context.Context, businessID, id int64, input domain.MenuItemUpdateInput) error {
	ownerID, err := s.BusinessOwnerID(ctx, 0, businessID)
	if err != nil {
		return err
	}
	if err := s.repo.Update(ctx, ownerID, businessID, id, input); err != nil {
		return err
	}
	invalidateMenuScope(ctx, businessID)
	return nil
}

func (s *Service) BusinessOwnerID(ctx context.Context, fallbackOwnerID, businessID int64) (int64, error) {
	if businessID == 0 {
		return fallbackOwnerID, nil
	}
	if s.owner == nil {
		return 0, errors.New("business owner resolver is unavailable")
	}
	return s.owner(ctx, businessID)
}
