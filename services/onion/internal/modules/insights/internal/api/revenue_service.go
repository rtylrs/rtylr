package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/insights/internal/application"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	insightsrevenuev1 "github.com/voxire/voxie/gen/go/insights/revenue/v1"
)

type RevenueAPI struct {
	Service *application.Service
}

func NewRevenueAPI(service *application.Service) *RevenueAPI {
	return &RevenueAPI{Service: service}
}

func (a *RevenueAPI) GetRevenueOverview(ctx context.Context, req *insightsrevenuev1.GetRevenueOverviewRequest) (*insightsrevenuev1.GetRevenueOverviewResponse, error) {
	businessID, err := resolveBusinessIDFor(ctx, "insights:dashboard:view")
	if err != nil {
		return nil, err
	}

	from, to, err := resolveDateRangeFromPeriod(req.GetPeriod())
	if err != nil {
		return nil, err
	}

	series, err := a.Service.GetRevenueSeries(ctx, businessID, from, to)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to load revenue overview")
	}

	trend := make([]*insightsrevenuev1.RevenuePoint, 0, len(series))
	var total int64
	for _, point := range series {
		total += point.AmountMinorUnits
		trend = append(trend, &insightsrevenuev1.RevenuePoint{
			Date:   point.Date,
			Amount: minorToMajor(point.AmountMinorUnits),
		})
	}

	return &insightsrevenuev1.GetRevenueOverviewResponse{
		Overview: &insightsrevenuev1.RevenueOverview{
			Total:  minorToMajor(total),
			Change: 0,
			Trend:  trend,
		},
	}, nil
}

func (a *RevenueAPI) GetRevenueByApp(ctx context.Context, _ *insightsrevenuev1.GetRevenueByAppRequest) (*insightsrevenuev1.GetRevenueByAppResponse, error) {
	businessID, err := resolveBusinessIDFor(ctx, "insights:dashboard:view")
	if err != nil {
		return nil, err
	}

	shares, err := a.Service.GetRevenueByApp(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to load revenue by app")
	}

	apps := make([]*insightsrevenuev1.RevenueByApp, 0, len(shares))
	for _, share := range shares {
		apps = append(apps, &insightsrevenuev1.RevenueByApp{
			App:        share.App,
			Amount:     minorToMajor(share.AmountMinorUnits),
			Percentage: share.Pct,
		})
	}

	return &insightsrevenuev1.GetRevenueByAppResponse{Apps: apps}, nil
}

func (a *RevenueAPI) GetRevenueComparison(ctx context.Context, req *insightsrevenuev1.GetRevenueComparisonRequest) (*insightsrevenuev1.GetRevenueComparisonResponse, error) {
	businessID, err := resolveBusinessIDFor(ctx, "insights:dashboard:view")
	if err != nil {
		return nil, err
	}

	from, to, err := resolveDateRangeFromPeriod(req.GetPeriod())
	if err != nil {
		return nil, err
	}

	payload, err := a.Service.GetRevenueComparison(ctx, businessID, from, to)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to load revenue comparison")
	}

	periodLabel := req.GetPeriod()
	if periodLabel == "" {
		periodLabel = "current"
	}

	return &insightsrevenuev1.GetRevenueComparisonResponse{
		Comparisons: []*insightsrevenuev1.RevenueComparison{{
			Period:   periodLabel,
			Current:  minorToMajor(payload.ThisPeriodMinorUnits),
			Previous: minorToMajor(payload.LastPeriodMinorUnits),
			Change:   payload.DeltaPct,
		}},
	}, nil
}
