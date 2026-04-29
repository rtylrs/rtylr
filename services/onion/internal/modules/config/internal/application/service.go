package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/config/internal/domain"
	"github.com/voxire/onion/internal/modules/config/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
)

const ConfigCacheTTL = 500 * time.Minute // 500 minutes = 8 hours

type repository interface {
	GetConfig(ctx context.Context, ownerID string) (domain.Config, error)
	CreateConfig(ctx context.Context, input domain.ConfigInput) error
	GetThemeByOwnerID(ctx context.Context, id string) (domain.Theme, error)
}

type Service struct {
	repo repository
}

func NewService(ctx context.Context) *Service {
	repo := infrastructure.NewRepository(ctx)

	return &Service{
		repo: repo,
	}
}

func (s *Service) GetGlobalConfig(ctx context.Context) (domain.Config, error) {
	return s.GetConfig(ctx, "1")
}

func (s *Service) GetConfig(ctx context.Context, ownerID string) (domain.Config, error) {
	cacheKey := "config:" + ownerID

	return cacher.Of[domain.Config]().GetOrComputeSingle(ctx, cacheKey, func(ctx context.Context) (domain.Config, error) {
		return s.repo.GetConfig(ctx, ownerID)
	}, cacher.WithExpiry(ctx, ConfigCacheTTL))
}

func (s *Service) CreateConfig(ctx context.Context, input domain.ConfigInput) error {
	return s.repo.CreateConfig(ctx, input)
}

func (s *Service) GetThemeByOwnerID(ctx context.Context, id string) (domain.Theme, error) {
	return s.repo.GetThemeByOwnerID(ctx, id)
}

