package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/documents/internal/domain"
	"github.com/voxire/onion/internal/modules/documents/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	documentsScope = "documents"
	documentsTTL   = 5 * time.Minute
)

type repository interface {
	List(ctx context.Context, ownerID int64) ([]domain.Document, error)
	Create(ctx context.Context, ownerID int64, input domain.DocumentInput) (int64, error)
	Delete(ctx context.Context, ownerID, id int64) error
}

type Service struct {
	repo repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func invalidateDocumentsScope(ctx context.Context, ownerID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(documentsScope, ownerID))
	}
}

func (s *Service) List(ctx context.Context, ownerID int64) ([]domain.Document, error) {
	return cacher.Of[[]domain.Document]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(documentsScope, ownerID, "list"),
		func(ctx context.Context) ([]domain.Document, error) {
			return s.repo.List(ctx, ownerID)
		},
		cacher.WithExpiry(ctx, documentsTTL),
	)
}

func (s *Service) Create(ctx context.Context, ownerID int64, input domain.DocumentInput) (int64, error) {
	id, err := s.repo.Create(ctx, ownerID, input)
	if err != nil {
		return 0, err
	}
	invalidateDocumentsScope(ctx, ownerID)
	return id, nil
}

func (s *Service) Delete(ctx context.Context, ownerID, id int64) error {
	if err := s.repo.Delete(ctx, ownerID, id); err != nil {
		return err
	}
	invalidateDocumentsScope(ctx, ownerID)
	return nil
}
