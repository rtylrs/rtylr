package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/insights/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// GET /insights/revenue/by-app
//
// Splits MTD revenue across the POS, ERP, CRM and Finance surfaces and
// computes each surface's percentage of the total.

func GetRevenueByAppRoutes() []string { return []string{"/insights/revenue/by-app"} }

type GetRevenueByAppHandler struct{ Service *application.Service }

type GetRevenueByAppRequest struct{}

type GetRevenueByAppResponse struct {
	ByApp []application.AppRevenueShare `json:"by_app"`
}

func (h *GetRevenueByAppHandler) Serve(ctx context.Context, _ *GetRevenueByAppRequest) (GetRevenueByAppResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetRevenueByAppResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetRevenueByAppResponse{}, errorModels.PermissionDeniedServerError()
	}
	shares, err := h.Service.GetRevenueByApp(ctx, businessID)
	if err != nil {
		return GetRevenueByAppResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to load revenue by app",
		}
	}
	return GetRevenueByAppResponse{ByApp: shares}, nil
}
