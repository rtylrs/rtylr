package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/hr_onboarding/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_onboarding/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	hrOnboardingScope = "hr_onboarding"
	hrOnboardingTTL   = 5 * time.Minute
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{
		repo: infrastructure.NewRepository(ctx),
	}
}

func invalidateHROnboardingScope(ctx context.Context, ownerID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(hrOnboardingScope, ownerID))
	}
}

type onboardingListResult[T any] struct {
	Items []T
	Total int64
}

func (s *Service) ListTemplates(ctx context.Context, ownerID int64, params domain.ListTemplatesParams) ([]domain.ChecklistTemplateRecord, int64, error) {
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	key := cacheKeys.Scoped(hrOnboardingScope, ownerID, fmt.Sprintf("tmpl|l=%d|o=%d", params.Limit, params.Offset))
	res, err := cacher.Of[onboardingListResult[domain.ChecklistTemplateRecord]]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (onboardingListResult[domain.ChecklistTemplateRecord], error) {
			items, total, err := s.repo.ListTemplates(ctx, ownerID, params.Limit, params.Offset)
			return onboardingListResult[domain.ChecklistTemplateRecord]{Items: items, Total: total}, err
		},
		cacher.WithExpiry(ctx, hrOnboardingTTL),
	)
	return res.Items, res.Total, err
}

func (s *Service) CreateTemplate(ctx context.Context, ownerID int64, params domain.CreateTemplateParams) (*domain.ChecklistTemplateRecord, error) {
	if params.TriggerType == "" {
		params.TriggerType = "onboarding"
	}
	rec, err := s.repo.CreateTemplate(ctx, ownerID, params)
	if err != nil {
		return nil, err
	}
	invalidateHROnboardingScope(ctx, ownerID)
	return rec, nil
}

func (s *Service) UpdateTemplate(ctx context.Context, ownerID, id int64, params domain.UpdateTemplateParams) error {
	if _, err := s.repo.GetTemplate(ctx, ownerID, id); err != nil {
		return err
	}
	if err := s.repo.UpdateTemplate(ctx, ownerID, id, params); err != nil {
		return err
	}
	invalidateHROnboardingScope(ctx, ownerID)
	return nil
}

func (s *Service) AddItemToTemplate(ctx context.Context, ownerID, templateID int64, params domain.CreateTemplateItemParams) (*domain.ChecklistItemRecord, error) {
	if _, err := s.repo.GetTemplate(ctx, ownerID, templateID); err != nil {
		return nil, err
	}
	if params.AssigneeRole == "" {
		params.AssigneeRole = "employee"
	}
	rec, err := s.repo.CreateTemplateItem(ctx, ownerID, templateID, params)
	if err != nil {
		return nil, err
	}
	invalidateHROnboardingScope(ctx, ownerID)
	return rec, nil
}

func (s *Service) TriggerChecklist(ctx context.Context, ownerID, employeeID, templateID int64, triggerDate time.Time) (*domain.EmployeeChecklistRecord, error) {
	if _, err := s.repo.GetTemplate(ctx, ownerID, templateID); err != nil {
		return nil, err
	}
	rec, err := s.repo.TriggerChecklist(ctx, ownerID, employeeID, templateID, triggerDate)
	if err != nil {
		return nil, err
	}
	invalidateHROnboardingScope(ctx, ownerID)
	return rec, nil
}

func (s *Service) CompleteChecklistItem(ctx context.Context, ownerID, itemID, completedBy int64) error {
	item, err := s.repo.GetChecklistItem(ctx, ownerID, itemID)
	if err != nil {
		return err
	}
	if item.Status == "complete" {
		return domain.ErrChecklistAlreadyCompleted
	}
	if err := s.repo.CompleteChecklistItem(ctx, ownerID, itemID, completedBy); err != nil {
		return err
	}
	total, completed, err := s.repo.GetChecklistCompletionStats(ctx, ownerID, item.EmployeeChecklistID)
	if err != nil {
		return err
	}
	if total > 0 && total == completed {
		if err := s.repo.UpdateEmployeeChecklistStatus(ctx, ownerID, item.EmployeeChecklistID, "completed"); err != nil {
			return err
		}
	}
	invalidateHROnboardingScope(ctx, ownerID)
	return nil
}

func (s *Service) GetEmployeeChecklist(ctx context.Context, ownerID, employeeID, checklistID int64) (*domain.EmployeeChecklistRecord, error) {
	key := cacheKeys.Scoped(hrOnboardingScope, ownerID, fmt.Sprintf("cklst|e=%d|c=%d", employeeID, checklistID))
	return cacher.Of[*domain.EmployeeChecklistRecord]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (*domain.EmployeeChecklistRecord, error) {
			return s.repo.GetEmployeeChecklist(ctx, ownerID, employeeID, checklistID)
		},
		cacher.WithExpiry(ctx, hrOnboardingTTL),
	)
}

func (s *Service) ListEmployeeChecklists(ctx context.Context, ownerID, employeeID int64, params domain.ListEmployeeChecklistsParams) ([]domain.EmployeeChecklistRecord, int64, error) {
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	key := cacheKeys.Scoped(hrOnboardingScope, ownerID, fmt.Sprintf("emp_cklsts|e=%d|l=%d|o=%d", employeeID, params.Limit, params.Offset))
	res, err := cacher.Of[onboardingListResult[domain.EmployeeChecklistRecord]]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (onboardingListResult[domain.EmployeeChecklistRecord], error) {
			items, total, err := s.repo.ListEmployeeChecklists(ctx, ownerID, employeeID, params.Limit, params.Offset)
			return onboardingListResult[domain.EmployeeChecklistRecord]{Items: items, Total: total}, err
		},
		cacher.WithExpiry(ctx, hrOnboardingTTL),
	)
	return res.Items, res.Total, err
}
