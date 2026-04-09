package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/voxire/onion/internal/platform/snips"
	"github.com/voxire/onion/internal/modules/echo/internal/domain"
)

var ErrClientNotInitialized = errors.New("AI client not initialized")

type Service struct {
	aiClient *snips.Client
}

func NewService(ctx context.Context) (*Service, error) {
	client, err := snips.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create AI client: %w", err)
	}

	return &Service{
		aiClient: client,
	}, nil
}

func (s *Service) GenerateTextStream(ctx context.Context, prompt string) (<-chan snips.StreamToken, error) {
	if s.aiClient == nil {
		return nil, ErrClientNotInitialized
	}

	return s.aiClient.CompleteTextStream(ctx, prompt)
}

func (s *Service) IsAvailable() bool {
	return s.aiClient != nil
}

func (s *Service) GetProviderStatus() map[string]bool {
	if s.aiClient == nil {
		return map[string]bool{}
	}

	return s.aiClient.GetProviderStatus()
}

func (s *Service) GetPrimaryProvider() string {
	if s.aiClient == nil {
		return ""
	}

	return s.aiClient.GetPrimaryProvider()
}

func (s *Service) GetProviderDetails(_ context.Context) domain.ProviderStatus {
	if s.aiClient == nil {
		return domain.ProviderStatus{
			Available:       false,
			PrimaryProvider: "",
			Providers:       map[string]bool{},
		}
	}

	return domain.ProviderStatus{
		Available:       s.aiClient != nil,
		PrimaryProvider: s.aiClient.GetPrimaryProvider(),
		Providers:       s.aiClient.GetProviderStatus(),
	}
}
