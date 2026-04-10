package application

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/voxire/onion/internal/modules/integrations/internal/domain"
	"github.com/voxire/onion/internal/modules/integrations/internal/infrastructure"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{
		repo: infrastructure.NewRepository(ctx),
	}
}

func (s *Service) List(ctx context.Context, ownerID int64) ([]domain.Integration, error) {
	return s.repo.List(ctx, ownerID)
}

func (s *Service) Connect(ctx context.Context, ownerID int64, provider string, config json.RawMessage) (int64, error) {
	if provider == "" {
		return 0, errors.New("provider is required")
	}

	if config == nil {
		config = json.RawMessage("{}")
	}

	return s.repo.Connect(ctx, ownerID, provider, config)
}

func (s *Service) Disconnect(ctx context.Context, ownerID, integrationID int64) error {
	if integrationID == 0 {
		return errors.New("integration id is required")
	}
	return s.repo.Disconnect(ctx, ownerID, integrationID)
}
