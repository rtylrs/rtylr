package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/subscriptions/internal/domain"
	"github.com/voxire/onion/internal/modules/subscriptions/internal/infrastructure"
)

type Service struct {
	repo domain.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) ListPlans(ctx context.Context) ([]domain.Plan, error) {
	return s.repo.ListPlans(ctx)
}

func (s *Service) GetMySubscription(ctx context.Context, ownerID int64) (*domain.Subscription, error) {
	return s.repo.GetSubscription(ctx, ownerID)
}

func (s *Service) UpsertSubscription(ctx context.Context, ownerID int64, planSlug string) error {
	plan, err := s.repo.GetPlanBySlug(ctx, planSlug)
	if err != nil {
		return err
	}

	return s.repo.UpsertSubscription(ctx, ownerID, plan.ID)
}

func (s *Service) GetOwnerApps(ctx context.Context, ownerID int64) ([]string, error) {
	return s.repo.GetOwnerApps(ctx, ownerID)
}
