package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/insights/internal/infrastructure"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) GetInsights(ctx context.Context, ownerID int64, workspaceID *int64) (infrastructure.InsightsData, error) {
	return s.repo.GetInsights(ctx, ownerID, workspaceID)
}

func (s *Service) GetTeamWorkload(ctx context.Context, ownerID int64, workspaceID *int64) ([]infrastructure.WorkloadEntry, error) {
	return s.repo.GetTeamWorkload(ctx, ownerID, workspaceID)
}

func (s *Service) GetRoadmap(ctx context.Context, ownerID int64, workspaceID *int64) ([]infrastructure.RoadmapProject, error) {
	return s.repo.GetRoadmap(ctx, ownerID, workspaceID)
}
