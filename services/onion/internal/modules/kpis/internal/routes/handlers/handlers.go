package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/kpis/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func KPIRoutes() []string { return []string{"/hub/kpis"} }

type KPIHandler struct{ Service *application.Service }
type KPIRequest struct{}

type KPIResponse struct {
	Revenue       float64 `json:"revenue"`
	OrderCount    int64   `json:"orderCount"`
	StaffCount    int64   `json:"staffCount"`
	CustomerCount int64   `json:"customerCount"`
	LowStockCount int64   `json:"lowStockCount"`
	ActiveApps    int64   `json:"activeApps"`
}

func (h *KPIHandler) Serve(ctx context.Context, _ *KPIRequest) (KPIResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return KPIResponse{}, errorModels.PermissionDeniedServerError()
	}

	kpi, err := h.Service.GetKPIs(ctx, device.User.UserID)
	if err != nil {
		return KPIResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to get KPIs",
		}
	}

	return KPIResponse{
		Revenue:       kpi.Revenue,
		OrderCount:    kpi.OrderCount,
		StaffCount:    kpi.StaffCount,
		CustomerCount: kpi.CustomerCount,
		LowStockCount: kpi.LowStockCount,
		ActiveApps:    kpi.ActiveApps,
	}, nil
}
