package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/recruit/careers/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
)

const careersCatalogKey = "cache:careers:active"

type Service struct {
	repo repository
}

func NewService(repo repository) *Service {
	return &Service{repo: repo}
}

// GetActiveCareers is a public/global read; cached under a single fixed key
// with a long TTL (manual flush or restart picks up new jobs).
func (s *Service) GetActiveCareers(ctx context.Context) ([]domain.Career, error) {
	return cacher.Of[[]domain.Career]().GetOrComputeSingle(ctx, careersCatalogKey,
		func(ctx context.Context) ([]domain.Career, error) {
			return s.repo.GetActiveCareers(ctx)
		},
		cacher.WithExpiry(ctx, 15*time.Minute),
	)
}

// InvalidateActiveCareersCache busts the cached public careers catalogue.
// Exposed so other code that writes into the shared careers table (e.g.
// the voxhire module) can invalidate it without needing careers' internal
// domain/cache types.
func InvalidateActiveCareersCache(ctx context.Context) {
	cacher.InvalidateTypedScope[[]domain.Career](ctx, careersCatalogKey)
}
