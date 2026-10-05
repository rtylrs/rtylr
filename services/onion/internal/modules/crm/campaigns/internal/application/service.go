package application

import (
	"context"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/crm/campaigns/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	campaignsScope = "campaigns"
	campaignsTTL   = 5 * time.Minute

	defaultListLimit int32 = 20
	maxListLimit     int32 = 100
)

type repository interface {
	List(ctx context.Context, businessID int64) ([]domain.Campaign, error)
	ListPaginated(ctx context.Context, businessID int64, limit, offset int32) ([]domain.Campaign, error)
	Count(ctx context.Context, businessID int64) (int64, error)
	GetByID(ctx context.Context, businessID, id int64) (domain.Campaign, error)
	Create(ctx context.Context, input domain.CampaignInput) (int64, error)
	Update(ctx context.Context, input domain.CampaignUpdateInput) error
	UpdateStatus(ctx context.Context, businessID, id int64, status string) error
	Delete(ctx context.Context, businessID, id int64) error
}

type statusRepository interface {
	ListByStatus(ctx context.Context, businessID int64, status string) ([]domain.Campaign, error)
}

func (s *Service) ListByStatus(ctx context.Context, businessID int64, status string) ([]domain.Campaign, error) {
	if status == "" {
		return s.List(ctx, businessID)
	}

	if repo, ok := s.repo.(statusRepository); ok {
		return repo.ListByStatus(ctx, businessID, status)
	}

	all, err := s.List(ctx, businessID)
	if err != nil {
		return nil, err
	}

	filtered := make([]domain.Campaign, 0, len(all))
	for _, campaign := range all {
		if campaign.Status == status {
			filtered = append(filtered, campaign)
		}
	}

	return filtered, nil
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, repo repository) *Service {
	return &Service{repo: repo}
}

func invalidateCampaignsScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(campaignsScope, businessID))
}

func (s *Service) List(ctx context.Context, businessID int64) ([]domain.Campaign, error) {
	return cacher.Of[[]domain.Campaign]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(campaignsScope, businessID, "list")+"|domain=v2",
		func(ctx context.Context) ([]domain.Campaign, error) {
			return s.repo.List(ctx, businessID)
		},
		cacher.WithExpiry(ctx, campaignsTTL),
	)
}

// ListPaginated returns a single page of campaigns plus the tenant's total
// count. A non-positive limit defaults to 20 and limit is capped at 100; a
// negative offset is clamped to zero.
func (s *Service) ListPaginated(ctx context.Context, businessID int64, limit, offset int32) (domain.CampaignPage, error) {
	if limit <= 0 {
		limit = defaultListLimit
	}
	if limit > maxListLimit {
		limit = maxListLimit
	}
	if offset < 0 {
		offset = 0
	}

	campaigns, err := s.repo.ListPaginated(ctx, businessID, limit, offset)
	if err != nil {
		return domain.CampaignPage{}, err
	}

	total, err := s.repo.Count(ctx, businessID)
	if err != nil {
		return domain.CampaignPage{}, err
	}

	return domain.CampaignPage{
		Campaigns: campaigns,
		Total:     total,
		Limit:     limit,
		Offset:    offset,
	}, nil
}

func (s *Service) GetByID(ctx context.Context, businessID, id int64) (domain.Campaign, error) {
	key := cacheKeys.Scoped(campaignsScope, businessID, "id="+strconv.FormatInt(id, 10))
	return cacher.Of[domain.Campaign]().GetOrComputeSingle(ctx, key+"|domain=v2",
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
	if err := domain.ValidateCampaignStatus(status); err != nil {
		return err
	}
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
