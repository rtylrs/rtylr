package application

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/campaigns/internal/domain"
	"github.com/voxire/onion/internal/modules/campaigns/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	campaignsScope = "campaigns"
	campaignsTTL   = 5 * time.Minute
)

type repository interface {
	List(ctx context.Context, businessID int64) ([]domain.Campaign, error)
	GetByID(ctx context.Context, businessID, id int64) (domain.Campaign, error)
	Create(ctx context.Context, input domain.CampaignInput) (int64, error)
	Update(ctx context.Context, input domain.CampaignUpdateInput) error
	UpdateStatus(ctx context.Context, businessID, id int64, status string) error
	Delete(ctx context.Context, businessID, id int64) error
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func invalidateCampaignsScope(ctx context.Context, businessID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(campaignsScope, businessID))
	}
}

func (s *Service) List(ctx context.Context, businessID int64) ([]domain.Campaign, error) {
	return cacher.Of[[]domain.Campaign]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(campaignsScope, businessID, "list"),
		func(ctx context.Context) ([]domain.Campaign, error) {
			return s.repo.List(ctx, businessID)
		},
		cacher.WithExpiry(ctx, campaignsTTL),
	)
}

func (s *Service) GetByID(ctx context.Context, businessID, id int64) (domain.Campaign, error) {
	key := cacheKeys.Scoped(campaignsScope, businessID, "id="+strconv.FormatInt(id, 10))
	return cacher.Of[domain.Campaign]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (domain.Campaign, error) {
			return s.repo.GetByID(ctx, businessID, id)
		},
		cacher.WithExpiry(ctx, campaignsTTL),
	)
}

func (s *Service) Create(ctx context.Context, input domain.CampaignInput) (int64, error) {
	if err := domain.ValidateCampaignInput(input); err != nil {
		return 0, err
	}
	id, err := s.repo.Create(ctx, input)
	if err != nil {
		return 0, err
	}
	invalidateCampaignsScope(ctx, input.BusinessID)
	return id, nil
}

func (s *Service) Update(ctx context.Context, input domain.CampaignUpdateInput) error {
	if err := domain.ValidateCampaignUpdateInput(input); err != nil {
		return err
	}
	if err := s.repo.Update(ctx, input); err != nil {
		return err
	}
	invalidateCampaignsScope(ctx, input.BusinessID)
	return nil
}

func (s *Service) UpdateStatus(ctx context.Context, businessID, id int64, status string) error {
	if err := s.repo.UpdateStatus(ctx, businessID, id, status); err != nil {
		return err
	}
	invalidateCampaignsScope(ctx, businessID)
	return nil
}

func (s *Service) Delete(ctx context.Context, businessID, id int64) error {
	if err := s.repo.Delete(ctx, businessID, id); err != nil {
		return err
	}
	invalidateCampaignsScope(ctx, businessID)
	return nil
}
