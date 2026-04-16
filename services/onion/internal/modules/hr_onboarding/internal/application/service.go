package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/hr_onboarding/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_onboarding/internal/infrastructure"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{
		repo: infrastructure.NewRepository(ctx),
	}
}

func (s *Service) ListTemplates(ctx context.Context, ownerID int64, params domain.ListTemplatesParams) ([]domain.ChecklistTemplateRecord, int64, error) {
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}

	return s.repo.ListTemplates(ctx, ownerID, params.Limit, params.Offset)
}

func (s *Service) CreateTemplate(ctx context.Context, ownerID int64, params domain.CreateTemplateParams) (*domain.ChecklistTemplateRecord, error) {
	if params.TriggerType == "" {
		params.TriggerType = "onboarding"
	}

	return s.repo.CreateTemplate(ctx, ownerID, params)
}

func (s *Service) UpdateTemplate(ctx context.Context, ownerID, id int64, params domain.UpdateTemplateParams) error {
	_, err := s.repo.GetTemplate(ctx, ownerID, id)
	if err != nil {
		return err
	}

	return s.repo.UpdateTemplate(ctx, ownerID, id, params)
}

func (s *Service) AddItemToTemplate(ctx context.Context, ownerID, templateID int64, params domain.CreateTemplateItemParams) (*domain.ChecklistItemRecord, error) {
	_, err := s.repo.GetTemplate(ctx, ownerID, templateID)
	if err != nil {
		return nil, err
	}

	if params.AssigneeRole == "" {
		params.AssigneeRole = "employee"
	}

	return s.repo.CreateTemplateItem(ctx, ownerID, templateID, params)
}

func (s *Service) TriggerChecklist(ctx context.Context, ownerID, employeeID, templateID int64, triggerDate time.Time) (*domain.EmployeeChecklistRecord, error) {
	_, err := s.repo.GetTemplate(ctx, ownerID, templateID)
	if err != nil {
		return nil, err
	}

	return s.repo.TriggerChecklist(ctx, ownerID, employeeID, templateID, triggerDate)
}

func (s *Service) CompleteChecklistItem(ctx context.Context, ownerID, itemID, completedBy int64) error {
	item, err := s.repo.GetChecklistItem(ctx, ownerID, itemID)
	if err != nil {
		return err
	}

	if item.Status == "complete" {
		return domain.ErrChecklistAlreadyCompleted
	}

	err = s.repo.CompleteChecklistItem(ctx, ownerID, itemID, completedBy)
	if err != nil {
		return err
	}

	total, completed, err := s.repo.GetChecklistCompletionStats(ctx, ownerID, item.EmployeeChecklistID)
	if err != nil {
		return err
	}

	if total > 0 && total == completed {
		return s.repo.UpdateEmployeeChecklistStatus(ctx, ownerID, item.EmployeeChecklistID, "completed")
	}

	return nil
}

func (s *Service) GetEmployeeChecklist(ctx context.Context, ownerID, employeeID, checklistID int64) (*domain.EmployeeChecklistRecord, error) {
	return s.repo.GetEmployeeChecklist(ctx, ownerID, employeeID, checklistID)
}

func (s *Service) ListEmployeeChecklists(ctx context.Context, ownerID, employeeID int64, params domain.ListEmployeeChecklistsParams) ([]domain.EmployeeChecklistRecord, int64, error) {
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}

	return s.repo.ListEmployeeChecklists(ctx, ownerID, employeeID, params.Limit, params.Offset)
}
