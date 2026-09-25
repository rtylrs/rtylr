package insights

import (
	"context"
)

const (
	getInsightsDashboardToolName          = "insights.get_dashboard"
	getInsightsRevenueSeriesToolName      = "insights.get_revenue_series"
	getInsightsRevenueComparisonToolName  = "insights.get_revenue_comparison"
	getInsightsRevenueByAppToolName       = "insights.get_revenue_by_app"
	getInsightsCustomerKPIsToolName       = "insights.get_customer_kpis"
	listInsightsTopCustomersToolName      = "insights.list_top_customers"
	getInsightsCustomerChurnToolName      = "insights.get_customer_churn"
	getInsightsLTVDistributionToolName    = "insights.get_ltv_distribution"
	getInsightsProductKPIsToolName        = "insights.get_product_kpis"
	listInsightsBestSellersToolName       = "insights.list_bestsellers"
	listInsightsSlowMoversToolName        = "insights.list_slow_movers"
	getInsightsSalesByCategoryToolName    = "insights.get_sales_by_category"
	getInsightsMarginDistributionToolName = "insights.get_margin_distribution"
	getInsightsCashFlowToolName           = "insights.get_cash_flow"
	getInsightsExpensesByCategoryToolName = "insights.get_expenses_by_category"
	getInsightsInvoiceAgingToolName       = "insights.get_invoice_aging"
	getInsightsOperationsKPIsToolName     = "insights.get_operations_kpis"
	getInsightsHeadcountToolName          = "insights.get_headcount"
	getInsightsShiftMetricsToolName       = "insights.get_shift_metrics"
	getInsightsTimeOffToolName            = "insights.get_time_off"
	getInsightsUtilizationToolName        = "insights.get_utilization"
	getInsightsSettingsToolName           = "insights.get_settings"
	scheduleInsightsReportToolName        = "insights.schedule_report"
)

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	switch name {
	case getInsightsDashboardToolName,
		getInsightsRevenueSeriesToolName,
		getInsightsRevenueComparisonToolName,
		getInsightsRevenueByAppToolName,
		getInsightsCustomerKPIsToolName,
		listInsightsTopCustomersToolName,
		getInsightsCustomerChurnToolName,
		getInsightsLTVDistributionToolName,
		getInsightsProductKPIsToolName,
		listInsightsBestSellersToolName,
		listInsightsSlowMoversToolName,
		getInsightsSalesByCategoryToolName,
		getInsightsMarginDistributionToolName,
		getInsightsCashFlowToolName,
		getInsightsExpensesByCategoryToolName,
		getInsightsInvoiceAgingToolName,
		getInsightsOperationsKPIsToolName,
		getInsightsHeadcountToolName,
		getInsightsShiftMetricsToolName,
		getInsightsTimeOffToolName,
		getInsightsUtilizationToolName,
		getInsightsSettingsToolName,
		scheduleInsightsReportToolName:
	default:
		return nil, false, nil
	}

	output, err := m.executeInsightsTool(ctx, name, input)
	return output, true, err
}
