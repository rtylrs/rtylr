package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr_onboarding/internal/application"
	"github.com/voxire/onion/internal/modules/hr_onboarding/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/tenant"
)

func CompleteChecklistItemRoutes() []string {
	return []string{"/hr/onboarding/checklists/items/{id}/complete"}
}

type CompleteChecklistItemHandler struct{ Service *application.Service }

type CompleteChecklistItemRequest struct{}

type CompleteChecklistItemResponse struct {
	Success bool `json:"success"`
}

func (h *CompleteChecklistItemHandler) Serve(ctx context.Context, _ *CompleteChecklistItemRequest) (CompleteChecklistItemResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CompleteChecklistItemResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CompleteChecklistItemResponse{}, errorModels.PermissionDeniedServerError()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	itemID, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return CompleteChecklistItemResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid item id"}
	}

	err := h.Service.CompleteChecklistItem(ctx, businessID, itemID, device.User.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrChecklistItemNotFound) {
			return CompleteChecklistItemResponse{}, &errorModels.ServerError{HTTPCode: http.StatusNotFound, Code: http.StatusNotFound, Message: "Checklist item not found"}
		}
		if errors.Is(err, domain.ErrChecklistAlreadyCompleted) {
			return CompleteChecklistItemResponse{}, &errorModels.ServerError{HTTPCode: http.StatusConflict, Code: http.StatusConflict, Message: "Checklist item is already completed"}
		}
		logger.Ctx(ctx).Error("failed to complete checklist item", logger.Any("owner_id", device.User.UserID), logger.Any("item_id", itemID), logger.Error(err))
		return CompleteChecklistItemResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to complete checklist item"}
	}

	return CompleteChecklistItemResponse{Success: true}, nil
}
