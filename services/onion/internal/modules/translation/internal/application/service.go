package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/translation/internal/domain"
	"github.com/voxire/onion/internal/modules/translation/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/logger"
)

type Service struct {
	repository *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	repo := infrastructure.NewRepository(ctx)

	return &Service{
		repository: repo,
	}
}

func (s *Service) GetTranslation(ctx context.Context, keyText string) (*domain.Translation, bool) {
	return s.repository.GetTranslation(ctx, keyText)
}

func (s *Service) CreateOrUpdateTranslation(ctx context.Context, input domain.TranslationInput) (string, bool) {
	if input.KeyText == "" || input.En == "" {
		return "", false
	}

	success := s.repository.UpsertTranslation(ctx, input)
	if !success {
		logger.Ctx(ctx).Error("failed to create/update translation")

		return "", false
	}

	return input.KeyText, true
}
