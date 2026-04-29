package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/automations/internal/domain"
	"github.com/voxire/onion/internal/modules/automations/internal/infrastructure"
)

type repository interface {
	List(ctx context.Context, ownerID int64) ([]domain.Automation, error)
	Create(ctx context.Context, input domain.AutomationInput) (int64, error)
	Update(ctx context.Context, input domain.AutomationUpdateInput) error
	Delete(ctx context.Context, ownerID, automationID int64) error
}

type Service struct {
	repo repository
}

func NewService(ctx context.Context) *Service {
	return &Service{
		repo: infrastructure.NewRepository(ctx),
	}
}

func (s *Service) List(ctx context.Context, ownerID int64) ([]domain.Automation, error) {
	return s.repo.List(ctx, ownerID)
}

func (s *Service) Create(ctx context.Context, input domain.AutomationInput) (int64, error) {
	return s.repo.Create(ctx, input)
}

func (s *Service) Update(ctx context.Context, input domain.AutomationUpdateInput) error {
	return s.repo.Update(ctx, input)
}

func (s *Service) Delete(ctx context.Context, ownerID, automationID int64) error {
	return s.repo.Delete(ctx, ownerID, automationID)
}
