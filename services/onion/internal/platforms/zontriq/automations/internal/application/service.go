package application

import (
	"context"

	"github.com/voxire/onion/internal/platforms/zontriq/automations/internal/domain"
	"github.com/voxire/onion/internal/platforms/zontriq/automations/internal/infrastructure"
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

func (s *Service) ListAutomations(ctx context.Context, zontriqUserID int, state *string, pageSize, pageOffset int) ([]domain.Automation, int64, error) {
	automations, err := s.repo.ListAutomationsByUserID(ctx, zontriqUserID, state, pageSize, pageOffset)
	if err != nil {
		return nil, 0, err
	}

	count, err := s.repo.CountAutomationsByUserID(ctx, zontriqUserID, state)
	if err != nil {
		return nil, 0, err
	}

	return automations, count, nil
}

func (s *Service) GetAutomation(ctx context.Context, id int) (*domain.AutomationDetail, []domain.Channel, error) {
	automation, err := s.repo.GetAutomationByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	channels, err := s.repo.ListAutomationChannels(ctx, int(automation.ID))
	if err != nil {
		return nil, nil, err
	}

	return automation, channels, nil
}

func (s *Service) CreateAutomation(ctx context.Context, zontriqUserID, seriesID int, name, frequency string, channelIDs []int64) (*domain.AutomationRow, error) {
	automation, err := s.repo.CreateAutomation(ctx, zontriqUserID, seriesID, name, frequency)
	if err != nil {
		return nil, err
	}

	for _, channelID := range channelIDs {
		if err := s.repo.CreateAutomationChannel(ctx, int(automation.ID), int(channelID)); err != nil {
			return nil, err
		}
	}

	return automation, nil
}

func (s *Service) UpdateAutomation(ctx context.Context, id int, name, frequency, state string, channelIDs []int64) (*domain.AutomationUpdatedRow, error) {
	automation, err := s.repo.UpdateAutomation(ctx, id, name, frequency, state)
	if err != nil {
		return nil, err
	}

	if err := s.repo.DeleteAutomationChannels(ctx, int(automation.ID)); err != nil {
		return nil, err
	}

	for _, channelID := range channelIDs {
		if err := s.repo.CreateAutomationChannel(ctx, int(automation.ID), int(channelID)); err != nil {
			return nil, err
		}
	}

	return automation, nil
}

func (s *Service) DeleteAutomation(ctx context.Context, id int) error {
	if err := s.repo.DeleteAutomationChannels(ctx, id); err != nil {
		return err
	}

	return s.repo.SoftDeleteAutomation(ctx, id)
}

func (s *Service) GetDueAutomations(ctx context.Context) ([]domain.DueAutomation, error) {
	return s.repo.GetDueAutomations(ctx)
}

func (s *Service) CountAutomationsForQuota(ctx context.Context, zontriqUserID int) (int64, error) {
	return s.repo.CountAutomationsByUserIDForQuota(ctx, zontriqUserID)
}

func (s *Service) GetZontriqUserIDByUserID(ctx context.Context, userID int) (int, error) {
	return s.repo.GetZontriqUserIDByUserID(ctx, userID)
}
