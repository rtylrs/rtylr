package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr_time/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
	"github.com/voxire/onion/internal/platform/logger"
)

func DeleteTimeEntryRoutes() []string { return []string{"/hr/time-entries/{id}"} }

type DeleteTimeEntryHandler struct{ Service *application.Service }

type DeleteTimeEntryRequest struct{}

type DeleteTimeEntryResponse struct {
	Success bool `json:"success"`
}

func (h *DeleteTimeEntryHandler) Validate(_ context.Context, _ *DeleteTimeEntryRequest) *errorModels.ServerError {
	return nil
}

func (h *DeleteTimeEntryHandler) Serve(ctx context.Context, _ *DeleteTimeEntryRequest) (DeleteTimeEntryResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return DeleteTimeEntryResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return DeleteTimeEntryResponse{}, errorModels.PermissionDeniedServerError()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return DeleteTimeEntryResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid time entry id"}
	}

	err := h.Service.DeleteTimeEntry(ctx, businessID, id)
	if err != nil {
		logger.Ctx(ctx).Error("failed to delete time entry", logger.Any("owner_id", device.User.UserID), logger.Any("entry_id", id), logger.Error(err))
		return DeleteTimeEntryResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to delete time entry"}
	}

	return DeleteTimeEntryResponse{Success: true}, nil
}
