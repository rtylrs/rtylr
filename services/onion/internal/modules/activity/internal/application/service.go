package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/activity/internal/domain"
	"github.com/voxire/onion/internal/modules/activity/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
)

const activityCacheTTL = 1 * time.Minute

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	repo := infrastructure.NewRepository(ctx)

	return &Service{
		repo: repo,
	}
}

func (s *Service) ListActivity(ctx context.Context, ownerID int64, limit int) ([]domain.ActivityLog, error) {
	cacheKey := fmt.Sprintf("activity:%d:%d", ownerID, limit)

	return cacher.Of[[]domain.ActivityLog]().GetOrComputeSingle(ctx, cacheKey, func(ctx context.Context) ([]domain.ActivityLog, error) {
		return s.repo.ListActivity(ctx, ownerID, limit)
	}, cacher.WithExpiry(ctx, activityCacheTTL))
}
