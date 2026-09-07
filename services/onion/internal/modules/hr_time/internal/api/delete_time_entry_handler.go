package api

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr_time/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func DeleteTimeEntryRoutes() []string { return []string{"/hr/time-entries/{id}"} }

type DeleteTimeEntryHandler struct{ Service *application.Service }

type DeleteTimeEntryRequest struct{}

type DeleteTimeEntryResponse struct {
	Success bool `json:"success"`
}

func (h *DeleteTimeEntryHandler) Validate(_ context.Context, _ *DeleteTimeEntryRequest) error {
	return nil
}

func (h *DeleteTimeEntryHandler) Serve(ctx context.Context, _ *DeleteTimeEntryRequest) (DeleteTimeEntryResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return DeleteTimeEntryResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return DeleteTimeEntryResponse{}, onekiterrors.PermissionDenied()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return DeleteTimeEntryResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid time entry id")
	}

	err := h.Service.DeleteTimeEntry(ctx, businessID, id)
	if err != nil {
		logger.Ctx(ctx).Error("failed to delete time entry", logger.Any("owner_id", device.User.UserID), logger.Any("entry_id", id), logger.Error(err))
		return DeleteTimeEntryResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to delete time entry")
	}

	return DeleteTimeEntryResponse{Success: true}, nil
}
