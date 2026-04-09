package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/invoices/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func DashboardRoutes() []string {
	return []string{"/finance/dashboard"}
}

type DashboardHandler struct {
	FinanceService *application.Service
}

type DashboardRequest struct{}

type DashboardResponse struct {
	TotalReceivable  float64 `json:"totalReceivable"`
	OverdueAmount    float64 `json:"overdueAmount"`
	TotalExpensesMTD float64 `json:"totalExpensesMtd"`
	NetProfitMTD     float64 `json:"netProfitMtd"`
}

func (h *DashboardHandler) Serve(ctx context.Context, _ *DashboardRequest) (DashboardResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)

	dashboard, err := h.FinanceService.GetDashboard(ctx, device.User.UserID)
	if err != nil {
		return DashboardResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get finance dashboard",
		}
	}

	return DashboardResponse{
		TotalReceivable:  dashboard.TotalReceivable,
		OverdueAmount:    dashboard.OverdueAmount,
		TotalExpensesMTD: dashboard.TotalExpensesMTD,
		NetProfitMTD:     dashboard.NetProfitMTD,
	}, nil
}
