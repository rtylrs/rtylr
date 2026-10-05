package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/flow/insights/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	flowInsightsScope = "flow_insights"
	flowInsightsTTL   = 2 * time.Minute
)

type repository interface {
	GetInsightsStats(ctx context.Context, businessID int64) (domain.InsightsStats, error)
	GetTeamWorkload(ctx context.Context, businessID int64) (domain.TeamWorkloadResult, error)
	GetRoadmap(ctx context.Context, businessID int64) (domain.RoadmapResult, error)
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, repo repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) GetInsights(ctx context.Context, businessID int64) (domain.InsightsStats, error) {
	return cacher.Of[domain.InsightsStats]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(flowInsightsScope, businessID, "stats")+"|domain=v2",
		func(ctx context.Context) (domain.InsightsStats, error) {
			return s.repo.GetInsightsStats(ctx, businessID)
		},
		cacher.WithExpiry(ctx, flowInsightsTTL),
	)
}

func (s *Service) GetTeamWorkload(ctx context.Context, businessID int64) (domain.TeamWorkloadResult, error) {
	return cacher.Of[domain.TeamWorkloadResult]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(flowInsightsScope, businessID, "workload")+"|domain=v2",
		func(ctx context.Context) (domain.TeamWorkloadResult, error) {
			return s.repo.GetTeamWorkload(ctx, businessID)
		},
		cacher.WithExpiry(ctx, flowInsightsTTL),
	)
}

func (s *Service) GetRoadmap(ctx context.Context, businessID int64) (domain.RoadmapResult, error) {
	return cacher.Of[domain.RoadmapResult]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(flowInsightsScope, businessID, "roadmap")+"|domain=v2",
		func(ctx context.Context) (domain.RoadmapResult, error) {
			return s.repo.GetRoadmap(ctx, businessID)
		},
		cacher.WithExpiry(ctx, flowInsightsTTL),
	)
}
