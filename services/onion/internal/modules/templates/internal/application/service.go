package application

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/templates/internal/domain"
	"github.com/voxire/onion/internal/modules/templates/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	templatesScope = "templates"
	templatesTTL   = 10 * time.Minute
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{
		repo: infrastructure.NewRepository(db),
	}
}

func invalidateTemplatesScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(templatesScope, businessID))
}

func (s *Service) List(ctx context.Context, businessID int64) ([]domain.Template, error) {
	return cacher.Of[[]domain.Template]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(templatesScope, businessID, "list"),
		func(ctx context.Context) ([]domain.Template, error) {
			return s.repo.List(ctx, businessID)
		},
		cacher.WithExpiry(ctx, templatesTTL),
	)
}

func (s *Service) Create(ctx context.Context, input domain.TemplateInput) (int64, error) {
	id, err := s.repo.Create(ctx, input)
	if err != nil {
		return 0, err
	}
	invalidateTemplatesScope(ctx, input.BusinessID)
	return id, nil
}

func (s *Service) Update(ctx context.Context, input domain.TemplateUpdateInput) error {
	if err := s.repo.Update(ctx, input); err != nil {
		return err
	}
	invalidateTemplatesScope(ctx, input.BusinessID)
	return nil
}

func (s *Service) Delete(ctx context.Context, businessID, templateID int64) error {
	if err := s.repo.Delete(ctx, businessID, templateID); err != nil {
		return err
	}
	invalidateTemplatesScope(ctx, businessID)
	return nil
}
