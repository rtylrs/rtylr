package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/hr_setup/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_setup/internal/infrastructure"
)

var validStepKeys = map[string]struct{}{
	"company_profile":  {},
	"work_schedule":    {},
	"org_structure":    {},
	"locations":        {},
	"roles_permissions": {},
	"payroll":          {},
	"notifications":    {},
	"employees":        {},
}

type repository interface {
	GetSetupSteps(ctx context.Context, ownerID int64) ([]*domain.SetupStep, error)
	UpsertSetupStep(ctx context.Context, ownerID int64, stepKey string, at time.Time) error
	GetHRConfig(ctx context.Context, ownerID int64) (*domain.HRConfig, error)
	UpsertHRConfig(ctx context.Context, ownerID int64, completedAt *time.Time) (*domain.HRConfig, error)
}

type Service struct {
	repo repository
}

func NewService(ctx context.Context) *Service {
	return &Service{
		repo: infrastructure.NewRepository(ctx),
	}
}

func (s *Service) GetSetupStatus(ctx context.Context, ownerID int64) ([]*domain.SetupStep, bool, error) {
	steps, err := s.repo.GetSetupSteps(ctx, ownerID)
	if err != nil {
		return nil, false, err
	}

	completedCount := 0
	for _, step := range steps {
		if step.CompletedAt != nil {
			completedCount++
		}
	}

	isComplete := completedCount >= len(validStepKeys)

	return steps, isComplete, nil
}

func (s *Service) CompleteSetupStep(ctx context.Context, ownerID int64, stepKey string) error {
	if _, ok := validStepKeys[stepKey]; !ok {
		return domain.ErrInvalidStepKey
	}

	return s.repo.UpsertSetupStep(ctx, ownerID, stepKey, time.Now())
}

func (s *Service) GetHRConfig(ctx context.Context, ownerID int64) (*domain.HRConfig, error) {
	return s.repo.GetHRConfig(ctx, ownerID)
}

func (s *Service) UpsertHRConfig(ctx context.Context, ownerID int64, setupCompletedAt *int64) (*domain.HRConfig, error) {
	var completedAt *time.Time
	if setupCompletedAt != nil {
		t := time.Unix(*setupCompletedAt, 0)
		completedAt = &t
	}

	return s.repo.UpsertHRConfig(ctx, ownerID, completedAt)
}
