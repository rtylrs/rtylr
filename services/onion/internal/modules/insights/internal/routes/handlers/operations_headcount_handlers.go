package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/insights/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

// GET /insights/operations/headcount

func GetHeadcountRoutes() []string { return []string{"/insights/operations/headcount"} }

type GetHeadcountHandler struct{ Service *application.Service }

type GetHeadcountRequest struct{}

type GetHeadcountResponse = application.HeadcountPayload

func (h *GetHeadcountHandler) Serve(ctx context.Context, _ *GetHeadcountRequest) (GetHeadcountResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetHeadcountResponse{}, errorModels.PermissionDeniedServerError()
	}
	payload, err := h.Service.GetHeadcount(ctx, device.User.UserID)
	if err != nil {
		return GetHeadcountResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to load headcount",
		}
	}
	return payload, nil
}
