package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/user/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func ChangePasswordRoutes() []string {
	return []string{
		"/account/password",
	}
}

type ChangePasswordHandler struct {
	UserService *application.Service
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type ChangePasswordResponse struct{}

func (h *ChangePasswordHandler) Validate(_ context.Context, req *ChangePasswordRequest) *errorModels.ServerError {
	if req.CurrentPassword == "" || req.NewPassword == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Current password and new password are required",
		}
	}

	return nil
}

func (h *ChangePasswordHandler) Serve(ctx context.Context, req *ChangePasswordRequest) (ChangePasswordResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)

	err := h.UserService.ChangePassword(ctx, device.User.UserID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		if errors.Is(err, application.ErrInvalidCredentials) {
			return ChangePasswordResponse{}, &errorModels.ServerError{
				Code:    http.StatusUnauthorized,
				Message: "Current password is incorrect",
			}
		}

		return ChangePasswordResponse{}, &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Failed to change password",
		}
	}

	return ChangePasswordResponse{}, nil
}
