package application

import (
	"context"
	"database/sql"
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
	List(ctx context.Context, businessID int64) ([]domain.Integration, error)
	Connect(ctx context.Context, businessID int64, provider string, config json.RawMessage) (int64, error)
	Disconnect(ctx context.Context, businessID, integrationID int64) error
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{
		repo: infrastructure.NewRepository(db),
	}
}

func invalidateIntegrationsScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(integrationsScope, businessID))
}

func (s *Service) List(ctx context.Context, businessID int64) ([]domain.Integration, error) {
	return cacher.Of[[]domain.Integration]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(integrationsScope, businessID, "list"),
		func(ctx context.Context) ([]domain.Integration, error) {
			return s.repo.List(ctx, businessID)
		},
		cacher.WithExpiry(ctx, integrationsTTL),
	)
}

func (s *Service) Connect(ctx context.Context, businessID int64, provider string, config json.RawMessage) (int64, error) {
	if provider == "" {
		return 0, errors.New("provider is required")
	}
	if config == nil {
		config = json.RawMessage("{}")
	}
	id, err := s.repo.Connect(ctx, businessID, provider, config)
	if err != nil {
		return 0, err
	}
	invalidateIntegrationsScope(ctx, businessID)
	return id, nil
}

func (s *Service) Disconnect(ctx context.Context, businessID, integrationID int64) error {
	if integrationID == 0 {
		return errors.New("integration id is required")
	}
	if err := s.repo.Disconnect(ctx, businessID, integrationID); err != nil {
		return err
	}
	invalidateIntegrationsScope(ctx, businessID)
	return nil
}
