package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/accounts/internal/application"
	"github.com/voxire/onion/internal/modules/accounts/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/tenant"
)

func DeleteAccountRoutes() []string { return []string{"/finance/accounts/{id}"} }

type DeleteAccountHandler struct{ Service *application.Service }

type DeleteAccountRequest struct{}

type DeleteAccountResponse struct {
	Message string `json:"message"`
}

func (h *DeleteAccountHandler) Serve(ctx context.Context, _ *DeleteAccountRequest) (DeleteAccountResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return DeleteAccountResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return DeleteAccountResponse{}, errorModels.PermissionDeniedServerError()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return DeleteAccountResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid account id"}
	}

	if err := h.Service.Delete(ctx, businessID, id); err != nil {
		if errors.Is(err, domain.ErrAccountNotFound) {
			return DeleteAccountResponse{}, &errorModels.ServerError{HTTPCode: http.StatusNotFound, Code: http.StatusNotFound, Message: "Account not found"}
		}
		logger.Ctx(ctx).Error("failed to delete account", logger.Any("owner_id", device.User.UserID), logger.Any("account_id", id), logger.Error(err))
		return DeleteAccountResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to delete account"}
	}

	return DeleteAccountResponse{Message: "Account deleted successfully"}, nil
}
