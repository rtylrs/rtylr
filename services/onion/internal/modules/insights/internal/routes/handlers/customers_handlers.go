package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/insights/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// GET /insights/customers

func GetCustomerKPIsRoutes() []string { return []string{"/insights/customers"} }

type GetCustomerKPIsHandler struct{ Service *application.Service }

type GetCustomerKPIsRequest struct{}

type GetCustomerKPIsResponse = application.CustomerKPIPayload

func (h *GetCustomerKPIsHandler) Serve(ctx context.Context, _ *GetCustomerKPIsRequest) (GetCustomerKPIsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetCustomerKPIsResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetCustomerKPIsResponse{}, errorModels.PermissionDeniedServerError()
	}
	payload, err := h.Service.GetCustomerKPIs(ctx, businessID)
	if err != nil {
		return GetCustomerKPIsResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to load customer KPIs",
		}
	}
	return payload, nil
}
