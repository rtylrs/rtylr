package insights

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"strconv"
	"strings"
	"time"

	insightsapp "github.com/voxire/onion/internal/modules/insights/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
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

func insightsToolPermission(name string) string {
	switch name {
	case scheduleInsightsReportToolName:
		return "insights:reports:create"
	case getInsightsCustomerKPIsToolName,
		listInsightsTopCustomersToolName,
		getInsightsCustomerChurnToolName,
		getInsightsLTVDistributionToolName:
		return "insights:customers:view"
	case getInsightsCashFlowToolName,
		getInsightsExpensesByCategoryToolName,
		getInsightsInvoiceAgingToolName:
		return "insights:financial:view"
	default:
		return "insights:dashboard:view"
	}
}

func parseScheduleReportToolInput(input map[string]any) (insightsapp.ScheduledReportInput, error) {
	if input == nil {
		input = map[string]any{}
	}

	templateID := normalizeScheduleReportTemplate(
		insightsToolFirstString(input, "template_id", "templateId", "report_id", "report", "template", "name"),
		input,
	)
	cadence := normalizeScheduleReportCadence(
		insightsToolFirstString(input, "cadence", "frequency", "interval", "period", "repeat"),
	)
	channel := normalizeScheduleReportChannel(
		insightsToolFirstString(input, "channel", "delivery_channel", "deliveryChannel", "via"),
	)
	recipient := strings.TrimSpace(insightsToolFirstString(input, "recipient", "to", "email", "slack_channel", "slackChannel", "channel_id", "channelId"))
	if recipient == "" {
		recipient = firstScheduleReportRecipient(input["recipients"])
	}
	if channel == "" {
		channel = inferScheduleReportChannel(recipient)
	}
	if cadence == "" {
		cadence = inferScheduleReportCadence(templateID)
	}

	if err := validateScheduleReportRecipient(channel, recipient); err != nil {
		return insightsapp.ScheduledReportInput{}, err
	}

	return insightsapp.ScheduledReportInput{
		TemplateID: templateID,
		Cadence:    cadence,
		Channel:    channel,
		Recipient:  recipient,
		IsActive:   scheduleReportOptionalBool(input["is_active"], input["isActive"]),
	}, nil
}

func normalizeScheduleReportTemplate(value string, input map[string]any) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	slug := scheduleReportSlug(value)
	switch slug {
	case "revenue", "sales", "revenue-report", "sales-report":
		cadence := normalizeScheduleReportCadence(
			insightsToolFirstString(input, "cadence", "frequency", "interval", "period", "repeat"),
		)
		if cadence == "daily" {
			return "revenue-daily"
		}
		return "revenue-monthly"
	case "daily-revenue", "daily-sales", "revenue-daily", "sales-daily":
		return "revenue-daily"
	case "monthly-revenue", "monthly-sales", "revenue-monthly", "sales-monthly":
		return "revenue-monthly"
	case "top-customer", "top-customers", "customers", "customer-leaderboard":
		return "top-customers"
	case "invoice-aging", "ar-aging", "accounts-receivable", "receivables":
		return "invoice-aging"
	case "expenses", "expense", "expenses-category", "expenses-by-category", "expense-category":
		return "expenses-category"
	case "best-seller", "best-sellers", "bestsellers", "best-selling-products":
		return "best-sellers"
	case "slow-mover", "slow-movers", "slow-moving-products":
		return "slow-movers"
	case "shift-coverage", "shifts", "shift-report":
		return "shift-coverage"
	case "headcount", "staff-count", "employee-count":
		return "headcount"
	default:
		return slug
	}
}

func normalizeScheduleReportCadence(value string) string {
	slug := scheduleReportSlug(value)
	switch slug {
	case "day", "daily", "every-day", "each-day":
		return "daily"
	case "week", "weekly", "every-week", "each-week":
		return "weekly"
	case "month", "monthly", "every-month", "each-month":
		return "monthly"
	default:
		return slug
	}
}

func normalizeScheduleReportChannel(value string) string {
	slug := scheduleReportSlug(value)
	switch slug {
	case "mail", "e-mail", "email":
		return "email"
	case "slack", "slack-channel":
		return "slack"
	default:
		return slug
	}
}

func inferScheduleReportCadence(templateID string) string {
	switch templateID {
	case "revenue-daily":
		return "daily"
	case "revenue-monthly":
		return "monthly"
	default:
		return ""
	}
}

func inferScheduleReportChannel(recipient string) string {
	recipient = strings.TrimSpace(recipient)
	if recipient == "" {
		return ""
	}
	if strings.HasPrefix(recipient, "#") {
		return "slack"
	}
	return "email"
}

func firstScheduleReportRecipient(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(v)
	case []string:
		if len(v) == 0 {
			return ""
		}
		return strings.TrimSpace(v[0])
	case []any:
		for _, item := range v {
			if recipient := strings.TrimSpace(insightsToolString(item)); recipient != "" {
				return recipient
			}
		}
	}
	return ""
}

func validateScheduleReportRecipient(channel, recipient string) error {
	if strings.TrimSpace(recipient) == "" {
		return errors.New("recipient is required")
	}
	switch strings.ToLower(strings.TrimSpace(recipient)) {
	case "me", "my email", "current user", "current_user":
		return errors.New("recipient must be an explicit email address or Slack channel")
	}
	switch channel {
	case "email":
		if _, err := mail.ParseAddress(recipient); err != nil {
			return errors.New("recipient must be a valid email address")
		}
	case "slack":
		if !strings.HasPrefix(recipient, "#") && !strings.HasPrefix(strings.ToUpper(recipient), "C") {
			return errors.New("recipient must be a Slack channel name or channel id")
		}
	}
	return nil
}

func scheduleReportOptionalBool(primary, fallback any) *bool {
	for _, value := range []any{primary, fallback} {
		switch v := value.(type) {
		case nil:
			continue
		case bool:
			result := v
			return &result
		case string:
			trimmed := strings.TrimSpace(strings.ToLower(v))
			if trimmed == "" {
				continue
			}
			result := trimmed != "false" && trimmed != "0" && trimmed != "no" && trimmed != "inactive"
			return &result
		}
	}
	return nil
}

func scheduleReportSlug(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.NewReplacer(
		"_", "-",
		" ", "-",
		".", "-",
		"/", "-",
		":", "-",
	).Replace(value)
	for strings.Contains(value, "--") {
		value = strings.ReplaceAll(value, "--", "-")
	}
	return strings.Trim(value, "-")
}

func insightsToolDateRange(input map[string]any) (time.Time, time.Time, error) {
	from := insightsToolFirstString(input, "from", "start", "start_date", "date_from")
	to := insightsToolFirstString(input, "to", "end", "end_date", "date_to")
	return insightsapp.ResolveDateRange(from, to)
}

func insightsToolRangePayload(from, to time.Time) map[string]string {
	return map[string]string{
		"from": from.Format("2006-01-02"),
		"to":   to.Format("2006-01-02"),
	}
}

func insightsToolFirstString(input map[string]any, keys ...string) string {
	for _, key := range keys {
		value := strings.TrimSpace(insightsToolString(input[key]))
		if value != "" {
			return value
		}
	}
	return ""
}

func insightsToolString(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case json.Number:
		return v.String()
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprint(v)
	}
}

func insightsToolInt32(input map[string]any, fallback int32, keys ...string) (int32, error) {
	for _, key := range keys {
		value, ok := input[key]
		if !ok || value == nil {
			continue
		}
		n, err := insightsToolNumericInt32(value)
		if err != nil {
			return 0, fmt.Errorf("invalid %s: %w", key, err)
		}
		if n <= 0 {
			return fallback, nil
		}
		return n, nil
	}
	return fallback, nil
}

func insightsToolNumericInt32(value any) (int32, error) {
	switch v := value.(type) {
	case int:
		return int32(v), nil // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	case int8:
		return int32(v), nil
	case int16:
		return int32(v), nil
	case int32:
		return v, nil
	case int64:
		return int32(v), nil // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	case uint:
		return int32(v), nil // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	case uint8:
		return int32(v), nil
	case uint16:
		return int32(v), nil
	case uint32:
		return int32(v), nil // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	case uint64:
		return int32(v), nil // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	case float32:
		return int32(v), nil
	case float64:
		return int32(v), nil
	case string:
		n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 32)
		if err != nil {
			return 0, err
		}
		return int32(n), nil
	case json.Number:
		n, err := strconv.ParseInt(v.String(), 10, 32)
		if err != nil {
			return 0, err
		}
		return int32(n), nil
	default:
		return 0, fmt.Errorf("expected integer, got %T", value)
	}
}
