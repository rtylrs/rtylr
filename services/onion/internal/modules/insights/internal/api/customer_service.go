package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/insights/internal/application"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	insightscustomerv1 "github.com/voxire/voxie/gen/go/insights/customer/v1"
)

type CustomerAPI struct {
	Service *application.Service
}

func NewCustomerAPI(service *application.Service) *CustomerAPI {
	return &CustomerAPI{Service: service}
}

func (a *CustomerAPI) GetCustomerAnalytics(ctx context.Context, _ *insightscustomerv1.GetCustomerAnalyticsRequest) (*insightscustomerv1.GetCustomerAnalyticsResponse, error) {
	businessID, err := resolveBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	kpis, err := a.Service.GetCustomerKPIs(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to load customer analytics")
	}

	var churnRate float64
	if kpis.Total > 0 {
		churnRate = (float64(kpis.Churned) / float64(kpis.Total)) * 100
	}
	var retention float64
	if kpis.Total > 0 {
		retention = (float64(kpis.Active) / float64(kpis.Total)) * 100
	}

	return &insightscustomerv1.GetCustomerAnalyticsResponse{
		Analytics: &insightscustomerv1.CustomerAnalytics{
			Total:            int32(kpis.Total),
			NewThisPeriod:    int32(kpis.NewThisMonth),
			RetentionRate:    retention,
			ChurnRate:        churnRate,
			AvgLifetimeValue: 0,
		},
	}, nil
}

func (a *CustomerAPI) GetChurnAnalysis(ctx context.Context, _ *insightscustomerv1.GetChurnAnalysisRequest) (*insightscustomerv1.GetChurnAnalysisResponse, error) {
	businessID, err := resolveBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	windows, err := a.Service.GetChurnRates(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to load churn analysis")
	}

	analysis := &insightscustomerv1.ChurnAnalysis{}
	if len(windows) > 0 {
		w := windows[0]
		analysis.ChurnRate = w.ChurnPct
		analysis.ChurnedCount = int32(w.ChurnedCount)
	}

	return &insightscustomerv1.GetChurnAnalysisResponse{Analysis: analysis}, nil
}

func (a *CustomerAPI) GetLTVAnalysis(ctx context.Context, _ *insightscustomerv1.GetLTVAnalysisRequest) (*insightscustomerv1.GetLTVAnalysisResponse, error) {
	businessID, err := resolveBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	buckets, err := a.Service.GetLTVDistribution(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to load LTV analysis")
	}

	segments := make([]*insightscustomerv1.LTVSegment, 0, len(buckets))
	for _, b := range buckets {
		mid := minorToMajor(b.LowerMinorUnits + b.UpperMinorUnits/2)
		segments = append(segments, &insightscustomerv1.LTVSegment{
			Segment:       b.Label,
			AvgLtv:        mid,
			CustomerCount: int32(b.CustomerCount),
		})
	}

	return &insightscustomerv1.GetLTVAnalysisResponse{
		Analysis: &insightscustomerv1.LTVAnalysis{
			Segments: segments,
		},
	}, nil
}

func (a *CustomerAPI) GetTopCustomers(ctx context.Context, req *insightscustomerv1.GetTopCustomersRequest) (*insightscustomerv1.GetTopCustomersResponse, error) {
	businessID, err := resolveBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := a.Service.GetTopCustomers(ctx, businessID, normalizeLimit(req.GetLimit()))
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to load top customers")
	}

	customers := make([]*insightscustomerv1.TopCustomerEntry, 0, len(rows))
	for _, row := range rows {
		customers = append(customers, &insightscustomerv1.TopCustomerEntry{
			Id:         row.CustomerID,
			Name:       row.Name,
			TotalSpent: minorToMajor(row.LifetimeSpendMinorUnits),
			OrderCount: int32(row.OrderCount),
		})
	}

	return &insightscustomerv1.GetTopCustomersResponse{Customers: customers}, nil
}
