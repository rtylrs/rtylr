package application

import (
	"context"
	"fmt"

	"github.com/voxire/onion/internal/platforms/zontriq/channels/internal/domain"
	"github.com/voxire/onion/internal/platforms/zontriq/channels/internal/infrastructure"
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

func (s *Service) ListChannelsByUserID(ctx context.Context, zontriqUserID int32, platform string) ([]domain.Channel, error) {
	channels, err := s.repo.ListChannelsByUserID(ctx, zontriqUserID, platform)
	if err != nil {
		return nil, fmt.Errorf("failed to list channels: %w", err)
	}

	return channels, nil
}

func (s *Service) GetChannelByID(ctx context.Context, channelID int32) (domain.Channel, error) {
	channel, err := s.repo.GetChannelByID(ctx, channelID)
	if err != nil {
		return domain.Channel{}, fmt.Errorf("failed to get channel: %w", err)
	}

	return channel, nil
}

func (s *Service) ConnectChannel(ctx context.Context, input infrastructure.CreateChannelInput) (domain.Channel, error) {
	channel, err := s.repo.CreateChannel(ctx, input)
	if err != nil {
		return domain.Channel{}, fmt.Errorf("failed to connect channel: %w", err)
	}

	return channel, nil
}

func (s *Service) DisconnectChannel(ctx context.Context, channelID int32) error {
	err := s.repo.SoftDeleteChannel(ctx, channelID)
	if err != nil {
		return fmt.Errorf("failed to disconnect channel: %w", err)
	}

	return nil
}

func (s *Service) RefreshChannelToken(ctx context.Context, input infrastructure.UpdateTokensInput) (domain.Channel, error) {
	channel, err := s.repo.UpdateChannelTokens(ctx, input)
	if err != nil {
		return domain.Channel{}, fmt.Errorf("failed to refresh channel token: %w", err)
	}

	return channel, nil
}

func (s *Service) GetZontriqUserID(ctx context.Context, userID int64) (int32, error) {
	zontriqUserID, err := s.repo.GetZontriqUserID(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get zontriq user ID: %w", err)
	}

	return zontriqUserID, nil
}

func (s *Service) CountChannelsByUserID(ctx context.Context, zontriqUserID int32) (int64, error) {
	count, err := s.repo.CountChannelsByUserID(ctx, zontriqUserID)
	if err != nil {
		return 0, fmt.Errorf("failed to count channels: %w", err)
	}

	return count, nil
}
