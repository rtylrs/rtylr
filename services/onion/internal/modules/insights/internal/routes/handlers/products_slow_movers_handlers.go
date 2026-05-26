package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/voxire/onion/internal/modules/insights/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

// GET /insights/products/slow-movers?days=60

func GetSlowMoversRoutes() []string { return []string{"/insights/products/slow-movers"} }

type GetSlowMoversHandler struct{ Service *application.Service }

type GetSlowMoversRequest struct {
	Days string `json:"days"`
}

type GetSlowMoversResponse struct {
	Products []application.SlowMoverPayload `json:"products"`
}

func (h *GetSlowMoversHandler) Serve(ctx context.Context, req *GetSlowMoversRequest) (GetSlowMoversResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetSlowMoversResponse{}, errorModels.PermissionDeniedServerError()
	}
	var days int32 = 60
	if req.Days != "" {
		parsed, err := strconv.ParseInt(req.Days, 10, 32)
		if err != nil || parsed <= 0 {
			return GetSlowMoversResponse{}, &errorModels.ServerError{
				HTTPCode: http.StatusBadRequest,
				Code:     http.StatusBadRequest,
				Message:  "Invalid days",
			}
		}
		days = int32(parsed)
	}
	rows, err := h.Service.GetSlowMovers(ctx, device.User.UserID, days)
	if err != nil {
		return GetSlowMoversResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to load slow movers",
		}
	}
	return GetSlowMoversResponse{Products: rows}, nil
}
