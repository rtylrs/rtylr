package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/insights/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// GET /insights/operations/utilization

func GetUtilizationRoutes() []string { return []string{"/insights/operations/utilization"} }

type GetUtilizationHandler struct{ Service *application.Service }

type GetUtilizationRequest struct{}

type GetUtilizationResponse = application.UtilizationPayload

func (h *GetUtilizationHandler) Serve(ctx context.Context, _ *GetUtilizationRequest) (GetUtilizationResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetUtilizationResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetUtilizationResponse{}, errorModels.PermissionDeniedServerError()
	}
	payload, err := h.Service.GetUtilization(ctx, businessID)
	if err != nil {
		return GetUtilizationResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to load utilization",
		}
	}
	return payload, nil
}
