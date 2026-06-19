package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/kpis/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
	hubdashboardv1 "github.com/voxire/voxie/gen/go/hub/dashboard/v1"
)

func KPIRoutes() []string { return []string{"/hub/kpis"} }

type (
	KPIHandler struct{ Service *application.Service }
	KPIRequest struct{}
)

type KPIResponse = hubdashboardv1.GetKPIsResponse

func (h *KPIHandler) Serve(ctx context.Context, _ *KPIRequest) (KPIResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return KPIResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return KPIResponse{}, errorModels.PermissionDeniedServerError()
	}

	kpi, err := h.Service.GetKPIs(ctx, businessID)
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
