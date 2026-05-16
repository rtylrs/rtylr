package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/training/internal/domain"
	"github.com/voxire/onion/internal/modules/training/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	trainingScope = "training"
	trainingTTL   = 5 * time.Minute
)

type repository interface {
	List(ctx context.Context, ownerID int64) ([]domain.TrainingCourse, error)
	Create(ctx context.Context, ownerID int64, input domain.TrainingCourseInput) (int64, error)
}

type Service struct {
	repo repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func invalidateTrainingScope(ctx context.Context, ownerID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(trainingScope, ownerID))
	}
}

func (s *Service) List(ctx context.Context, ownerID int64) ([]domain.TrainingCourse, error) {
	return cacher.Of[[]domain.TrainingCourse]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(trainingScope, ownerID, "list"),
		func(ctx context.Context) ([]domain.TrainingCourse, error) {
			return s.repo.List(ctx, ownerID)
		},
		cacher.WithExpiry(ctx, trainingTTL),
	)
}

func (s *Service) Create(ctx context.Context, ownerID int64, input domain.TrainingCourseInput) (int64, error) {
	id, err := s.repo.Create(ctx, ownerID, input)
	if err != nil {
		return 0, err
	}
	invalidateTrainingScope(ctx, ownerID)
	return id, nil
}
