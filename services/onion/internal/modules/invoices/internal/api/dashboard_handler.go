package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/invoices/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func DashboardRoutes() []string {
	return []string{"/finance/dashboard"}
}

type DashboardHandler struct {
	FinanceService *application.Service
}

type DashboardRequest struct{}

type DashboardResponse struct {
	TotalReceivable  float64 `json:"total_receivable"`
	OverdueAmount    float64 `json:"overdue_amount"`
	TotalExpensesMTD float64 `json:"total_expenses_mtd"`
	TotalRevenueMTD  float64 `json:"total_revenue_mtd"`
	NetProfitMTD     float64 `json:"net_profit_mtd"`
}

func (h *DashboardHandler) Serve(ctx context.Context, _ *DashboardRequest) (DashboardResponse, error) {
	device := authPkg.DeviceCtx(ctx)

	if device.User == nil {
		return DashboardResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireFinancePermission(ctx, "finance:reports:view"); err != nil {
		return DashboardResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return DashboardResponse{}, onekiterrors.PermissionDenied()
	}

	dashboard, err := h.FinanceService.GetDashboard(ctx, businessID)
	if err != nil {
		return DashboardResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get finance dashboard")
	}

	return DashboardResponse{
		TotalReceivable:  dashboard.TotalReceivable,
		OverdueAmount:    dashboard.OverdueAmount,
		TotalExpensesMTD: dashboard.TotalExpensesMTD,
		TotalRevenueMTD:  dashboard.TotalRevenueMTD,
		NetProfitMTD:     dashboard.NetProfitMTD,
	}, nil
}
