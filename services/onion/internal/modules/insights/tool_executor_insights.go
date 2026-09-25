package insights

import (
	"context"
	"errors"
	"fmt"

	insightsapp "github.com/voxire/onion/internal/modules/insights/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func (m *Module) executeInsightsTool(ctx context.Context, name string, input map[string]any) (map[string]any, error) {
	if m == nil || m.service == nil {
		return nil, errors.New("insights service is unavailable")
	}
	if input == nil {
		input = map[string]any{}
	}

	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, onekiterrors.PermissionDenied()
	}
	if !authPkg.HasAppPermission(authUser, authorization.AppInsights, insightsToolPermission(name)) {
		return nil, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	switch name {
	case getInsightsDashboardToolName:
		dashboard, err := m.service.GetDashboard(ctx, businessID)
		if err != nil {
			return nil, err
		}
		return map[string]any{"dashboard": dashboard}, nil

	case getInsightsRevenueSeriesToolName:
		from, to, err := insightsToolDateRange(input)
		if err != nil {
			return nil, err
		}
		points, err := m.service.GetRevenueSeries(ctx, businessID, from, to)
		if err != nil {
			return nil, err
		}
		var total int64
		for _, point := range points {
			total += point.AmountMinorUnits
		}
		return map[string]any{
			"range":             insightsToolRangePayload(from, to),
			"points":            points,
			"returned":          len(points),
			"total_minor_units": total,
		}, nil

	case getInsightsRevenueComparisonToolName:
		from, to, err := insightsToolDateRange(input)
		if err != nil {
			return nil, err
		}
		comparison, err := m.service.GetRevenueComparison(ctx, businessID, from, to)
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"range":      insightsToolRangePayload(from, to),
			"comparison": comparison,
		}, nil

	case getInsightsRevenueByAppToolName:
		shares, err := m.service.GetRevenueByApp(ctx, businessID)
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"apps":     shares,
			"returned": len(shares),
		}, nil

	case getInsightsCustomerKPIsToolName:
		from, to, err := insightsToolDateRange(input)
		if err != nil {
			return nil, err
		}
		kpis, err := m.service.GetCustomerKPIs(ctx, businessID, from, to)
		if err != nil {
			return nil, err
		}
		return map[string]any{"customer_kpis": kpis, "range": insightsToolRangePayload(from, to)}, nil

	case listInsightsTopCustomersToolName:
		limit, err := insightsToolInt32(input, 10, "limit", "page_size")
		if err != nil {
			return nil, err
		}
		page, err := insightsToolInt32(input, 1, "page")
		if err != nil {
			return nil, err
		}
		from, to, err := insightsToolDateRange(input)
		if err != nil {
			return nil, err
		}
		sortBy, sortDir, err := insightsapp.NormalizeTopCustomerSort(
			insightsToolFirstString(input, "sort_by", "sortBy"),
			insightsToolFirstString(input, "sort_dir", "sortDir"),
		)
		if err != nil {
			return nil, err
		}
		result, err := m.service.GetTopCustomers(ctx, businessID, insightsapp.TopCustomerQuery{
			From: from, To: to, Page: page, Limit: limit, SortBy: sortBy, SortDir: sortDir,
		})
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"customers": result.Rows,
			"limit":     limit,
			"page":      page,
			"total":     result.Total,
			"range":     insightsToolRangePayload(from, to),
			"returned":  len(result.Rows),
		}, nil

	case getInsightsCustomerChurnToolName:
		from, to, err := insightsToolDateRange(input)
		if err != nil {
			return nil, err
		}
		windows, err := m.service.GetChurnRates(ctx, businessID, from, to)
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"windows":  windows,
			"returned": len(windows),
		}, nil

	case getInsightsLTVDistributionToolName:
		buckets, err := m.service.GetLTVDistribution(ctx, businessID)
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"buckets":  buckets,
			"returned": len(buckets),
		}, nil

	case getInsightsProductKPIsToolName:
		kpis, err := m.service.GetProductKPIs(ctx, businessID)
		if err != nil {
			return nil, err
		}
		return map[string]any{"product_kpis": kpis}, nil

	case listInsightsBestSellersToolName:
		limit, err := insightsToolInt32(input, 10, "limit", "page_size")
		if err != nil {
			return nil, err
		}
		products, err := m.service.GetBestSellers(ctx, businessID, limit)
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"products": products,
			"limit":    limit,
			"returned": len(products),
		}, nil

	case listInsightsSlowMoversToolName:
		days, err := insightsToolInt32(input, 60, "days", "window_days")
		if err != nil {
			return nil, err
		}
		products, err := m.service.GetSlowMovers(ctx, businessID, days)
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"products": products,
			"days":     days,
			"returned": len(products),
		}, nil

	case getInsightsSalesByCategoryToolName:
		categories, err := m.service.GetSalesByCategory(ctx, businessID)
		if err != nil {
			return nil, err
		}
		var total int64
		for _, category := range categories {
			total += category.RevenueMinorUnits
		}
		return map[string]any{
			"categories":        categories,
			"returned":          len(categories),
			"total_minor_units": total,
		}, nil

	case getInsightsMarginDistributionToolName:
		buckets, err := m.service.GetMarginDistribution(ctx, businessID)
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"buckets":  buckets,
			"returned": len(buckets),
		}, nil

	case getInsightsCashFlowToolName:
		from, to, err := insightsToolDateRange(input)
		if err != nil {
			return nil, err
		}
		points, err := m.service.GetCashFlowSeries(ctx, businessID, from, to)
		if err != nil {
			return nil, err
		}
		var inflow, outflow int64
		for _, point := range points {
			inflow += point.InflowMinorUnits
			outflow += point.OutflowMinorUnits
		}
		return map[string]any{
			"range":               insightsToolRangePayload(from, to),
			"points":              points,
			"returned":            len(points),
			"inflow_minor_units":  inflow,
			"outflow_minor_units": outflow,
			"net_minor_units":     inflow - outflow,
		}, nil

	case getInsightsExpensesByCategoryToolName:
		from, to, err := insightsToolDateRange(input)
		if err != nil {
			return nil, err
		}
		categories, err := m.service.GetExpensesByCategory(ctx, businessID, from, to)
		if err != nil {
			return nil, err
		}
		var total int64
		for _, category := range categories {
			total += category.AmountMinorUnits
		}
		return map[string]any{
			"range":             insightsToolRangePayload(from, to),
			"categories":        categories,
			"returned":          len(categories),
			"total_minor_units": total,
		}, nil

	case getInsightsInvoiceAgingToolName:
		buckets, err := m.service.GetInvoiceAging(ctx, businessID)
		if err != nil {
			return nil, err
		}
		var total int64
		for _, bucket := range buckets {
			total += bucket.AmountMinorUnits
		}
		return map[string]any{
			"buckets":           buckets,
			"returned":          len(buckets),
			"total_minor_units": total,
		}, nil

	case getInsightsOperationsKPIsToolName:
		kpis, err := m.service.GetOperationsKPIs(ctx, businessID)
		if err != nil {
			return nil, err
		}
		return map[string]any{"operations_kpis": kpis}, nil

	case getInsightsHeadcountToolName:
		from, to, err := insightsToolDateRange(input)
		if err != nil {
			return nil, err
		}
		headcount, err := m.service.GetHeadcount(ctx, businessID, from, to)
		if err != nil {
			return nil, err
		}
		return map[string]any{"headcount": headcount, "range": insightsToolRangePayload(from, to)}, nil

	case getInsightsShiftMetricsToolName:
		from, to, err := insightsToolDateRange(input)
		if err != nil {
			return nil, err
		}
		metrics, err := m.service.GetShiftMetrics(ctx, businessID, from, to)
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"range":   insightsToolRangePayload(from, to),
			"metrics": metrics,
		}, nil

	case getInsightsTimeOffToolName:
		from, to, err := insightsToolDateRange(input)
		if err != nil {
			return nil, err
		}
		timeOff, err := m.service.GetTimeOffSummary(ctx, businessID, from, to)
		if err != nil {
			return nil, err
		}
		return map[string]any{"time_off": timeOff, "range": insightsToolRangePayload(from, to)}, nil

	case getInsightsUtilizationToolName:
		from, to, err := insightsToolDateRange(input)
		if err != nil {
			return nil, err
		}
		utilization, err := m.service.GetUtilization(ctx, businessID, from, to)
		if err != nil {
			return nil, err
		}
		return map[string]any{"utilization": utilization, "range": insightsToolRangePayload(from, to)}, nil

	case getInsightsSettingsToolName:
		settings, err := m.service.GetSettings(ctx, businessID)
		if err != nil {
			return nil, err
		}
		return map[string]any{"settings": settings}, nil

	case scheduleInsightsReportToolName:
		schedule, err := parseScheduleReportToolInput(input)
		if err != nil {
			return nil, err
		}
		report, err := m.service.CreateScheduledReport(ctx, businessID, schedule)
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"scheduled_report": report,
			"message":          "Scheduled report created.",
		}, nil
	}

	return nil, fmt.Errorf("unsupported insights tool %q", name)
}
