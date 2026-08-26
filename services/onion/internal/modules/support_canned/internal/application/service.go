package application

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/support_canned/internal/domain"
	"github.com/voxire/onion/internal/modules/support_canned/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	supportCannedScope = "support_canned"
	supportCannedTTL   = 10 * time.Minute
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{
		repo: infrastructure.NewRepository(db),
	}
}

func invalidateSupportCannedScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(supportCannedScope, businessID))
}

func (s *Service) List(ctx context.Context, businessID int64) ([]domain.CannedResponse, error) {
	return cacher.Of[[]domain.CannedResponse]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(supportCannedScope, businessID, "list"),
		func(ctx context.Context) ([]domain.CannedResponse, error) {
			return s.repo.List(ctx, businessID)
		},
		cacher.WithExpiry(ctx, supportCannedTTL),
	)
}

func (s *Service) Create(ctx context.Context, input domain.CannedResponseInput) (int64, error) {
	id, err := s.repo.Create(ctx, input)
	if err != nil {
		return 0, err
	}
	invalidateSupportCannedScope(ctx, input.BusinessID)
	return id, nil
}

func (s *Service) Update(ctx context.Context, businessID, id int64, input domain.CannedResponseInput) error {
	if err := s.repo.Update(ctx, businessID, id, input); err != nil {
		return err
	}
	invalidateSupportCannedScope(ctx, businessID)
	return nil
}

func (s *Service) Delete(ctx context.Context, businessID, id int64) error {
	if err := s.repo.Delete(ctx, businessID, id); err != nil {
		return err
	}
	invalidateSupportCannedScope(ctx, businessID)
	return nil
}
