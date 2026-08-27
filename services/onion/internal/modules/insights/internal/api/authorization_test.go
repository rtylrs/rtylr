package api

import (
	"context"
	"errors"
	"net/http"
	"testing"

	onekithttp "github.com/corezio/onekit/http"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	insightscustomerv1 "github.com/voxire/voxie/gen/go/insights/customer/v1"
	insightsdashboardv1 "github.com/voxire/voxie/gen/go/insights/dashboard/v1"
	insightsfinancialv1 "github.com/voxire/voxie/gen/go/insights/financial/v1"
	insightsoperationv1 "github.com/voxire/voxie/gen/go/insights/operation/v1"
	insightsproductv1 "github.com/voxire/voxie/gen/go/insights/product/v1"
	insightsreportv1 "github.com/voxire/voxie/gen/go/insights/report/v1"
	insightsrevenuev1 "github.com/voxire/voxie/gen/go/insights/revenue/v1"
	insightssettingv1 "github.com/voxire/voxie/gen/go/insights/setting/v1"
)

type insightsAuthorizationCall func(context.Context) error

func insightsUser(permissions ...string) *authPkg.AuthUser {
	return &authPkg.AuthUser{
		BusinessID: 9,
		UserID:     7,
		EffectiveAccess: &authorization.EffectiveAccess{
			Apps: []authorization.EffectiveAppAccess{{
				AppID:       authorization.AppInsights,
				Licensed:    true,
				HasAccess:   true,
				Permissions: permissions,
			}},
		},
	}
}

func insightsStatus(err error) int {
	var oneKitErr *onekithttp.Error
	if errors.As(err, &oneKitErr) {
		return int(oneKitErr.GetStatusCode())
	}
	return 0
}

func assertInsightsDeniedBeforeService(t *testing.T, permission string, call insightsAuthorizationCall) {
	t.Helper()
	ctx := authPkg.WithAuthUser(context.Background(), insightsUser())
	if err := call(ctx); insightsStatus(err) != http.StatusForbidden {
		t.Fatalf("permission %s returned %v, want a 403 before service access", permission, err)
	}
}

func TestInsightsAPIsRequireExactPermissionsBeforeServiceAccess(t *testing.T) {
	cases := []struct {
		name       string
		permission string
		call       insightsAuthorizationCall
	}{
		{"dashboard summary", "insights:dashboard:view", func(ctx context.Context) error {
			_, err := (&DashboardAPI{}).GetDashboardSummary(ctx, &insightsdashboardv1.GetDashboardSummaryRequest{})
			return err
		}},
		{"dashboard flow insights", "insights:dashboard:view", func(ctx context.Context) error {
			_, err := (&DashboardAPI{}).GetInsights(ctx, &insightsdashboardv1.GetInsightsRequest{})
			return err
		}},
		{"dashboard roadmap", "insights:dashboard:view", func(ctx context.Context) error {
			_, err := (&DashboardAPI{}).GetRoadmap(ctx, &insightsdashboardv1.GetRoadmapRequest{})
			return err
		}},
		{"dashboard workload", "insights:dashboard:view", func(ctx context.Context) error {
			_, err := (&DashboardAPI{}).GetWorkload(ctx, &insightsdashboardv1.GetWorkloadRequest{})
			return err
		}},
		{"revenue overview", "insights:dashboard:view", func(ctx context.Context) error {
			_, err := (&RevenueAPI{}).GetRevenueOverview(ctx, &insightsrevenuev1.GetRevenueOverviewRequest{})
			return err
		}},
		{"revenue by app", "insights:dashboard:view", func(ctx context.Context) error {
			_, err := (&RevenueAPI{}).GetRevenueByApp(ctx, &insightsrevenuev1.GetRevenueByAppRequest{})
			return err
		}},
		{"revenue comparison", "insights:dashboard:view", func(ctx context.Context) error {
			_, err := (&RevenueAPI{}).GetRevenueComparison(ctx, &insightsrevenuev1.GetRevenueComparisonRequest{})
			return err
		}},
		{"headcount", "insights:dashboard:view", func(ctx context.Context) error {
			_, err := (&OperationAPI{}).GetHeadcountMetrics(ctx, &insightsoperationv1.GetHeadcountMetricsRequest{})
			return err
		}},
		{"operations summary", "insights:dashboard:view", func(ctx context.Context) error {
			_, err := (&OperationAPI{}).GetOperationsSummary(ctx, &insightsoperationv1.GetOperationsSummaryRequest{})
			return err
		}},
		{"shift metrics", "insights:dashboard:view", func(ctx context.Context) error {
			_, err := (&OperationAPI{}).GetShiftMetrics(ctx, &insightsoperationv1.GetShiftMetricsRequest{})
			return err
		}},
		{"time off metrics", "insights:dashboard:view", func(ctx context.Context) error {
			_, err := (&OperationAPI{}).GetTimeOffMetrics(ctx, &insightsoperationv1.GetTimeOffMetricsRequest{})
			return err
		}},
		{"utilization", "insights:dashboard:view", func(ctx context.Context) error {
			_, err := (&OperationAPI{}).GetUtilization(ctx, &insightsoperationv1.GetUtilizationRequest{})
			return err
		}},
		{"best sellers", "insights:dashboard:view", func(ctx context.Context) error {
			_, err := (&ProductAPI{}).GetBestSellers(ctx, &insightsproductv1.GetBestSellersRequest{})
			return err
		}},
		{"category performance", "insights:dashboard:view", func(ctx context.Context) error {
			_, err := (&ProductAPI{}).GetCategoryPerformance(ctx, &insightsproductv1.GetCategoryPerformanceRequest{})
			return err
		}},
		{"product margins", "insights:dashboard:view", func(ctx context.Context) error {
			_, err := (&ProductAPI{}).GetProductMargins(ctx, &insightsproductv1.GetProductMarginsRequest{})
			return err
		}},
		{"slow movers", "insights:dashboard:view", func(ctx context.Context) error {
			_, err := (&ProductAPI{}).GetSlowMovers(ctx, &insightsproductv1.GetSlowMoversRequest{})
			return err
		}},
		{"product performance", "insights:dashboard:view", func(ctx context.Context) error {
			_, err := (&ProductAPI{}).GetProductPerformance(ctx, &insightsproductv1.GetProductPerformanceRequest{})
			return err
		}},
		{"cash flow", "insights:financial:view", func(ctx context.Context) error {
			_, err := (&FinancialAPI{}).GetCashFlowAnalytics(ctx, &insightsfinancialv1.GetCashFlowAnalyticsRequest{})
			return err
		}},
		{"expense analytics", "insights:financial:view", func(ctx context.Context) error {
			_, err := (&FinancialAPI{}).GetExpenseAnalytics(ctx, &insightsfinancialv1.GetExpenseAnalyticsRequest{})
			return err
		}},
		{"invoice aging", "insights:financial:view", func(ctx context.Context) error {
			_, err := (&FinancialAPI{}).GetInvoiceAging(ctx, &insightsfinancialv1.GetInvoiceAgingRequest{})
			return err
		}},
		{"customer analytics", "insights:customers:view", func(ctx context.Context) error {
			_, err := (&CustomerAPI{}).GetCustomerAnalytics(ctx, &insightscustomerv1.GetCustomerAnalyticsRequest{})
			return err
		}},
		{"customer churn", "insights:customers:view", func(ctx context.Context) error {
			_, err := (&CustomerAPI{}).GetChurnAnalysis(ctx, &insightscustomerv1.GetChurnAnalysisRequest{})
			return err
		}},
		{"customer ltv", "insights:customers:view", func(ctx context.Context) error {
			_, err := (&CustomerAPI{}).GetLTVAnalysis(ctx, &insightscustomerv1.GetLTVAnalysisRequest{})
			return err
		}},
		{"top customers", "insights:customers:view", func(ctx context.Context) error {
			_, err := (&CustomerAPI{}).GetTopCustomers(ctx, &insightscustomerv1.GetTopCustomersRequest{})
			return err
		}},
		{"report templates", "insights:reports:create", func(ctx context.Context) error {
			_, err := (&ReportAPI{}).ListReportTemplates(ctx, &insightsreportv1.ListReportTemplatesRequest{})
			return err
		}},
		{"scheduled report list", "insights:reports:create", func(ctx context.Context) error {
			_, err := (&ReportAPI{}).ListScheduledReports(ctx, &insightsreportv1.ListScheduledReportsRequest{})
			return err
		}},
		{"scheduled report create", "insights:reports:create", func(ctx context.Context) error {
			_, err := (&ReportAPI{}).CreateScheduledReport(ctx, &insightsreportv1.CreateScheduledReportRequest{})
			return err
		}},
		{"scheduled report update", "insights:reports:create", func(ctx context.Context) error {
			_, err := (&ReportAPI{}).UpdateScheduledReport(ctx, &insightsreportv1.UpdateScheduledReportRequest{})
			return err
		}},
		{"scheduled report delete", "insights:reports:create", func(ctx context.Context) error {
			_, err := (&ReportAPI{}).DeleteScheduledReport(ctx, &insightsreportv1.DeleteScheduledReportRequest{})
			return err
		}},
		{"revenue report generate", "insights:reports:create", func(ctx context.Context) error {
			_, err := (&ReportAPI{}).GenerateRevenueReport(ctx, &insightsreportv1.GenerateRevenueReportRequest{})
			return err
		}},
		{"revenue report export", "insights:reports:export", func(ctx context.Context) error {
			_, err := (&ReportAPI{}).ExportRevenueReport(ctx, &insightsreportv1.ExportRevenueReportRequest{})
			return err
		}},
		{"insights settings get", "insights:dashboard:view", func(ctx context.Context) error {
			_, err := (&SettingAPI{}).GetInsightsSettings(ctx, &insightssettingv1.GetInsightsSettingsRequest{})
			return err
		}},
		{"insights settings update", "insights:dashboard:view", func(ctx context.Context) error {
			_, err := (&SettingAPI{}).UpdateInsightsSettings(ctx, &insightssettingv1.UpdateInsightsSettingsRequest{})
			return err
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertInsightsDeniedBeforeService(t, tc.permission, tc.call)
		})
	}
}

func TestInsightsPermissionHelperRejectsWrongAppAndUnlicensedUsers(t *testing.T) {
	ctx := context.Background()
	wrongApp := &authPkg.AuthUser{BusinessID: 9, UserID: 7, EffectiveAccess: &authorization.EffectiveAccess{
		Apps: []authorization.EffectiveAppAccess{{AppID: authorization.AppFinance, Licensed: true, HasAccess: true, Permissions: []string{"insights:dashboard:view"}}},
	}}
	if _, err := resolveBusinessIDFor(authPkg.WithAuthUser(ctx, wrongApp), "insights:dashboard:view"); insightsStatus(err) != http.StatusForbidden {
		t.Fatalf("wrong app returned %v, want 403", err)
	}
	unlicensed := insightsUser("insights:dashboard:view")
	unlicensed.EffectiveAccess.Apps[0].Licensed = false
	if _, err := resolveBusinessIDFor(authPkg.WithAuthUser(ctx, unlicensed), "insights:dashboard:view"); insightsStatus(err) != http.StatusForbidden {
		t.Fatalf("unlicensed app returned %v, want 403", err)
	}
}
