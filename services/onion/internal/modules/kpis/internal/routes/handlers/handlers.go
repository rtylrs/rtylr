package handlers

import (
	"context"
	"net/http"

	hubv1 "github.com/voxire/voxie/gen/go/hub/v1"

	"github.com/voxire/onion/internal/modules/kpis/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func KPIRoutes() []string { return []string{"/hub/kpis"} }

type KPIHandler struct{ Service *application.Service }
type KPIRequest struct{}

type KPIResponse struct {
	Revenue       float64 `json:"revenue"`
	OrderCount    int64   `json:"order_count"`
	StaffCount    int64   `json:"staff_count"`
	CustomerCount int64   `json:"customer_count"`
	LowStockCount int64   `json:"low_stock_count"`
	ActiveApps    int64   `json:"active_apps"`
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
