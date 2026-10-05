package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/crm/feedback/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	feedbackScope = "feedback"
	feedbackTTL   = 2 * time.Minute
)

type repository interface {
	List(ctx context.Context, businessID int64, filters domain.ListFilters, page, pageSize int) ([]domain.Feedback, int, error)
	Create(ctx context.Context, input domain.FeedbackInput) (int64, error)
	CustomerExistsInBusiness(ctx context.Context, businessID, customerID int64) bool
}

// CustomerExistsInBusiness guards feedback writes against cross-tenant
// customer ids (crm_feedback has no FK).
func (s *Service) CustomerExistsInBusiness(ctx context.Context, businessID, customerID int64) bool {
	return s.repo.CustomerExistsInBusiness(ctx, businessID, customerID)
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, repo repository) *Service {
	return &Service{
		repo: repo,
	}
}

func invalidateFeedbackScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(feedbackScope, businessID))
}

type feedbackListResult struct {
	Items []domain.Feedback
	Total int
}

func (s *Service) List(ctx context.Context, businessID int64, filters domain.ListFilters, page, pageSize int) ([]domain.Feedback, int, error) {
	key := cacheKeys.Scoped(feedbackScope, businessID, fmt.Sprintf("list|f=%+v|p=%d|ps=%d", filters, page, pageSize))
	res, err := cacher.Of[feedbackListResult]().GetOrComputeSingle(ctx, key+"|domain=v2",
		func(ctx context.Context) (feedbackListResult, error) {
			items, total, err := s.repo.List(ctx, businessID, filters, page, pageSize)
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
	invalidateFeedbackScope(ctx, input.BusinessID)
	return id, nil
}
