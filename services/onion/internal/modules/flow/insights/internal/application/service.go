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
	GetLoggedTime(ctx context.Context, businessID int64, from, to time.Time) (domain.LoggedTime, error)
	ListDashboardProjects(ctx context.Context, businessID, userID int64) ([]domain.DashboardProject, error)
	GetDashboardProject(ctx context.Context, businessID, userID, projectID int64) (domain.DashboardProject, bool, error)
	GetDashboardProjectByID(ctx context.Context, businessID, projectID int64) (domain.DashboardProject, bool, error)
	ListDashboardProjectsByIDs(ctx context.Context, businessID int64, projectIDs []int64) ([]domain.DashboardProject, error)
	GetProjectTaskProgress(ctx context.Context, businessID, projectID int64) (domain.ProjectTaskProgress, error)
	GetProjectTimeTotals(ctx context.Context, businessID, projectID int64, thisFrom, thisTo, priorFrom, priorTo time.Time) (domain.ProjectTimeTotals, error)
}

type Service struct {
	repo   repository
	access ProjectAccess
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

// GetLoggedTime is not cached here: its window ends at the caller's "now", so
// no two calls share a key. The caller caches the summary it builds.
func (s *Service) GetLoggedTime(ctx context.Context, businessID int64, from, to time.Time) (domain.LoggedTime, error) {
	return s.repo.GetLoggedTime(ctx, businessID, from, to)
}

// The project dashboard reads below are not cached here. Visibility depends on
// the caller, and a project's tasks and time change while someone is looking
// at it, so each read goes to the database; all four are indexed by project.


// GetProjectTaskProgress returns one project's task completion and overdue work.
func (s *Service) GetProjectTaskProgress(ctx context.Context, businessID, projectID int64) (domain.ProjectTaskProgress, error) {
	return s.repo.GetProjectTaskProgress(ctx, businessID, projectID)
}

// GetProjectTimeTotals returns the minutes logged on one project.
func (s *Service) GetProjectTimeTotals(ctx context.Context, businessID, projectID int64, thisFrom, thisTo, priorFrom, priorTo time.Time) (domain.ProjectTimeTotals, error) {
	return s.repo.GetProjectTimeTotals(ctx, businessID, projectID, thisFrom, thisTo, priorFrom, priorTo)
}
