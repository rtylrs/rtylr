package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/finance/invoices/internal/application"
	"github.com/voxire/onion/internal/modules/finance/invoices/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	financereportv1 "github.com/voxire/voxie/gen/go/finance/report/v1"
)

type ReportAPI struct {
	Service *application.Service
}

func NewReportAPI(service *application.Service) *ReportAPI {
	return &ReportAPI{Service: service}
}

func (a *ReportAPI) GetPLReport(ctx context.Context, req *financereportv1.GetPLReportRequest) (*financereportv1.GetPLReportResponse, error) {
	if err := requireFinancePermission(ctx, "finance:reports:view"); err != nil {
		return nil, err
	}
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

	byCategory := make([]*financereportv1.PLCategoryTotal, 0, len(report.ExpensesByCategory))
	for _, total := range report.ExpensesByCategory {
		byCategory = append(byCategory, &financereportv1.PLCategoryTotal{
			CategoryId: total.CategoryID,
			Name:       total.Name,
			Color:      total.Color,
			Amount:     total.Amount,
		})
	}
	return &financereportv1.GetPLReportResponse{
		Period:              report.Period,
		Revenue:             report.Revenue,
		Cogs:                report.COGS,
		GrossProfit:         report.GrossProfit,
		OperatingExpenses:   report.OperatingExpenses,
		NetProfit:           report.NetProfit,
		MarginPct:           report.MarginPct,
		ExpensesByCategory:  byCategory,
		PendingExpenses:     report.PendingExpenses,
		PendingExpenseCount: report.PendingExpenseCount,
	}, nil
}

func (a *ReportAPI) GetTaxReport(ctx context.Context, req *financereportv1.GetTaxReportRequest) (*financereportv1.GetTaxReportResponse, error) {
	if err := requireFinancePermission(ctx, "finance:reports:view"); err != nil {
		return nil, err
	}
	businessID, errModel := invoiceBusinessID(ctx)
	if errModel != nil {
		return nil, errModel
	}

	period := req.GetPeriod()
	if period == "" {
		period = "month"
	}

	report, err := a.Service.GetTaxReport(ctx, businessID, period, req.GetDate())
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to generate tax report")
	}

	return &financereportv1.GetTaxReportResponse{
		Period:       report.Period,
		TaxCollected: report.TaxCollected,
		TaxOwed:      report.TaxOwed,
	}, nil
}

func (a *ReportAPI) ListExpenseCategoryDetails(ctx context.Context, _ *financereportv1.ListExpenseCategoryDetailsRequest) (*financereportv1.ListExpenseCategoryDetailsResponse, error) {
	if err := requireFinancePermission(ctx, "finance:expenses:view"); err != nil {
		return nil, err
	}
	businessID, errModel := invoiceBusinessID(ctx)
	if errModel != nil {
		return nil, errModel
	}

	categories, err := a.Service.ListExpenseCategories(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list expense categories")
	}

	payloads := make([]*financereportv1.ExpenseCategoryDetail, 0, len(categories))
	for _, c := range categories {
		payloads = append(payloads, newExpenseCategoryDetailPayload(c))
	}

	return &financereportv1.ListExpenseCategoryDetailsResponse{Categories: payloads}, nil
}

func (a *ReportAPI) CreateExpenseCategoryDetail(ctx context.Context, req *financereportv1.CreateExpenseCategoryDetailRequest) (*financereportv1.CreateExpenseCategoryDetailResponse, error) {
	if err := requireFinancePermission(ctx, "finance:expenses:create"); err != nil {
		return nil, err
	}
	businessID, errModel := invoiceBusinessID(ctx)
	if errModel != nil {
		return nil, errModel
	}

	if req.GetName() == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "Category name is required")
	}

	category, err := a.Service.CreateExpenseCategory(ctx, businessID, domain.ExpenseCategoryInput{
		Name:  req.GetName(),
		Color: req.GetColor(),
		Icon:  req.GetIcon(),
	})
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create expense category")
	}

	return &financereportv1.CreateExpenseCategoryDetailResponse{Category: newExpenseCategoryDetailPayload(*category)}, nil
}

func (a *ReportAPI) ListTaxConfig(ctx context.Context, _ *financereportv1.ListTaxConfigRequest) (*financereportv1.ListTaxConfigResponse, error) {
	if err := requireFinancePermission(ctx, "finance:tax:configure"); err != nil {
		return nil, err
	}
	businessID, errModel := invoiceBusinessID(ctx)
	if errModel != nil {
		return nil, errModel
	}

	configs, err := a.Service.ListTaxConfigs(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list tax configs")
	}

	payloads := make([]*financereportv1.TaxConfig, 0, len(configs))
	for _, c := range configs {
		payloads = append(payloads, newTaxConfigDetailPayload(c))
	}

	return &financereportv1.ListTaxConfigResponse{Configs: payloads}, nil
}

func (a *ReportAPI) CreateTaxConfig(ctx context.Context, req *financereportv1.CreateTaxConfigRequest) (*financereportv1.CreateTaxConfigResponse, error) {
	if err := requireFinancePermission(ctx, "finance:tax:configure"); err != nil {
		return nil, err
	}
	businessID, errModel := invoiceBusinessID(ctx)
	if errModel != nil {
		return nil, errModel
	}

	if req.GetName() == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "Tax config name is required")
	}
	if req.GetRate() < 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Tax rate must be non-negative")
	}

	config, err := a.Service.CreateTaxConfig(ctx, businessID, domain.TaxConfigInput{
		Name:      req.GetName(),
		Rate:      req.GetRate(),
		Type:      req.GetType(),
		IsDefault: req.GetIsDefault(),
	})
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create tax config")
	}

	return &financereportv1.CreateTaxConfigResponse{Config: newTaxConfigDetailPayload(*config)}, nil
}

func newExpenseCategoryDetailPayload(c domain.ExpenseCategory) *financereportv1.ExpenseCategoryDetail {
	return &financereportv1.ExpenseCategoryDetail{
		Id:        c.ID,
		OwnerId:   c.OwnerID,
		Name:      c.Name,
		Color:     c.Color,
		Icon:      c.Icon,
		CreatedAt: c.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func newTaxConfigDetailPayload(t domain.TaxConfig) *financereportv1.TaxConfig {
	return &financereportv1.TaxConfig{
		Id:        t.ID,
		OwnerId:   t.OwnerID,
		Name:      t.Name,
		Rate:      t.Rate,
		Type:      t.Type,
		IsDefault: t.IsDefault,
	}
}
