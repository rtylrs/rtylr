package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/insights/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

// GET /insights/operations/shifts?from=&to=

func GetShiftMetricsRoutes() []string { return []string{"/insights/operations/shifts"} }

type GetShiftMetricsHandler struct{ Service *application.Service }

type GetShiftMetricsRequest struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type GetShiftMetricsResponse = application.ShiftMetricsPayload

func (h *GetShiftMetricsHandler) Serve(ctx context.Context, req *GetShiftMetricsRequest) (GetShiftMetricsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetShiftMetricsResponse{}, errorModels.PermissionDeniedServerError()
	}
	from, to, err := application.ResolveDateRange(req.From, req.To)
	if err != nil {
		return GetShiftMetricsResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Invalid date range",
		}
	}
	payload, err := h.Service.GetShiftMetrics(ctx, device.User.UserID, from, to)
	if err != nil {
		return GetShiftMetricsResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to load shift metrics",
		}
	}
	return payload, nil
}
