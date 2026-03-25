package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/user/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	accountv1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/shared/account/v1"
	"github.com/voxire/onion/internal/platform/utils/converter"
)

func UpdateProfileRoutes() []string {
	return []string{
		"/account/profile",
	}
}

type UpdateProfileHandler struct {
	UserService *application.Service
}

func (h *UpdateProfileHandler) Validate(_ context.Context, req *accountv1.UpdateProfileRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Name is required",
		}
	}

	return nil
}

func (h *UpdateProfileHandler) Serve(ctx context.Context, req *accountv1.UpdateProfileRequest) (accountv1.UpdateProfileResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)

	user, err := h.UserService.UpdateProfile(ctx, device.User.UserID, req.Name)
	if err != nil {
		return accountv1.UpdateProfileResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to update profile",
		}
	}

	return accountv1.UpdateProfileResponse{
		User: converter.Convert[*accountv1.User](user),
	}, nil
}
