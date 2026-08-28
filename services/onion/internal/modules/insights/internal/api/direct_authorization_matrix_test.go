package api

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/insights/internal/application"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
	insightscustomer "github.com/voxire/voxie/gen/go/insights/customer/v1"
	insightsdashboard "github.com/voxire/voxie/gen/go/insights/dashboard/v1"
	insightsfinancial "github.com/voxire/voxie/gen/go/insights/financial/v1"
	insightsreport "github.com/voxire/voxie/gen/go/insights/report/v1"
)

func TestDirectAuthorizationMatrix(t *testing.T) {
	testutil.RunDirectAuthorizationCases(t, []testutil.DirectAuthorizationCase{
		{Name: "insights dashboard view", App: authorization.AppInsights, Permission: "insights:dashboard:view", Call: func(ctx context.Context) (any, error) {
			response, err := NewDashboardAPI(application.NewService(ctx, testutil.DirectAuthorizationDB(ctx)), nil).GetDashboardSummary(ctx, nil)
			return response, err
		}},
		{Name: "insights financial view", App: authorization.AppInsights, Permission: "insights:financial:view", Call: func(ctx context.Context) (any, error) {
			response, err := NewFinancialAPI(application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))).GetCashFlowAnalytics(ctx, nil)
			return response, err
		}},
		{Name: "insights customers view", App: authorization.AppInsights, Permission: "insights:customers:view", Call: func(ctx context.Context) (any, error) {
			response, err := NewCustomerAPI(application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))).GetCustomerAnalytics(ctx, nil)
			return response, err
		}},
		{Name: "insights reports create", App: authorization.AppInsights, Permission: "insights:reports:create", Call: func(ctx context.Context) (any, error) {
			response, err := NewReportAPI(application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))).CreateReportTemplate(ctx, &insightsreport.CreateReportTemplateRequest{Name: "Direct Authorization Report", Description: "Direct authorization report", Type: "revenue", Format: "json"})
			return response, err
		}},
		{Name: "insights reports share", App: authorization.AppInsights, Permission: "insights:reports:share", Call: func(ctx context.Context) (any, error) {
			response, err := NewReportAPI(application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))).UpdateScheduledReport(ctx, &insightsreport.UpdateScheduledReportRequest{Id: 7206, TemplateSlug: "revenue-monthly", Frequency: "weekly", Channel: "email", RecipientEmail: "owner@test.local", IsActive: true})
			return response, err
		}},
		{Name: "insights reports export", App: authorization.AppInsights, Permission: "insights:reports:export", Call: func(ctx context.Context) (any, error) {
			response, err := NewReportAPI(application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))).ExportRevenueReport(ctx, nil)
			return response, err
		}},
		{Name: "insights dashboard settings boundary", App: authorization.AppInsights, Permission: "insights:dashboard:view", Call: func(ctx context.Context) (any, error) {
			response, err := NewSettingAPI(application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))).GetInsightsSettings(ctx, nil)
			return response, err
		}},
		{Name: "insights financial endpoint boundary", App: authorization.AppInsights, Permission: "insights:financial:view", Call: func(ctx context.Context) (any, error) {
			response, err := NewFinancialAPI(application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))).GetExpenseAnalytics(ctx, &insightsfinancial.GetExpenseAnalyticsRequest{})
			return response, err
		}},
		{Name: "insights customers endpoint boundary", App: authorization.AppInsights, Permission: "insights:customers:view", Call: func(ctx context.Context) (any, error) {
			response, err := NewCustomerAPI(application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))).GetChurnAnalysis(ctx, &insightscustomer.GetChurnAnalysisRequest{})
			return response, err
		}},
		{Name: "insights report list boundary", App: authorization.AppInsights, Permission: "insights:reports:create", Call: func(ctx context.Context) (any, error) {
			response, err := NewReportAPI(application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))).ListReportTemplates(ctx, &insightsreport.ListReportTemplatesRequest{})
			return response, err
		}},
		{Name: "insights dashboard endpoint boundary", App: authorization.AppInsights, Permission: "insights:dashboard:view", Call: func(ctx context.Context) (any, error) {
			response, err := NewDashboardAPI(application.NewService(ctx, testutil.DirectAuthorizationDB(ctx)), nil).GetDashboardSummary(ctx, &insightsdashboard.GetDashboardSummaryRequest{})
			return response, err
		}},
	})
}
