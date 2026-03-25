package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/user/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	accountv1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/shared/account/v1"
)

func ChangePasswordRoutes() []string {
	return []string{
		"/account/password",
	}
}

type ChangePasswordHandler struct {
	UserService *application.Service
}

func (h *ChangePasswordHandler) Validate(_ context.Context, req *accountv1.ChangePasswordRequest) *errorModels.ServerError {
	if req.CurrentPassword == "" || req.NewPassword == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Current password and new password are required",
		}
	}

	return nil
}

func (h *ChangePasswordHandler) Serve(ctx context.Context, req *accountv1.ChangePasswordRequest) (accountv1.ChangePasswordResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)

	err := h.UserService.ChangePassword(ctx, device.User.UserID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		if errors.Is(err, application.ErrInvalidCredentials) {
			return accountv1.ChangePasswordResponse{}, &errorModels.ServerError{
				Code:    http.StatusUnauthorized,
				Message: "Current password is incorrect",
			}
		}

		return accountv1.ChangePasswordResponse{}, &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Failed to change password",
		}
	}

	return accountv1.ChangePasswordResponse{}, nil
}
