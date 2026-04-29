package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/flow_insights/internal/domain"
	"github.com/voxire/onion/internal/modules/flow_insights/internal/infrastructure"
)

type repository interface {
	GetInsightsStats(ctx context.Context, ownerID int64) (domain.InsightsStats, error)
	GetTeamWorkload(ctx context.Context, ownerID int64) (domain.TeamWorkloadResult, error)
	GetRoadmap(ctx context.Context, ownerID int64) (domain.RoadmapResult, error)
}

type Service struct {
	repo repository
}

func NewService(ctx context.Context) *Service {
	return &Service{
		repo: infrastructure.NewRepository(ctx),
	}
}

func (s *Service) GetInsights(ctx context.Context, ownerID int64) (domain.InsightsStats, error) {
	return s.repo.GetInsightsStats(ctx, ownerID)
}

func (s *Service) GetTeamWorkload(ctx context.Context, ownerID int64) (domain.TeamWorkloadResult, error) {
	return s.repo.GetTeamWorkload(ctx, ownerID)
}

func (s *Service) GetRoadmap(ctx context.Context, ownerID int64) (domain.RoadmapResult, error) {
	return s.repo.GetRoadmap(ctx, ownerID)
}
