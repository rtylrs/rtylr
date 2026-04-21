package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/insights/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

// GET /insights/customers/churn

func GetChurnRoutes() []string { return []string{"/insights/customers/churn"} }

type GetChurnHandler struct{ Service *application.Service }

type GetChurnRequest struct{}

type GetChurnResponse struct {
	Windows []application.ChurnWindowPayload `json:"windows"`
}

func (h *GetChurnHandler) Serve(ctx context.Context, _ *GetChurnRequest) (GetChurnResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetChurnResponse{}, errorModels.PermissionDeniedServerError()
	}
	windows, err := h.Service.GetChurnRates(ctx, device.User.UserID)
	if err != nil {
		return GetChurnResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to load churn rates",
		}
	}
	return GetChurnResponse{Windows: windows}, nil
}
