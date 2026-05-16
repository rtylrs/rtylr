package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/feedback/internal/domain"
	"github.com/voxire/onion/internal/modules/feedback/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	feedbackScope = "feedback"
	feedbackTTL   = 2 * time.Minute
)

type repository interface {
	List(ctx context.Context, ownerID int64, filters domain.ListFilters, page, pageSize int) ([]domain.Feedback, int, error)
	Create(ctx context.Context, input domain.FeedbackInput) (int64, error)
}

type Service struct {
	repo repository
}

func NewService(ctx context.Context) *Service {
	return &Service{
		repo: infrastructure.NewRepository(ctx),
	}
}

func invalidateFeedbackScope(ctx context.Context, ownerID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(feedbackScope, ownerID))
	}
}

type feedbackListResult struct {
	Items []domain.Feedback
	Total int
}

func (s *Service) List(ctx context.Context, ownerID int64, filters domain.ListFilters, page, pageSize int) ([]domain.Feedback, int, error) {
	key := cacheKeys.Scoped(feedbackScope, ownerID, fmt.Sprintf("list|f=%+v|p=%d|ps=%d", filters, page, pageSize))
	res, err := cacher.Of[feedbackListResult]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (feedbackListResult, error) {
			items, total, err := s.repo.List(ctx, ownerID, filters, page, pageSize)
			return feedbackListResult{Items: items, Total: total}, err
		},
		cacher.WithExpiry(ctx, feedbackTTL),
	)
	return res.Items, res.Total, err
}

func (s *Service) Create(ctx context.Context, input domain.FeedbackInput) (int64, error) {
	id, err := s.repo.Create(ctx, input)
	if err != nil {
		return 0, err
	}
	invalidateFeedbackScope(ctx, input.OwnerID)
	return id, nil
}
