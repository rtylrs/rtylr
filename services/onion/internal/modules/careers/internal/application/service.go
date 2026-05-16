package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/careers/internal/domain"
	"github.com/voxire/onion/internal/modules/careers/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
)

const careersCatalogKey = "cache:careers:active"

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
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
