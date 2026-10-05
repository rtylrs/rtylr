package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/pos/daysummary/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const dayLiveTTL = 1 * time.Minute

type repository interface {
	GetDaySummary(ctx context.Context, businessID int64, date string) (domain.DaySummary, error)
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, repo repository) *Service {
	return &Service{repo: repo}
}

// GetDaySummary derives from orders and sits under the order namespace so any
// order mutation wipes today's summary too. Past dates are effectively
// immutable but we still TTL them defensively (refunds, voids, late edits).
func (s *Service) GetDaySummary(ctx context.Context, businessID int64, date string) (domain.DaySummary, error) {
	key := fmt.Sprintf("%s|day_summary=%s", cacheKeys.OrdersList(businessID), date)
	return cacher.Of[domain.DaySummary]().GetOrComputeSingle(ctx, key+"|domain=v2",
		func(ctx context.Context) (domain.DaySummary, error) {
			return s.repo.GetDaySummary(ctx, businessID, date)
		},
		cacher.WithExpiry(ctx, dayLiveTTL),
	)
}
