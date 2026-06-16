package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/user/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func SwitchBusinessRoutes() []string { return []string{"/auth/switch-business"} }

type (
	SwitchBusinessHandler  struct{ UserService *application.Service }
	SwitchBusinessRequest  struct{ BusinessID int64 `json:"business_id"` }
	SwitchBusinessResponse struct{ Token string `json:"token"` }
)

func (h *SwitchBusinessHandler) Validate(_ context.Context, req *SwitchBusinessRequest) *errorModels.ServerError {
	if req.BusinessID <= 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "business_id is required"}
	}
	return nil
}

func (h *SwitchBusinessHandler) Serve(ctx context.Context, req *SwitchBusinessRequest) (SwitchBusinessResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return SwitchBusinessResponse{}, errorModels.PermissionDeniedServerError()
	}

	token, err := h.UserService.SwitchBusiness(ctx, device.User.UserID, req.BusinessID)
	if err != nil {
		if errors.Is(err, application.ErrBusinessAccessDenied) {
			return SwitchBusinessResponse{}, errorModels.PermissionDeniedServerError()
		}
		return SwitchBusinessResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to switch business",
		}
	}
	return SwitchBusinessResponse{Token: token}, nil
}
