package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/invoices/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func PLReportRoutes() []string {
	return []string{"/finance/pl"}
}

type PLReportHandler struct {
	FinanceService *application.Service
}

type PLReportRequest struct {
	Period string `json:"period"`
	Date   string `json:"date"`
}

type PLReportResponse struct {
	Period            string  `json:"period"`
	Revenue           float64 `json:"revenue"`
	COGS              float64 `json:"cogs"`
	GrossProfit       float64 `json:"gross_profit"`
	OperatingExpenses float64 `json:"operating_expenses"`
	NetProfit         float64 `json:"net_profit"`
	MarginPct         float64 `json:"margin_pct"`
}

func (h *PLReportHandler) Serve(ctx context.Context, req *PLReportRequest) (PLReportResponse, error) {
	device := authPkg.DeviceCtx(ctx)

	if device.User == nil {
		return PLReportResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireFinancePermission(ctx, "finance:reports:view"); err != nil {
		return PLReportResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return PLReportResponse{}, onekiterrors.PermissionDenied()
	}

	period := req.Period
	if period == "" {
		period = "month"
	}

	report, err := h.FinanceService.GetPLReport(ctx, businessID, period, req.Date)
	if err != nil {
		return PLReportResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to generate P&L report")
	}

	return PLReportResponse{
		Period:            report.Period,
		Revenue:           report.Revenue,
		COGS:              report.COGS,
		GrossProfit:       report.GrossProfit,
		OperatingExpenses: report.OperatingExpenses,
		NetProfit:         report.NetProfit,
		MarginPct:         report.MarginPct,
	}, nil
}
