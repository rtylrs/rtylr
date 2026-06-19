package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/accounts/internal/application"
	"github.com/voxire/onion/internal/modules/accounts/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/tenant"
)

func UpdateAccountRoutes() []string { return []string{"/finance/accounts/{id}"} }

type UpdateAccountHandler struct{ Service *application.Service }

type UpdateAccountRequest struct {
	Name     string  `json:"name"`
	Type     string  `json:"type"`
	Currency string  `json:"currency"`
	Balance  float64 `json:"balance"`
	IsActive bool    `json:"is_active"`
}

type UpdateAccountResponse struct {
	Message string `json:"message"`
}

func (h *UpdateAccountHandler) Validate(_ context.Context, req *UpdateAccountRequest) *errorModels.ServerError {
	if strings.TrimSpace(req.Name) == "" {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "name is required"}
	}
	if strings.TrimSpace(req.Type) == "" {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "type is required"}
	}
	return nil
}

func (h *UpdateAccountHandler) Serve(ctx context.Context, req *UpdateAccountRequest) (UpdateAccountResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return UpdateAccountResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdateAccountResponse{}, errorModels.PermissionDeniedServerError()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return UpdateAccountResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid account id"}
	}

	input := domain.AccountInput{
		Name:     req.Name,
		Type:     req.Type,
		Currency: req.Currency,
		Balance:  req.Balance,
		IsActive: req.IsActive,
	}

	if err := h.Service.Update(ctx, businessID, id, input); err != nil {
		if errors.Is(err, domain.ErrNegativeBalance) {
			return UpdateAccountResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Account balance cannot be negative"}
		}
		if errors.Is(err, domain.ErrAccountNotFound) {
			return UpdateAccountResponse{}, &errorModels.ServerError{HTTPCode: http.StatusNotFound, Code: http.StatusNotFound, Message: "Account not found"}
		}
		logger.Ctx(ctx).Error("failed to update account", logger.Any("owner_id", device.User.UserID), logger.Any("account_id", id), logger.Error(err))
		return UpdateAccountResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to update account"}
	}

	return UpdateAccountResponse{Message: "Account updated successfully"}, nil
}
