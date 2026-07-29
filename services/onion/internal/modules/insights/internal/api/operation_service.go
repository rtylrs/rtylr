package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/insights/internal/application"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	insightsoperationv1 "github.com/voxire/voxie/gen/go/insights/operation/v1"
)

type OperationAPI struct {
	Service *application.Service
}

func NewOperationAPI(service *application.Service) *OperationAPI {
	return &OperationAPI{Service: service}
}

func (a *OperationAPI) GetHeadcountMetrics(ctx context.Context, _ *insightsoperationv1.GetHeadcountMetricsRequest) (*insightsoperationv1.GetHeadcountMetricsResponse, error) {
	businessID, err := resolveBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	row, err := a.Service.GetHeadcount(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to load headcount metrics")
	}

	return &insightsoperationv1.GetHeadcountMetricsResponse{
		Metrics: &insightsoperationv1.HeadcountMetrics{
			Current:            int32(row.ActiveStaff),
			HiredThisPeriod:    int32(maxInt64(row.MomDelta, 0)),
			DepartedThisPeriod: int32(maxInt64(-row.MomDelta, 0)),
			TurnoverRate:       row.MomDeltaPct,
		},
	}, nil
}

func (a *OperationAPI) GetShiftMetrics(ctx context.Context, req *insightsoperationv1.GetShiftMetricsRequest) (*insightsoperationv1.GetShiftMetricsResponse, error) {
	businessID, err := resolveBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	from, to, err := resolveDateRangeFromPeriod(req.GetPeriod())
	if err != nil {
		return nil, err
	}

	row, err := a.Service.GetShiftMetrics(ctx, businessID, from, to)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to load shift metrics")
	}

	return &insightsoperationv1.GetShiftMetricsResponse{
		Metrics: &insightsoperationv1.ShiftMetrics{
			TotalShifts:  int32(row.Scheduled),
			Covered:      int32(row.Completed),
			Missed:       int32(row.NoShows),
			CoverageRate: row.CompletionRate,
		},
	}, nil
}

func (a *OperationAPI) GetTimeOffMetrics(ctx context.Context, _ *insightsoperationv1.GetTimeOffMetricsRequest) (*insightsoperationv1.GetTimeOffMetricsResponse, error) {
	businessID, err := resolveBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	row, err := a.Service.GetTimeOffSummary(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to load time-off metrics")
	}

	return &insightsoperationv1.GetTimeOffMetricsResponse{
		Metrics: &insightsoperationv1.TimeOffMetrics{
			TotalRequests: int32(row.Pending + row.Approved),
			Approved:      int32(row.Approved),
			Pending:       int32(row.Pending),
		},
	}, nil
}

func (a *OperationAPI) GetUtilization(ctx context.Context, _ *insightsoperationv1.GetUtilizationRequest) (*insightsoperationv1.GetUtilizationResponse, error) {
	businessID, err := resolveBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	row, err := a.Service.GetUtilization(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to load utilization metrics")
	}

	return &insightsoperationv1.GetUtilizationResponse{
		Metrics: &insightsoperationv1.UtilizationMetrics{
			AvgUtilization: row.UtilizationPct,
			TotalHours:     row.ScheduledHours,
			AvailableHours: row.AvailableHours,
		},
	}, nil
}

func (a *OperationAPI) GetOperationsSummary(ctx context.Context, _ *insightsoperationv1.GetOperationsSummaryRequest) (*insightsoperationv1.GetOperationsSummaryResponse, error) {
	businessID, err := resolveBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	kpis, err := a.Service.GetOperationsKPIs(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to load operations summary")
	}

	return &insightsoperationv1.GetOperationsSummaryResponse{
		ActiveStaff:       kpis.ActiveStaff,
		ShiftsThisWeek:    kpis.ShiftsThisWeek,
		PendingTimeOff:    kpis.PendingTimeOff,
		OpenLeaveRequests: kpis.OpenLeaveRequests,
	}, nil
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
