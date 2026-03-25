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

func AccountRoutes() []string {
	return []string{
		"/account",
		"/user/account",
	}
}

type AccountHandler struct {
	UserService *application.Service
}

func (h *AccountHandler) Serve(ctx context.Context, _ *accountv1.GetProfileRequest) (accountv1.GetProfileResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)

	user, err := h.UserService.GetUserByID(ctx, device.User.UserID)
	if err != nil {
		return accountv1.GetProfileResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get user account",
		}
	}

	return accountv1.GetProfileResponse{
		User: converter.Convert[*accountv1.User](user),
	}, nil
}
