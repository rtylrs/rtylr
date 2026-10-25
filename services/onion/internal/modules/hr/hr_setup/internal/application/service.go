package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/hr/hr_setup/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	hrSetupScope = "hr_setup"
	hrSetupTTL   = 5 * time.Minute
)

var validStepKeys = map[string]struct{}{
	"company_profile":   {},
	"work_schedule":     {},
	"org_structure":     {},
	"locations":         {},
	"roles_permissions": {},
	"payroll":           {},
	"notifications":     {},
	"employees":         {},
}

type repository interface {
	GetSetupSteps(ctx context.Context, businessID int64) ([]*domain.SetupStep, error)
	UpsertSetupStep(ctx context.Context, businessID int64, stepKey string, at time.Time) error
	GetHRConfig(ctx context.Context, businessID int64) (*domain.HRConfig, error)
	UpsertHRConfig(ctx context.Context, businessID int64, completedAt *time.Time) (*domain.HRConfig, error)
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, repo repository) *Service {
	return &Service{
		repo: repo,
	}
}

func invalidateHRSetupScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(hrSetupScope, businessID))
}

type setupStatus struct {
	Steps      []*domain.SetupStep
	IsComplete bool
}

func (s *Service) GetSetupStatus(ctx context.Context, businessID int64) ([]*domain.SetupStep, bool, error) {
	res, err := cacher.Of[setupStatus]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(hrSetupScope, businessID, "status")+"|domain=v2",
		func(ctx context.Context) (setupStatus, error) {
			steps, err := s.repo.GetSetupSteps(ctx, businessID)
			if err != nil {
				return setupStatus{}, err
			}
			completed := 0
			for _, step := range steps {
				if step.CompletedAt != nil {
					completed++
				}
			}
			return setupStatus{Steps: steps, IsComplete: completed >= len(validStepKeys)}, nil
		},
		cacher.WithExpiry(ctx, hrSetupTTL),
	)
	return res.Steps, res.IsComplete, err
}

func (s *Service) CompleteSetupStep(ctx context.Context, businessID int64, stepKey string) error {
	if _, ok := validStepKeys[stepKey]; !ok {
		return domain.ErrInvalidStepKey
	}
	if err := s.repo.UpsertSetupStep(ctx, businessID, stepKey, time.Now()); err != nil {
		return err
	}
	invalidateHRSetupScope(ctx, businessID)
	return nil
}

func (s *Service) GetHRConfig(ctx context.Context, businessID int64) (*domain.HRConfig, error) {
	return cacher.Of[*domain.HRConfig]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(hrSetupScope, businessID, "config")+"|domain=v3",
		func(ctx context.Context) (*domain.HRConfig, error) {
			return s.repo.GetHRConfig(ctx, businessID)
		},
		cacher.WithExpiry(ctx, hrSetupTTL),
	)
}

func (s *Service) UpsertHRConfig(ctx context.Context, businessID int64, setupCompletedAt *int64) (*domain.HRConfig, error) {
	var completedAt *time.Time
	if setupCompletedAt != nil {
		t := time.Unix(*setupCompletedAt, 0)
		completedAt = &t
	}
	cfg, err := s.repo.UpsertHRConfig(ctx, businessID, completedAt)
	if err != nil {
		return nil, err
	}
	invalidateHRSetupScope(ctx, businessID)
	return cfg, nil
}
