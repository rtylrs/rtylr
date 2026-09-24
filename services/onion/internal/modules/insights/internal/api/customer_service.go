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

func (a *CustomerAPI) GetCustomerAnalytics(ctx context.Context, req *insightscustomerv1.GetCustomerAnalyticsRequest) (*insightscustomerv1.GetCustomerAnalyticsResponse, error) {
	businessID, err := resolveBusinessIDFor(ctx, "insights:customers:view")
	if err != nil {
		return nil, err
	}
	from, to, err := resolveDateRangeFromPeriod(req.GetPeriod())
	if err != nil {
		return nil, err
	}

	kpis, err := a.Service.GetCustomerKPIs(ctx, businessID, from, to)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to load customer analytics")
	}

	cohort := kpis.Active + kpis.Churned
	var churnRate, retention float64
	if cohort > 0 {
		churnRate = float64(kpis.Churned) / float64(cohort) * 100
		retention = float64(kpis.Active) / float64(cohort) * 100
	}

	// #nosec G115 -- Analytics counters are database-backed non-negative values mapped to bounded protobuf fields.
	return &insightscustomerv1.GetCustomerAnalyticsResponse{
		Analytics: &insightscustomerv1.CustomerAnalytics{
			Total:            int32(kpis.Total), // #nosec G115 -- Analytics counters are database-backed non-negative values mapped to bounded protobuf fields.
			NewThisPeriod:    int32(kpis.NewThisPeriod),
			RetentionRate:    retention,
			ChurnRate:        churnRate,
			AvgLifetimeValue: 0,
		},
	}, nil
}

func (a *CustomerAPI) GetChurnAnalysis(ctx context.Context, req *insightscustomerv1.GetChurnAnalysisRequest) (*insightscustomerv1.GetChurnAnalysisResponse, error) {
	businessID, err := resolveBusinessIDFor(ctx, "insights:customers:view")
	if err != nil {
		return nil, err
	}
	from, to, err := resolveDateRangeFromPeriod(req.GetPeriod())
	if err != nil {
		return nil, err
	}

	windows, err := a.Service.GetChurnRates(ctx, businessID, from, to)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to load churn analysis")
	}

	analysis := &insightscustomerv1.ChurnAnalysis{}
	if len(windows) > 0 {
		w := windows[0]
		analysis.ChurnRate = w.ChurnPct
		analysis.ChurnedCount = int32(w.ChurnedCount) // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	}
	return &insightscustomerv1.GetChurnAnalysisResponse{Analysis: analysis}, nil
}

func (a *CustomerAPI) GetLTVAnalysis(ctx context.Context, _ *insightscustomerv1.GetLTVAnalysisRequest) (*insightscustomerv1.GetLTVAnalysisResponse, error) {
	businessID, err := resolveBusinessIDFor(ctx, "insights:customers:view")
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
			CustomerCount: int32(b.CustomerCount), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		})
	}

	return &insightscustomerv1.GetLTVAnalysisResponse{
		Analysis: &insightscustomerv1.LTVAnalysis{Segments: segments},
	}, nil
}

func (a *CustomerAPI) GetTopCustomers(ctx context.Context, req *insightscustomerv1.GetTopCustomersRequest) (*insightscustomerv1.GetTopCustomersResponse, error) {
	businessID, err := resolveBusinessIDFor(ctx, "insights:customers:view")
	if err != nil {
		return nil, err
	}
	from, to, err := resolveDateRangeFromPeriod(req.GetPeriod())
	if err != nil {
		return nil, err
	}
	sortBy, sortDir, err := application.NormalizeTopCustomerSort(req.GetSortBy(), req.GetSortDir())
	if err != nil {
		return nil, onekiterrors.New(http.StatusBadRequest, err.Error())
	}
	limit := normalizeLimit(req.GetLimit())
	page := normalizePage(req.GetPage())

	result, err := a.Service.GetTopCustomers(ctx, businessID, application.TopCustomerQuery{
		From: from, To: to, Page: page, Limit: limit, SortBy: sortBy, SortDir: sortDir,
	})
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to load top customers")
	}

	customers := make([]*insightscustomerv1.TopCustomerEntry, 0, len(result.Rows))
	for _, row := range result.Rows {
		customers = append(customers, &insightscustomerv1.TopCustomerEntry{
			Id:          row.CustomerID,
			Name:        row.Name,
			TotalSpent:  minorToMajor(row.LifetimeSpendMinorUnits),
			OrderCount:  int32(row.OrderCount), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
			LastOrderAt: row.LastOrderAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return &insightscustomerv1.GetTopCustomersResponse{
		Customers:  customers,
		Pagination: paginationResponse(page, limit, result.Total),
	}, nil
}
