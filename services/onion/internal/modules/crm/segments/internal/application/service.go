package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/crm/segments/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	segmentsScope = "segments"
	segmentsTTL   = 5 * time.Minute
)

type repository interface {
	List(ctx context.Context, businessID int64) ([]domain.Segment, error)
	GetByID(ctx context.Context, businessID, id int64) (domain.Segment, error)
	Create(ctx context.Context, input domain.SegmentInput) (int64, error)
	Update(ctx context.Context, input domain.SegmentUpdateInput) error
	Delete(ctx context.Context, businessID, id int64) error
	ListCustomers(ctx context.Context, businessID int64, filters []byte, page, pageSize int) ([]domain.SegmentCustomer, int, error)
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, repo repository) *Service {
	return &Service{repo: repo}
}

func invalidateSegmentsScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(segmentsScope, businessID))
}

func (s *Service) List(ctx context.Context, businessID int64) ([]domain.Segment, error) {
	return cacher.Of[[]domain.Segment]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(segmentsScope, businessID, "list")+"|domain=v2",
		func(ctx context.Context) ([]domain.Segment, error) {
			return s.repo.List(ctx, businessID)
		},
		cacher.WithExpiry(ctx, segmentsTTL),
	)
}

func (s *Service) Get(ctx context.Context, businessID, id int64) (domain.Segment, error) {
	return s.repo.GetByID(ctx, businessID, id)
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

func (s *Service) GetDetail(ctx context.Context, businessID, id int64) (domain.SegmentDetail, error) {
	segment, err := s.Get(ctx, businessID, id)
	if err != nil {
		return domain.SegmentDetail{}, err
	}

	customers, _, err := s.repo.ListCustomers(ctx, businessID, segment.Filters, 1, 50)
	if err != nil {
		return domain.SegmentDetail{}, err
	}

	return domain.SegmentDetail{Segment: segment, Customers: customers}, nil
}

func (s *Service) GetCustomers(ctx context.Context, businessID, id int64, page, pageSize int) ([]domain.SegmentCustomer, int, error) {
	segment, err := s.Get(ctx, businessID, id)
	if err != nil {
		return nil, 0, err
	}

	return s.repo.ListCustomers(ctx, businessID, segment.Filters, page, pageSize)
}
