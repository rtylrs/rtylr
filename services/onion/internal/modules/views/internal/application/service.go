package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/views/internal/domain"
	"github.com/voxire/onion/internal/modules/views/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	viewsScope = "views"
	viewsTTL   = 5 * time.Minute
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func invalidateViewsScope(ctx context.Context, ownerID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(viewsScope, ownerID))
	}
}

func (s *Service) List(ctx context.Context, ownerID, workspaceID int64) ([]domain.View, error) {
	key := cacheKeys.Scoped(viewsScope, ownerID, fmt.Sprintf("list|w=%d", workspaceID))
	return cacher.Of[[]domain.View]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) ([]domain.View, error) {
			return s.repo.List(ctx, ownerID, workspaceID)
		},
		cacher.WithExpiry(ctx, viewsTTL),
	)
}

func (s *Service) Create(ctx context.Context, input domain.ViewInput) (int64, error) {
	id, err := s.repo.Create(ctx, input)
	if err != nil {
		return 0, err
	}
	invalidateViewsScope(ctx, input.OwnerID)
	return id, nil
}

func (s *Service) Update(ctx context.Context, input domain.ViewUpdateInput) error {
	if err := s.repo.Update(ctx, input); err != nil {
		return err
	}
	invalidateViewsScope(ctx, input.OwnerID)
	return nil
}

func (s *Service) Delete(ctx context.Context, id, ownerID int64) error {
	if err := s.repo.Delete(ctx, id, ownerID); err != nil {
		return err
	}
	invalidateViewsScope(ctx, ownerID)
	return nil
}
