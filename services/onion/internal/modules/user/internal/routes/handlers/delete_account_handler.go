package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/user/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func DeleteAccountRoutes() []string {
	return []string{
		"/account/delete",
	}
}

type DeleteAccountHandler struct {
	UserService *application.Service
}

type DeleteAccountRequest struct {
	Password string `json:"password"`
}

type DeleteAccountResponse struct{}

func (h *DeleteAccountHandler) Validate(_ context.Context, req *DeleteAccountRequest) *errorModels.ServerError {
	if req.Password == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Password is required",
		}
	}

	return nil
}

func (h *DeleteAccountHandler) Serve(ctx context.Context, req *DeleteAccountRequest) (DeleteAccountResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)

	err := h.UserService.DeleteAccount(ctx, device.User.UserID, req.Password)
	if err != nil {
		if errors.Is(err, application.ErrInvalidCredentials) {
			return DeleteAccountResponse{}, &errorModels.ServerError{
				Code:    http.StatusUnauthorized,
				Message: "Password is incorrect",
			}
		}

		return DeleteAccountResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to delete account",
		}
	}

	return DeleteAccountResponse{}, nil
}
