package application

import (
	"context"
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

func NewService(ctx context.Context) *Service {
	return &Service{
		repo: infrastructure.NewRepository(ctx),
	}
}

func invalidateTemplatesScope(ctx context.Context, ownerID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(templatesScope, ownerID))
	}
}

func (s *Service) List(ctx context.Context, ownerID int64) ([]domain.Template, error) {
	return cacher.Of[[]domain.Template]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(templatesScope, ownerID, "list"),
		func(ctx context.Context) ([]domain.Template, error) {
			return s.repo.List(ctx, ownerID)
		},
		cacher.WithExpiry(ctx, templatesTTL),
	)
}

func (s *Service) Create(ctx context.Context, input domain.TemplateInput) (int64, error) {
	id, err := s.repo.Create(ctx, input)
	if err != nil {
		return 0, err
	}
	invalidateTemplatesScope(ctx, input.OwnerID)
	return id, nil
}

func (s *Service) Update(ctx context.Context, input domain.TemplateUpdateInput) error {
	if err := s.repo.Update(ctx, input); err != nil {
		return err
	}
	invalidateTemplatesScope(ctx, input.OwnerID)
	return nil
}

func (s *Service) Delete(ctx context.Context, ownerID, templateID int64) error {
	if err := s.repo.Delete(ctx, ownerID, templateID); err != nil {
		return err
	}
	invalidateTemplatesScope(ctx, ownerID)
	return nil
}
