package application

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/flow_insights/internal/domain"
	"github.com/voxire/onion/internal/modules/flow_insights/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	flowInsightsScope = "flow_insights"
	flowInsightsTTL   = 2 * time.Minute
)

type repository interface {
	GetInsightsStats(ctx context.Context, ownerID int64) (domain.InsightsStats, error)
	GetTeamWorkload(ctx context.Context, ownerID int64) (domain.TeamWorkloadResult, error)
	GetRoadmap(ctx context.Context, ownerID int64) (domain.RoadmapResult, error)
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{
		repo: infrastructure.NewRepository(db),
	}
}

func (s *Service) GetInsights(ctx context.Context, ownerID int64) (domain.InsightsStats, error) {
	return cacher.Of[domain.InsightsStats]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(flowInsightsScope, ownerID, "stats"),
		func(ctx context.Context) (domain.InsightsStats, error) {
			return s.repo.GetInsightsStats(ctx, ownerID)
		},
		cacher.WithExpiry(ctx, flowInsightsTTL),
	)
}

func (s *Service) GetTeamWorkload(ctx context.Context, ownerID int64) (domain.TeamWorkloadResult, error) {
	return cacher.Of[domain.TeamWorkloadResult]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(flowInsightsScope, ownerID, "workload"),
		func(ctx context.Context) (domain.TeamWorkloadResult, error) {
			return s.repo.GetTeamWorkload(ctx, ownerID)
		},
		cacher.WithExpiry(ctx, flowInsightsTTL),
	)
}

func (s *Service) GetRoadmap(ctx context.Context, ownerID int64) (domain.RoadmapResult, error) {
	return cacher.Of[domain.RoadmapResult]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(flowInsightsScope, ownerID, "roadmap"),
		func(ctx context.Context) (domain.RoadmapResult, error) {
			return s.repo.GetRoadmap(ctx, ownerID)
		},
		cacher.WithExpiry(ctx, flowInsightsTTL),
	)
}
