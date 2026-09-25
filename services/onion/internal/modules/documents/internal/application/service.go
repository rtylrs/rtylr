package application

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/documents/internal/domain"
	"github.com/voxire/onion/internal/modules/documents/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	documentsScope = "documents"
	documentsTTL   = 5 * time.Minute

	defaultListLimit int32 = 20
	maxListLimit     int32 = 100
)

type repository interface {
	List(ctx context.Context, businessID int64) ([]domain.Document, error)
	ListPaginated(ctx context.Context, businessID int64, limit, offset int32) ([]domain.Document, error)
	Create(ctx context.Context, businessID int64, input domain.DocumentInput) (int64, error)
	Delete(ctx context.Context, businessID, id int64) error
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func invalidateDocumentsScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(documentsScope, businessID))
}

func (s *Service) List(ctx context.Context, businessID int64) ([]domain.Document, error) {
	return cacher.Of[[]domain.Document]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(documentsScope, businessID, "list"),
		func(ctx context.Context) ([]domain.Document, error) {
			return s.repo.List(ctx, businessID)
		},
		cacher.WithExpiry(ctx, documentsTTL),
	)
}

// ListPaginated returns a single page of documents. A non-positive limit
// defaults to 20 and limit is capped at 100; a negative offset is clamped to
// zero.
func (s *Service) ListPaginated(ctx context.Context, businessID int64, limit, offset int32) ([]domain.Document, error) {
	if limit <= 0 {
		limit = defaultListLimit
	}
	if limit > maxListLimit {
		limit = maxListLimit
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.ListPaginated(ctx, businessID, limit, offset)
}

func (s *Service) Create(ctx context.Context, businessID int64, input domain.DocumentInput) (int64, error) {
	id, err := s.repo.Create(ctx, businessID, input)
	if err != nil {
		return 0, err
	}
	invalidateDocumentsScope(ctx, businessID)
	return id, nil
}

func (s *Service) Delete(ctx context.Context, businessID, id int64) error {
	if err := s.repo.Delete(ctx, businessID, id); err != nil {
		return err
	}
	invalidateDocumentsScope(ctx, businessID)
	return nil
}
