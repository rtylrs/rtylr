package application

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/segments/internal/domain"
	"github.com/voxire/onion/internal/modules/segments/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	segmentsScope = "segments"
	segmentsTTL   = 5 * time.Minute
)

type repository interface {
	List(ctx context.Context, businessID int64) ([]domain.Segment, error)
	Create(ctx context.Context, input domain.SegmentInput) (int64, error)
	Update(ctx context.Context, input domain.SegmentUpdateInput) error
	Delete(ctx context.Context, businessID, id int64) error
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func invalidateSegmentsScope(ctx context.Context, businessID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(segmentsScope, businessID))
	}
}

func (s *Service) List(ctx context.Context, businessID int64) ([]domain.Segment, error) {
	return cacher.Of[[]domain.Segment]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(segmentsScope, businessID, "list"),
		func(ctx context.Context) ([]domain.Segment, error) {
			return s.repo.List(ctx, businessID)
		},
		cacher.WithExpiry(ctx, segmentsTTL),
	)
}

func (s *Service) Create(ctx context.Context, input domain.SegmentInput) (int64, error) {
	id, err := s.repo.Create(ctx, input)
	if err != nil {
		return 0, err
	}
	invalidateSegmentsScope(ctx, input.BusinessID)
	return id, nil
}

func (s *Service) Update(ctx context.Context, input domain.SegmentUpdateInput) error {
	if err := s.repo.Update(ctx, input); err != nil {
		return err
	}
	invalidateSegmentsScope(ctx, input.BusinessID)
	return nil
}

func (s *Service) Delete(ctx context.Context, businessID, id int64) error {
	if err := s.repo.Delete(ctx, businessID, id); err != nil {
		return err
	}
	invalidateSegmentsScope(ctx, businessID)
	return nil
}
