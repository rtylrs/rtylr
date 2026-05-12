package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/campaigns/internal/domain"
	"github.com/voxire/onion/internal/modules/campaigns/internal/infrastructure"
)

type repository interface {
	List(ctx context.Context, ownerID int64) ([]domain.Campaign, error)
	GetByID(ctx context.Context, ownerID, id int64) (domain.Campaign, error)
	Create(ctx context.Context, input domain.CampaignInput) (int64, error)
	Update(ctx context.Context, input domain.CampaignUpdateInput) error
	UpdateStatus(ctx context.Context, ownerID, id int64, status string) error
	Delete(ctx context.Context, ownerID, id int64) error
}

type Service struct {
	repo repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) List(ctx context.Context, ownerID int64) ([]domain.Campaign, error) {
	return s.repo.List(ctx, ownerID)
}

func (s *Service) GetByID(ctx context.Context, ownerID, id int64) (domain.Campaign, error) {
	return s.repo.GetByID(ctx, ownerID, id)
}

func (s *Service) Create(ctx context.Context, input domain.CampaignInput) (int64, error) {
	return s.repo.Create(ctx, input)
}

func (s *Service) Update(ctx context.Context, input domain.CampaignUpdateInput) error {
	return s.repo.Update(ctx, input)
}

func (s *Service) UpdateStatus(ctx context.Context, ownerID, id int64, status string) error {
	return s.repo.UpdateStatus(ctx, ownerID, id, status)
}

func (s *Service) Delete(ctx context.Context, ownerID, id int64) error {
	return s.repo.Delete(ctx, ownerID, id)
}
