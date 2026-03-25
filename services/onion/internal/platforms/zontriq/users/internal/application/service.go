package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/platforms/zontriq/users/internal/domain"
	"github.com/voxire/onion/internal/platforms/zontriq/users/internal/infrastructure"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	repo := infrastructure.NewRepository(ctx)

	return &Service{
		repo: repo,
	}
}

func (s *Service) GetZontriqUserByUserID(ctx context.Context, userID int) (*domain.ZontriqUser, error) {
	return s.repo.GetZontriqUserByUserID(ctx, userID)
}

func (s *Service) CreateZontriqUser(ctx context.Context, userID, planID int) (*domain.ZontriqUser, error) {
	return s.repo.CreateZontriqUser(ctx, userID, planID)
}

func (s *Service) GetUserPlanByUserID(ctx context.Context, userID int) (*domain.Plan, error) {
	return s.repo.GetUserPlanByUserID(ctx, userID)
}

func (s *Service) CompleteOnboarding(ctx context.Context, userID int) error {
	return s.repo.CompleteOnboarding(ctx, userID)
}

func (s *Service) UpdateZontriqUserPlan(ctx context.Context, id, planID int) error {
	return s.repo.UpdateZontriqUserPlan(ctx, id, planID)
}

func (s *Service) SoftDeleteZontriqUser(ctx context.Context, id int) error {
	return s.repo.SoftDeleteZontriqUser(ctx, id)
}

func (s *Service) ListPlans(ctx context.Context) ([]domain.Plan, error) {
	return s.repo.ListPlans(ctx)
}

func (s *Service) GetPlanByID(ctx context.Context, planID int) (*domain.Plan, error) {
	return s.repo.GetPlanByID(ctx, planID)
}

func (s *Service) GetCurrentUsage(ctx context.Context, zontriqUserID int) (*domain.Usage, error) {
	return s.repo.GetCurrentUsage(ctx, zontriqUserID)
}

func (s *Service) IncrementVideosGenerated(ctx context.Context, usageID int) error {
	return s.repo.IncrementVideosGenerated(ctx, usageID)
}

func (s *Service) IncrementVideosPublished(ctx context.Context, usageID int) error {
	return s.repo.IncrementVideosPublished(ctx, usageID)
}

func (s *Service) CreateUsagePeriod(ctx context.Context, zontriqUserID int, periodStart, periodEnd time.Time) (*domain.Usage, error) {
	return s.repo.CreateUsagePeriod(ctx, zontriqUserID, periodStart, periodEnd)
}

func (s *Service) GetPreferences(ctx context.Context, zontriqUserID int) (*domain.Preferences, error) {
	return s.repo.GetPreferences(ctx, zontriqUserID)
}

func (s *Service) CreatePreferences(ctx context.Context, prefs domain.Preferences) (*domain.Preferences, error) {
	return s.repo.CreatePreferences(ctx, prefs)
}

func (s *Service) CreateWaitlistEntry(ctx context.Context, email, source string) error {
	return s.repo.CreateWaitlistEntry(ctx, email, source)
}

func (s *Service) UpdatePreferences(ctx context.Context, zontriqUserID int, prefs domain.Preferences) (*domain.Preferences, error) {
	return s.repo.UpdatePreferences(ctx, zontriqUserID, prefs)
}
