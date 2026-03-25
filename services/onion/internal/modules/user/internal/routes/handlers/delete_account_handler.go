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

func DeleteAccountRoutes() []string {
	return []string{
		"/account/delete",
	}
}

type DeleteAccountHandler struct {
	UserService *application.Service
}

func (h *DeleteAccountHandler) Validate(_ context.Context, req *accountv1.DeleteAccountRequest) *errorModels.ServerError {
	if req.Password == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Password is required",
		}
	}

	return nil
}

func (h *DeleteAccountHandler) Serve(ctx context.Context, req *accountv1.DeleteAccountRequest) (accountv1.DeleteAccountResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)

	err := h.UserService.DeleteAccount(ctx, device.User.UserID, req.Password)
	if err != nil {
		if errors.Is(err, application.ErrInvalidCredentials) {
			return accountv1.DeleteAccountResponse{}, &errorModels.ServerError{
				Code:    http.StatusUnauthorized,
				Message: "Password is incorrect",
			}
		}

		return accountv1.DeleteAccountResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to delete account",
		}
	}

	return accountv1.DeleteAccountResponse{}, nil
}
