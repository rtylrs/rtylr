package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/hr/training/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	trainingScope = "training"
	trainingTTL   = 5 * time.Minute
)

type repository interface {
	List(ctx context.Context, businessID int64) ([]domain.TrainingCourse, error)
	Create(ctx context.Context, businessID int64, input domain.TrainingCourseInput) (int64, error)
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, repo repository) *Service {
	return &Service{repo: repo}
}

func invalidateTrainingScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(trainingScope, businessID))
}

func (s *Service) List(ctx context.Context, businessID int64) ([]domain.TrainingCourse, error) {
	return cacher.Of[[]domain.TrainingCourse]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(trainingScope, businessID, "list")+"|domain=v2",
		func(ctx context.Context) ([]domain.TrainingCourse, error) {
			return s.repo.List(ctx, businessID)
		},
		cacher.WithExpiry(ctx, trainingTTL),
	)
}

func (s *Service) Create(ctx context.Context, businessID int64, input domain.TrainingCourseInput) (int64, error) {
	id, err := s.repo.Create(ctx, businessID, input)
	if err != nil {
		return 0, err
	}
	invalidateTrainingScope(ctx, businessID)
	return id, nil
}
