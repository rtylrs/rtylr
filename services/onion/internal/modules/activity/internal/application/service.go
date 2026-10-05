package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/activity/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
)

const activityCacheTTL = 1 * time.Minute

type repository interface {
	ListActivity(ctx context.Context, businessID int64, app string, limit, offset int) ([]domain.ActivityLog, int64, error)
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, repo repository) *Service {
	return &Service{
		repo: repo,
	}
}

// ListActivity returns a single page of activity for the business, optionally
// filtered to one app (empty = all), with the total count for that filter.
func (s *Service) ListActivity(ctx context.Context, businessID int64, app string, limit, offset int) (domain.ActivityPage, error) {
	cacheKey := fmt.Sprintf("activity:%d:%s:%d:%d", businessID, app, limit, offset)

	return cacher.Of[domain.ActivityPage]().GetOrComputeSingle(ctx, cacheKey+"|domain=v2", func(ctx context.Context) (domain.ActivityPage, error) {
		entries, total, err := s.repo.ListActivity(ctx, businessID, app, limit, offset)
		if err != nil {
			return domain.ActivityPage{}, err
		}
		return domain.ActivityPage{Entries: entries, Total: total}, nil
	}, cacher.WithExpiry(ctx, activityCacheTTL))
}
