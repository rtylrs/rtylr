package api

import (
	"context"
	"net/http"

	financereportv1 "github.com/voxire/voxie/gen/go/finance/report/v1"

	"github.com/voxire/onion/internal/modules/invoices/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

type ReportAPI struct {
	Service *application.Service
}

func NewReportAPI(service *application.Service) *ReportAPI {
	return &ReportAPI{Service: service}
}

func (a *ReportAPI) GetPLReport(ctx context.Context, req *financereportv1.GetPLReportRequest) (*financereportv1.GetPLReportResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	period := req.GetPeriod()
	if period == "" {
		period = "month"
	}

	report, err := a.Service.GetPLReport(ctx, businessID, period, req.GetDate())
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to generate P&L report")
	}

	return &financereportv1.GetPLReportResponse{
		Period:            report.Period,
		Revenue:           report.Revenue,
		Cogs:              report.COGS,
		GrossProfit:       report.GrossProfit,
		OperatingExpenses: report.OperatingExpenses,
		NetProfit:         report.NetProfit,
		MarginPct:         report.MarginPct,
	}, nil
}

func (a *ReportAPI) GetTaxReport(_ context.Context, _ *financereportv1.GetTaxReportRequest) (*financereportv1.GetTaxReportResponse, error) {
	return nil, notImplemented("GetTaxReport")
}

func (a *ReportAPI) ListExpenseCategoryDetails(_ context.Context, _ *financereportv1.ListExpenseCategoryDetailsRequest) (*financereportv1.ListExpenseCategoryDetailsResponse, error) {
	return nil, notImplemented("ListExpenseCategoryDetails")
}

func (a *ReportAPI) CreateExpenseCategoryDetail(_ context.Context, _ *financereportv1.CreateExpenseCategoryDetailRequest) (*financereportv1.CreateExpenseCategoryDetailResponse, error) {
	return nil, notImplemented("CreateExpenseCategoryDetail")
}

func (a *ReportAPI) ListTaxConfig(_ context.Context, _ *financereportv1.ListTaxConfigRequest) (*financereportv1.ListTaxConfigResponse, error) {
	return nil, notImplemented("ListTaxConfig")
}

func (a *ReportAPI) CreateTaxConfig(_ context.Context, _ *financereportv1.CreateTaxConfigRequest) (*financereportv1.CreateTaxConfigResponse, error) {
	return nil, notImplemented("CreateTaxConfig")
}

func notImplemented(method string) error {
	return onekiterrors.New(http.StatusNotImplemented, method+" is not implemented on OneKit yet")
}
