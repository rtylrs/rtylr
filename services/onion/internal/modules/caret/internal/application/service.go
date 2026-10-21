package application

import (
	"context"
	"errors"

	"github.com/voxire/onion/internal/modules/caret/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/snips"
)

var ErrClientNotInitialized = errors.New("AI client not initialized")

type Service struct {
	aiClient *snips.Client
}

func NewService(ctx context.Context) (*Service, error) {
	client, err := snips.NewClient(ctx)
	if err != nil {
		logger.Ctx(ctx).Debug("AI client not initialized; AI features disabled", logger.Error(err))
		return &Service{aiClient: nil}, nil
	}

	return &Service{
		aiClient: client,
	}, nil
}

func (s *Service) GenerateTextStream(ctx context.Context, prompt string) (<-chan snips.StreamToken, error) {
	if s == nil || s.aiClient == nil {
		return nil, ErrClientNotInitialized
	}

	return s.aiClient.CompleteTextStream(ctx, prompt)
}

// CompleteText is the non-streaming completion used for internal steps like
// generating a structured query spec in Analyst mode.
func (s *Service) CompleteText(ctx context.Context, prompt string) (string, error) {
	if s == nil || s.aiClient == nil {
		return "", ErrClientNotInitialized
	}

	return s.aiClient.CompleteText(ctx, prompt)
}

func (s *Service) CompleteTextWithOptions(ctx context.Context, prompt string, opts snips.CompletionOptions) (string, error) {
	if s == nil || s.aiClient == nil {
		return "", ErrClientNotInitialized
	}

	return s.aiClient.CompleteTextWithOptions(ctx, prompt, opts)
}

func (s *Service) IsAvailable() bool {
	return s != nil && s.aiClient != nil
}

func (s *Service) GetProviderStatus() map[string]bool {
	if s == nil || s.aiClient == nil {
		return map[string]bool{}
	}

	return s.aiClient.GetProviderStatus()
}

func (s *Service) GetPrimaryProvider() string {
	if s == nil || s.aiClient == nil {
		return ""
	}

	return s.aiClient.GetPrimaryProvider()
}

func (s *Service) GetProviderDetails(_ context.Context) domain.ProviderStatus {
	if s == nil || s.aiClient == nil {
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
