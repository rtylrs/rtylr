package api

import (
	"context"
	"net/http"

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
		DefaultPeriod: "this_month",
		Currency:      payload.DefaultCurrency,
	}
}

func (a *SettingAPI) GetInsightsSettings(ctx context.Context, _ *insightssettingv1.GetInsightsSettingsRequest) (*insightssettingv1.GetInsightsSettingsResponse, error) {
	businessID, err := resolveBusinessID(ctx)
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
	businessID, err := resolveBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	body := req.GetSettings()
	payload, err := a.Service.UpdateSettings(ctx, businessID, application.SettingsPayload{
		DefaultCurrency: body.GetCurrency(),
		FiscalYearStart: "01-01",
		Timezone:        "UTC",
	})
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to save insights settings")
	}

	return &insightssettingv1.UpdateInsightsSettingsResponse{Settings: settingsToProto(payload)}, nil
}

type DashboardAPI struct {
	Service *application.Service
}

func NewDashboardAPI(service *application.Service) *DashboardAPI {
	return &DashboardAPI{Service: service}
}

func (a *DashboardAPI) GetDashboardSummary(ctx context.Context, _ *insightsdashboardv1.GetDashboardSummaryRequest) (*insightsdashboardv1.GetDashboardSummaryResponse, error) {
	businessID, err := resolveBusinessID(ctx)
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

// Flow dashboard RPCs are not backed by the insights module yet.
func (a *DashboardAPI) GetInsights(context.Context, *insightsdashboardv1.GetInsightsRequest) (*insightsdashboardv1.GetInsightsResponse, error) {
	return nil, notImplemented("GetInsights")
}

func (a *DashboardAPI) GetKPIs(context.Context, *insightsdashboardv1.GetKPIsRequest) (*insightsdashboardv1.GetKPIsResponse, error) {
	return nil, notImplemented("GetKPIs")
}

func (a *DashboardAPI) GetRoadmap(context.Context, *insightsdashboardv1.GetRoadmapRequest) (*insightsdashboardv1.GetRoadmapResponse, error) {
	return nil, notImplemented("GetRoadmap")
}

func (a *DashboardAPI) GetWorkload(context.Context, *insightsdashboardv1.GetWorkloadRequest) (*insightsdashboardv1.GetWorkloadResponse, error) {
	return nil, notImplemented("GetWorkload")
}
