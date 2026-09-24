package api

import (
	"context"
	"net/http"

	flowinsights "github.com/voxire/onion/internal/modules/flow_insights"
	"github.com/voxire/onion/internal/modules/insights/internal/application"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	insightsdashboardv1 "github.com/voxire/voxie/gen/go/insights/dashboard/v1"
	insightssettingv1 "github.com/voxire/voxie/gen/go/insights/setting/v1"
)

type SettingAPI struct {
	Service *application.Service
}

func NewSettingAPI(service *application.Service) *SettingAPI {
	return &SettingAPI{Service: service}
}

func settingsToProto(payload application.SettingsPayload) *insightssettingv1.InsightsSettings {
	return &insightssettingv1.InsightsSettings{
		DefaultPeriod:    payload.DefaultPeriod,
		Currency:         payload.DefaultCurrency,
		ComparePrevious:  payload.ComparePrevious,
		DashboardWidgets: append([]string(nil), payload.DashboardWidgets...),
		FavoriteReports:  append([]string(nil), payload.FavoriteReports...),
	}
}

func (a *SettingAPI) GetInsightsSettings(ctx context.Context, _ *insightssettingv1.GetInsightsSettingsRequest) (*insightssettingv1.GetInsightsSettingsResponse, error) {
	businessID, err := resolveBusinessIDFor(ctx, "insights:dashboard:view")
	if err != nil {
		return nil, err
	}

	payload, err := a.Service.GetSettings(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to load insights settings")
	}

	return &insightssettingv1.GetInsightsSettingsResponse{Settings: settingsToProto(payload)}, nil
}

func (a *SettingAPI) UpdateInsightsSettings(ctx context.Context, req *insightssettingv1.UpdateInsightsSettingsRequest) (*insightssettingv1.UpdateInsightsSettingsResponse, error) {
	businessID, err := resolveBusinessIDFor(ctx, "insights:dashboard:view")
	if err != nil {
		return nil, err
	}

	body := req.GetSettings()
	payload, err := a.Service.UpdateSettings(ctx, businessID, application.SettingsPayload{
		DefaultPeriod:    body.GetDefaultPeriod(),
		DefaultCurrency:  body.GetCurrency(),
		ComparePrevious:  body.GetComparePrevious(),
		DashboardWidgets: append([]string(nil), body.GetDashboardWidgets()...),
		FavoriteReports:  append([]string(nil), body.GetFavoriteReports()...),
		FiscalYearStart:  "01-01",
		Timezone:         "UTC",
	})
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to save insights settings")
	}

	return &insightssettingv1.UpdateInsightsSettingsResponse{Settings: settingsToProto(payload)}, nil
}

// FlowReader is the slice of Flow's insights module that the dashboard needs.
// Narrow on purpose: insights should not learn the flow_* schema, and Flow keeps
// ownership of what its own numbers mean.
type FlowReader interface {
	FlowInsights(ctx context.Context, businessID int64) (flowinsights.FlowDelivery, error)
	FlowTeamWorkload(ctx context.Context, businessID int64) ([]flowinsights.FlowWorkloadMember, error)
	FlowRoadmap(ctx context.Context, businessID int64) ([]flowinsights.FlowRoadmapProject, error)
}

type DashboardAPI struct {
	Service *application.Service
	Flow    FlowReader
}

func NewDashboardAPI(service *application.Service, flow FlowReader) *DashboardAPI {
	return &DashboardAPI{Service: service, Flow: flow}
}

func (a *DashboardAPI) GetDashboardSummary(ctx context.Context, _ *insightsdashboardv1.GetDashboardSummaryRequest) (*insightsdashboardv1.GetDashboardSummaryResponse, error) {
	businessID, err := resolveBusinessIDFor(ctx, "insights:dashboard:view")
	if err != nil {
		return nil, err
	}

	snap, err := a.Service.GetDashboard(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to load insights dashboard")
	}

	return &insightsdashboardv1.GetDashboardSummaryResponse{
		Revenue: &insightsdashboardv1.InsightsMetricCard{
			Value:    snap.Revenue.Value,
			DeltaPct: snap.Revenue.DeltaPct,
		},
		Customers: &insightsdashboardv1.InsightsMetricCard{
			Value:    snap.Customers.Value,
			DeltaPct: snap.Customers.DeltaPct,
		},
		Orders: &insightsdashboardv1.InsightsMetricCard{
			Value:    snap.Orders.Value,
			DeltaPct: snap.Orders.DeltaPct,
		},
		Margin: &insightsdashboardv1.InsightsPctCard{
			Value:    snap.Margin.Value,
			DeltaPct: snap.Margin.DeltaPct,
		},
	}, nil
}

// GetInsights reports Flow's delivery numbers as this contract shapes them.
//
// Distinct from flow.insights.v1's getInsights, which returns velocity,
// burndown and throughput series. This one is the owner-dashboard projection:
// four task counts, a cycle time and a week-over-week velocity delta.
func (a *DashboardAPI) GetInsights(ctx context.Context, _ *insightsdashboardv1.GetInsightsRequest) (*insightsdashboardv1.GetInsightsResponse, error) {
	businessID, err := resolveBusinessIDFor(ctx, "insights:dashboard:view")
	if err != nil {
		return nil, err
	}

	stats, err := a.Flow.FlowInsights(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to load Flow insights")
	}

	return toInsightsResponse(stats), nil
}

// GetRoadmap reports Flow projects with progress, as this contract shapes them.
//
// flow.insights.v1's RoadmapProject carries colour, description, workspace and
// string ids for Flow's own board. This one carries int64 ids, a lifecycle
// status and explicit task counts. Milestones are not in this shape at all, so
// they are summed into the counts rather than dropped silently.
func (a *DashboardAPI) GetRoadmap(ctx context.Context, _ *insightsdashboardv1.GetRoadmapRequest) (*insightsdashboardv1.GetRoadmapResponse, error) {
	businessID, err := resolveBusinessIDFor(ctx, "insights:dashboard:view")
	if err != nil {
		return nil, err
	}

	projectsIn, err := a.Flow.FlowRoadmap(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to load Flow roadmap")
	}

	return &insightsdashboardv1.GetRoadmapResponse{Projects: toRoadmapProjects(projectsIn)}, nil
}

// GetWorkload reports per-person Flow load.
//
// open_tasks and total_active have no column behind them; they are derived from
// the three counts Flow does report, and the arithmetic is stated here so a
// reader does not have to infer it:
//
//	open_tasks   = total - completed - in_progress  (assigned, not started)
//	total_active = total - completed               (anything still outstanding)
//
// Neither can go negative given Flow's own counts, but both are clamped because
// this feeds a dashboard and a negative task count would be read as a bug in the
// display rather than in the data.
//
// WorkloadEntry has no name field, so this returns user ids only. A caller that
// wants to show people rather than numbers needs a separate identity lookup.
func (a *DashboardAPI) GetWorkload(ctx context.Context, _ *insightsdashboardv1.GetWorkloadRequest) (*insightsdashboardv1.GetWorkloadResponse, error) {
	businessID, err := resolveBusinessIDFor(ctx, "insights:dashboard:view")
	if err != nil {
		return nil, err
	}

	members, err := a.Flow.FlowTeamWorkload(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to load Flow workload")
	}

	return &insightsdashboardv1.GetWorkloadResponse{Workload: toWorkloadEntries(members)}, nil
}

func toInsightsResponse(stats flowinsights.FlowDelivery) *insightsdashboardv1.GetInsightsResponse {
	return &insightsdashboardv1.GetInsightsResponse{
		TotalTasks:       stats.TotalTasks,
		CompletedTasks:   stats.CompletedTasks,
		InProgressTasks:  stats.InProgressTasks,
		OverdueTasks:     stats.OverdueTasks,
		AvgCycleTimeDays: stats.AvgCompletionDays,
		Velocity:         stats.VelocityThisWeek,
		VelocityChange:   stats.VelocityThisWeek - stats.VelocityLastWeek,
	}
}

func toRoadmapProjects(in []flowinsights.FlowRoadmapProject) []*insightsdashboardv1.RoadmapProject {
	out := make([]*insightsdashboardv1.RoadmapProject, 0, len(in))
	for _, p := range in {
		out = append(out, &insightsdashboardv1.RoadmapProject{
			ProjectId:       p.ID,
			ProjectName:     p.Name,
			Status:          p.Status,
			TotalTasks:      p.TotalTasks,
			CompletedTasks:  p.CompletedTasks,
			ProgressPercent: p.Progress,
			StartDate:       p.StartDate,
			DueDate:         p.DueDate,
		})
	}

	return out
}

func toWorkloadEntries(in []flowinsights.FlowWorkloadMember) []*insightsdashboardv1.WorkloadEntry {
	out := make([]*insightsdashboardv1.WorkloadEntry, 0, len(in))
	for _, m := range in {
		out = append(out, &insightsdashboardv1.WorkloadEntry{
			UserId:          m.UserID,
			OpenTasks:       nonNegative(m.TotalTasks - m.CompletedTasks - m.InProgressTasks),
			InProgressTasks: m.InProgressTasks,
			CompletedTasks:  m.CompletedTasks,
			OverdueTasks:    m.OverdueTasks,
			TotalActive:     nonNegative(m.TotalTasks - m.CompletedTasks),
		})
	}

	return out
}

func nonNegative(v int64) int64 {
	if v < 0 {
		return 0
	}

	return v
}
