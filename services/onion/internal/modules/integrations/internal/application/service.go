package application

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/voxire/onion/internal/modules/integrations/internal/domain"
	"github.com/voxire/onion/internal/modules/integrations/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	integrationsScope = "integrations"
	integrationsTTL   = 5 * time.Minute
)

type repository interface {
	List(ctx context.Context, ownerID int64) ([]domain.Integration, error)
	Connect(ctx context.Context, ownerID int64, provider string, config json.RawMessage) (int64, error)
	Disconnect(ctx context.Context, ownerID, integrationID int64) error
}

type Service struct {
	repo repository
}

func NewService(ctx context.Context) *Service {
	return &Service{
		repo: infrastructure.NewRepository(ctx),
	}
}

func invalidateIntegrationsScope(ctx context.Context, ownerID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(integrationsScope, ownerID))
	}
}

func (s *Service) List(ctx context.Context, ownerID int64) ([]domain.Integration, error) {
	return cacher.Of[[]domain.Integration]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(integrationsScope, ownerID, "list"),
		func(ctx context.Context) ([]domain.Integration, error) {
			return s.repo.List(ctx, ownerID)
		},
		cacher.WithExpiry(ctx, integrationsTTL),
	)
}

func (s *Service) Connect(ctx context.Context, ownerID int64, provider string, config json.RawMessage) (int64, error) {
	if provider == "" {
		return 0, errors.New("provider is required")
	}
	if config == nil {
		config = json.RawMessage("{}")
	}
	id, err := s.repo.Connect(ctx, ownerID, provider, config)
	if err != nil {
		return 0, err
	}
	invalidateIntegrationsScope(ctx, ownerID)
	return id, nil
}

func (s *Service) Disconnect(ctx context.Context, ownerID, integrationID int64) error {
	if integrationID == 0 {
		return errors.New("integration id is required")
	}
	if err := s.repo.Disconnect(ctx, ownerID, integrationID); err != nil {
		return err
	}
	invalidateIntegrationsScope(ctx, ownerID)
	return nil
}
