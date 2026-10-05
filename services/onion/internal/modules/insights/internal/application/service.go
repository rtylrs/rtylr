package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/insights/internal/domain"
)

// Service coordinates read models and the small set of Insights-owned settings.
type Service struct {
	repo      repository
	dashboard DashboardReaders
}

type FinanceDashboardReader interface {
	InsightsPaidInvoiceTotal(context.Context, int64, time.Time, time.Time) (int64, error)
	InsightsExpenseTotal(context.Context, int64, time.Time, time.Time) (int64, error)
}

type POSDashboardReader interface {
	CountOrdersCreatedBetween(context.Context, int64, time.Time, time.Time) (int64, error)
}

type CRMDashboardReader interface {
	CountCustomersCreatedBetween(context.Context, int64, time.Time, time.Time) (int64, error)
}

type DashboardReaders struct {
	Finance FinanceDashboardReader
	POS     POSDashboardReader
	CRM     CRMDashboardReader
}

type repository interface {
	GetInsightsGeneration(context.Context, int64) (time.Time, error)
	GetProductMarginRows(context.Context, int64) ([]domain.ProductMarginRow, error)
	GetTopCustomers(context.Context, int64, domain.TopCustomerQuery) (domain.TopCustomerPage, error)
	GetCustomerKPIs(context.Context, int64, time.Time, time.Time) (domain.CustomerKPIs, error)
	GetChurnRates(context.Context, int64, time.Time, time.Time) ([]domain.ChurnRateRow, error)
	GetLTVDistribution(context.Context, int64) ([]domain.LTVBucketRow, error)
	GetCashFlowSeries(context.Context, int64, time.Time, time.Time) ([]domain.CashFlowDayRow, error)
	GetExpensesByCategory(context.Context, int64, time.Time, time.Time) ([]domain.ExpenseByCategoryRow, error)
	GetInvoiceAging(context.Context, int64) ([]domain.InvoiceAgingBucketRow, error)
	GetOperationsKPIs(context.Context, int64) (domain.OperationsKPIs, error)
	GetHeadcount(context.Context, int64, time.Time, time.Time) (domain.HeadcountSnapshot, error)
	GetShiftMetrics(context.Context, int64, time.Time, time.Time) (domain.ShiftMetricsRow, error)
	GetTimeOffSummary(context.Context, int64, time.Time, time.Time) (domain.TimeOffSummaryRow, error)
	GetUtilization(context.Context, int64, time.Time, time.Time) (domain.UtilizationRow, error)
	GetProductKPIs(context.Context, int64) (domain.ProductKPIs, error)
	GetBestSellers(context.Context, int64, int32) ([]domain.BestSellerRow, error)
	GetSalesByCategory(context.Context, int64) ([]domain.CategorySalesRow, error)
	GetMarginDistribution(context.Context, int64) ([]domain.MarginBucketRow, error)
	GetSlowMovers(context.Context, int64, int32) ([]domain.SlowMoverRow, error)
	GetRevenueSeries(context.Context, int64, time.Time, time.Time) ([]domain.RevenuePoint, error)
	GetRevenueByApp(context.Context, int64) ([]domain.RevenueByAppRow, error)
	ListScheduledReports(context.Context, int64, string, string, int32, int32) ([]domain.ScheduledReportRow, int64, error)
	CreateScheduledReport(context.Context, domain.ScheduledReportWrite) (int64, error)
	UpdateScheduledReport(context.Context, domain.ScheduledReportWrite) error
	DeleteScheduledReport(context.Context, int64, int64) error
	GetScheduledReport(context.Context, int64, int64) (domain.ScheduledReportRow, error)
	GetSettings(context.Context, int64) (domain.SettingsRow, error)
	UpsertSettings(context.Context, domain.SettingsRow) error
}

func NewService(repo repository, dashboard DashboardReaders) *Service {
	return &Service{repo: repo, dashboard: dashboard}
}

// GetProductMarginRows returns per-product margin rows for margin reporting.
func (s *Service) GetProductMarginRows(ctx context.Context, businessID int64) ([]domain.ProductMarginRow, error) {
	return s.repo.GetProductMarginRows(ctx, businessID)
}
