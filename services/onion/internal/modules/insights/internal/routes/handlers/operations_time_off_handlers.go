package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/insights/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// GET /insights/operations/time-off

func GetTimeOffRoutes() []string { return []string{"/insights/operations/time-off"} }

type GetTimeOffHandler struct{ Service *application.Service }

type GetTimeOffRequest struct{}

type GetTimeOffResponse = application.TimeOffPayload

func (h *GetTimeOffHandler) Serve(ctx context.Context, _ *GetTimeOffRequest) (GetTimeOffResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetTimeOffResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetTimeOffResponse{}, errorModels.PermissionDeniedServerError()
	}
	payload, err := h.Service.GetTimeOffSummary(ctx, businessID)
	if err != nil {
		return GetTimeOffResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to load time-off summary",
		}
	}
	return payload, nil
}
