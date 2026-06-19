package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/insights/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// GET /insights/operations

func GetOperationsKPIsRoutes() []string { return []string{"/insights/operations"} }

type GetOperationsKPIsHandler struct{ Service *application.Service }

type GetOperationsKPIsRequest struct{}

type GetOperationsKPIsResponse = application.OperationsKPIPayload

func (h *GetOperationsKPIsHandler) Serve(ctx context.Context, _ *GetOperationsKPIsRequest) (GetOperationsKPIsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetOperationsKPIsResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetOperationsKPIsResponse{}, errorModels.PermissionDeniedServerError()
	}
	payload, err := h.Service.GetOperationsKPIs(ctx, businessID)
	if err != nil {
		return GetOperationsKPIsResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to load operations KPIs",
		}
	}
	return payload, nil
}
