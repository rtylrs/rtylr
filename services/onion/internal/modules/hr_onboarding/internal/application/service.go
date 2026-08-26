package application

import (
	"context"
	"database/sql"
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

type repository interface {
	CreateTemplate(ctx context.Context, businessID int64, params domain.CreateTemplateParams) (*domain.ChecklistTemplateRecord, error)
	GetTemplate(ctx context.Context, businessID, id int64) (*domain.ChecklistTemplateRecord, error)
	ListTemplates(ctx context.Context, businessID int64, limit, offset int32) ([]domain.ChecklistTemplateRecord, int64, error)
	UpdateTemplate(ctx context.Context, businessID, id int64, params domain.UpdateTemplateParams) error
	CreateTemplateItem(ctx context.Context, businessID, templateID int64, params domain.CreateTemplateItemParams) (*domain.ChecklistItemRecord, error)
	ListTemplateItems(ctx context.Context, businessID, templateID int64) ([]domain.ChecklistItemRecord, error)
	TriggerChecklist(ctx context.Context, businessID, employeeID, templateID int64, triggerDate time.Time) (*domain.EmployeeChecklistRecord, error)
	GetEmployeeChecklist(ctx context.Context, businessID, employeeID, checklistID int64) (*domain.EmployeeChecklistRecord, error)
	ListEmployeeChecklists(ctx context.Context, businessID, employeeID int64, limit, offset int32) ([]domain.EmployeeChecklistRecord, int64, error)
	UpdateEmployeeChecklistStatus(ctx context.Context, businessID, checklistID int64, status string) error
	CreateEmployeeChecklistItem(ctx context.Context, businessID, employeeChecklistID, checklistItemID, assignedToEmployeeID int64, dueDate string) error
	GetChecklistItem(ctx context.Context, businessID, itemID int64) (*domain.EmployeeChecklistItemRecord, error)
	CompleteChecklistItem(ctx context.Context, businessID, itemID, completedBy int64) error
	GetChecklistCompletionStats(ctx context.Context, businessID, checklistID int64) (total, completed int, err error)
	ListEmployeeChecklistItems(ctx context.Context, businessID, employeeChecklistID int64) ([]domain.EmployeeChecklistItemRecord, error)
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{
		repo: infrastructure.NewRepository(db),
	}
}

func invalidateHROnboardingScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(hrOnboardingScope, businessID))
}

type onboardingListResult[T any] struct {
	Items []T
	Total int64
}

func (s *Service) ListTemplates(ctx context.Context, businessID int64, params domain.ListTemplatesParams) ([]domain.ChecklistTemplateRecord, int64, error) {
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	key := cacheKeys.Scoped(hrOnboardingScope, businessID, fmt.Sprintf("tmpl|l=%d|o=%d", params.Limit, params.Offset))
	res, err := cacher.Of[onboardingListResult[domain.ChecklistTemplateRecord]]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (onboardingListResult[domain.ChecklistTemplateRecord], error) {
			items, total, err := s.repo.ListTemplates(ctx, businessID, params.Limit, params.Offset)
			return onboardingListResult[domain.ChecklistTemplateRecord]{Items: items, Total: total}, err
		},
		cacher.WithExpiry(ctx, hrOnboardingTTL),
	)
	return res.Items, res.Total, err
}

func (s *Service) CreateTemplate(ctx context.Context, businessID int64, params domain.CreateTemplateParams) (*domain.ChecklistTemplateRecord, error) {
	if params.TriggerType == "" {
		params.TriggerType = "onboarding"
	}
	rec, err := s.repo.CreateTemplate(ctx, businessID, params)
	if err != nil {
		return nil, err
	}
	invalidateHROnboardingScope(ctx, businessID)
	return rec, nil
}

func (s *Service) UpdateTemplate(ctx context.Context, businessID, id int64, params domain.UpdateTemplateParams) error {
	if _, err := s.repo.GetTemplate(ctx, businessID, id); err != nil {
		return err
	}
	if err := s.repo.UpdateTemplate(ctx, businessID, id, params); err != nil {
		return err
	}
	invalidateHROnboardingScope(ctx, businessID)
	return nil
}

func (s *Service) AddItemToTemplate(ctx context.Context, businessID, templateID int64, params domain.CreateTemplateItemParams) (*domain.ChecklistItemRecord, error) {
	if _, err := s.repo.GetTemplate(ctx, businessID, templateID); err != nil {
		return nil, err
	}
	if params.AssigneeRole == "" {
		params.AssigneeRole = "employee"
	}
	rec, err := s.repo.CreateTemplateItem(ctx, businessID, templateID, params)
	if err != nil {
		return nil, err
	}
	invalidateHROnboardingScope(ctx, businessID)
	return rec, nil
}

func (s *Service) TriggerChecklist(ctx context.Context, businessID, employeeID, templateID int64, triggerDate time.Time) (*domain.EmployeeChecklistRecord, error) {
	if _, err := s.repo.GetTemplate(ctx, businessID, templateID); err != nil {
		return nil, err
	}
	rec, err := s.repo.TriggerChecklist(ctx, businessID, employeeID, templateID, triggerDate)
	if err != nil {
		return nil, err
	}
	invalidateHROnboardingScope(ctx, businessID)
	return rec, nil
}

func (s *Service) CompleteChecklistItem(ctx context.Context, businessID, itemID, completedBy int64) error {
	item, err := s.repo.GetChecklistItem(ctx, businessID, itemID)
	if err != nil {
		return err
	}
	if item.Status == "complete" {
		return domain.ErrChecklistAlreadyCompleted
	}
	if err := s.repo.CompleteChecklistItem(ctx, businessID, itemID, completedBy); err != nil {
		return err
	}
	total, completed, err := s.repo.GetChecklistCompletionStats(ctx, businessID, item.EmployeeChecklistID)
	if err != nil {
		return err
	}
	if total > 0 && total == completed {
		if err := s.repo.UpdateEmployeeChecklistStatus(ctx, businessID, item.EmployeeChecklistID, "completed"); err != nil {
			return err
		}
	}
	invalidateHROnboardingScope(ctx, businessID)
	return nil
}

func (s *Service) GetEmployeeChecklist(ctx context.Context, businessID, employeeID, checklistID int64) (*domain.EmployeeChecklistRecord, error) {
	key := cacheKeys.Scoped(hrOnboardingScope, businessID, fmt.Sprintf("cklst|e=%d|c=%d", employeeID, checklistID))
	return cacher.Of[*domain.EmployeeChecklistRecord]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (*domain.EmployeeChecklistRecord, error) {
			return s.repo.GetEmployeeChecklist(ctx, businessID, employeeID, checklistID)
		},
		cacher.WithExpiry(ctx, hrOnboardingTTL),
	)
}

func (s *Service) ListEmployeeChecklists(ctx context.Context, businessID, employeeID int64, params domain.ListEmployeeChecklistsParams) ([]domain.EmployeeChecklistRecord, int64, error) {
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	key := cacheKeys.Scoped(hrOnboardingScope, businessID, fmt.Sprintf("emp_cklsts|e=%d|l=%d|o=%d", employeeID, params.Limit, params.Offset))
	res, err := cacher.Of[onboardingListResult[domain.EmployeeChecklistRecord]]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (onboardingListResult[domain.EmployeeChecklistRecord], error) {
			items, total, err := s.repo.ListEmployeeChecklists(ctx, businessID, employeeID, params.Limit, params.Offset)
			return onboardingListResult[domain.EmployeeChecklistRecord]{Items: items, Total: total}, err
		},
		cacher.WithExpiry(ctx, hrOnboardingTTL),
	)
	return res.Items, res.Total, err
}
