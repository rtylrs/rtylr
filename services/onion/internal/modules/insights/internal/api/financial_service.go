package api

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/insights/internal/application"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	insightsfinancialv1 "github.com/voxire/voxie/gen/go/insights/financial/v1"
)

type FinancialAPI struct {
	Service *application.Service
}

func NewFinancialAPI(service *application.Service) *FinancialAPI {
	return &FinancialAPI{Service: service}
}

func resolveFinancialPeriod(period string) (time.Time, time.Time, error) {
	now := time.Now().UTC()
	switch strings.ToLower(strings.TrimSpace(period)) {
	case "month", "this_month":
		from := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		return from, from.AddDate(0, 1, -1), nil
	case "quarter", "this_quarter":
		firstMonth := time.Month(((int(now.Month())-1)/3)*3 + 1)
		from := time.Date(now.Year(), firstMonth, 1, 0, 0, 0, 0, time.UTC)
		return from, from.AddDate(0, 3, -1), nil
	case "year", "this_year":
		from := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
		return from, time.Date(now.Year(), 12, 31, 0, 0, 0, 0, time.UTC), nil
	default:
		return resolveDateRangeFromPeriod(period)
	}
}

func (a *FinancialAPI) GetCashFlowAnalytics(ctx context.Context, req *insightsfinancialv1.GetCashFlowAnalyticsRequest) (*insightsfinancialv1.GetCashFlowAnalyticsResponse, error) {
	businessID, err := resolveBusinessIDFor(ctx, "insights:financial:view")
	if err != nil {
		return nil, err
	}

	from, to, err := resolveFinancialPeriod(req.GetPeriod())
	if err != nil {
		return nil, err
	}

	points, err := a.Service.GetCashFlowSeries(ctx, businessID, from, to)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to load cash flow analytics")
	}

	trend := make([]*insightsfinancialv1.CashFlowPoint, 0, len(points))
	var inflows, outflows int64
	for _, p := range points {
		inflows += p.InflowMinorUnits
		outflows += p.OutflowMinorUnits
		trend = append(trend, &insightsfinancialv1.CashFlowPoint{
			Date:    p.Date,
			Inflow:  minorToMajor(p.InflowMinorUnits),
			Outflow: minorToMajor(p.OutflowMinorUnits),
			Net:     minorToMajor(p.InflowMinorUnits - p.OutflowMinorUnits),
		})
	}

	return &insightsfinancialv1.GetCashFlowAnalyticsResponse{
		Analytics: &insightsfinancialv1.CashFlowAnalytics{
			Inflows:  minorToMajor(inflows),
			Outflows: minorToMajor(outflows),
			Net:      minorToMajor(inflows - outflows),
			Trend:    trend,
		},
	}, nil
}

func (a *FinancialAPI) GetExpenseAnalytics(ctx context.Context, req *insightsfinancialv1.GetExpenseAnalyticsRequest) (*insightsfinancialv1.GetExpenseAnalyticsResponse, error) {
	businessID, err := resolveBusinessIDFor(ctx, "insights:financial:view")
	if err != nil {
		return nil, err
	}

	from, to, err := resolveDateRangeFromPeriod(req.GetPeriod())
	if err != nil {
		return nil, err
	}

	rows, err := a.Service.GetExpensesByCategory(ctx, businessID, from, to)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to load expense analytics")
	}

	byCategory := make([]*insightsfinancialv1.ExpenseByCategory, 0, len(rows))
	var total int64
	for _, row := range rows {
		total += row.AmountMinorUnits
	}
	for _, row := range rows {
		var pct float64
		if total > 0 {
			pct = (float64(row.AmountMinorUnits) / float64(total)) * 100
		}
		byCategory = append(byCategory, &insightsfinancialv1.ExpenseByCategory{
			Category:   row.CategoryName,
			Amount:     minorToMajor(row.AmountMinorUnits),
			Percentage: pct,
		})
	}

	return &insightsfinancialv1.GetExpenseAnalyticsResponse{
		Analytics: &insightsfinancialv1.ExpenseAnalytics{
			Total:      minorToMajor(total),
			ByCategory: byCategory,
		},
	}, nil
}

func (a *FinancialAPI) GetInvoiceAging(ctx context.Context, _ *insightsfinancialv1.GetInvoiceAgingRequest) (*insightsfinancialv1.GetInvoiceAgingResponse, error) {
	businessID, err := resolveBusinessIDFor(ctx, "insights:financial:view")
	if err != nil {
		return nil, err
	}

	rows, err := a.Service.GetInvoiceAging(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to load invoice aging")
	}

	buckets := make([]*insightsfinancialv1.InvoiceAgingBucket, 0, len(rows))
	var totalOutstanding int64
	for _, row := range rows {
		totalOutstanding += row.AmountMinorUnits
		buckets = append(buckets, &insightsfinancialv1.InvoiceAgingBucket{
			Bucket: row.Bucket,
			Count:  int32(row.InvoiceCount), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
			Amount: minorToMajor(row.AmountMinorUnits),
		})
	}

	return &insightsfinancialv1.GetInvoiceAgingResponse{
		Aging: &insightsfinancialv1.InvoiceAging{
			TotalOutstanding: minorToMajor(totalOutstanding),
			Buckets:          buckets,
		},
	}, nil
}

// CentsToFloat converts integer cents to standard float amount.
func CentsToFloat(cents int64) float64 {
	return float64(cents) / 100.0
}
