package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/business/internal/domain"
	"github.com/voxire/onion/internal/modules/business/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const businessProfileCacheTTL = 10 * time.Minute

type repository interface {
	GetProfile(ctx context.Context, ownerID int64) (domain.BusinessProfile, error)
}

type Service struct {
	repo repository
}

func NewService(ctx context.Context) *Service {
	return &Service{
		repo: infrastructure.NewRepository(ctx),
	}
}

func (s *Service) GetProfile(ctx context.Context, ownerID int64) (domain.BusinessProfile, error) {
	return cacher.Of[domain.BusinessProfile]().GetOrComputeSingle(
		ctx,
		cacheKeys.BusinessConfig(ownerID),
		func(ctx context.Context) (domain.BusinessProfile, error) {
			return s.repo.GetProfile(ctx, ownerID)
		},
		cacher.WithExpiry(ctx, businessProfileCacheTTL),
	)
}
