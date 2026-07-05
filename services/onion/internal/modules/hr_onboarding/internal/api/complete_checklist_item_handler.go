package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr_onboarding/internal/application"
	"github.com/voxire/onion/internal/modules/hr_onboarding/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
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

func (h *CompleteChecklistItemHandler) Serve(ctx context.Context, _ *CompleteChecklistItemRequest) (CompleteChecklistItemResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CompleteChecklistItemResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CompleteChecklistItemResponse{}, onekiterrors.PermissionDenied()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	itemID, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return CompleteChecklistItemResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid item id")
	}

	err := h.Service.CompleteChecklistItem(ctx, businessID, itemID, device.User.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrChecklistItemNotFound) {
			return CompleteChecklistItemResponse{}, onekiterrors.New(http.StatusNotFound, "Checklist item not found")
		}
		if errors.Is(err, domain.ErrChecklistAlreadyCompleted) {
			return CompleteChecklistItemResponse{}, onekiterrors.New(http.StatusConflict, "Checklist item is already completed")
		}
		logger.Ctx(ctx).Error("failed to complete checklist item", logger.Any("owner_id", device.User.UserID), logger.Any("item_id", itemID), logger.Error(err))
		return CompleteChecklistItemResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to complete checklist item")
	}

	return CompleteChecklistItemResponse{Success: true}, nil
}
