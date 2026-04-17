package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_shifts/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
)

func PublishRosterRoutes() []string { return []string{"/hr/shifts/roster/publish"} }

type PublishRosterHandler struct{ Service *application.Service }

type PublishRosterRequest struct {
	LocationID int64  `json:"location_id"`
	WeekStart  string `json:"week_start"`
}

type PublishRosterResponse struct {
	Success bool `json:"success"`
}

func (h *PublishRosterHandler) Validate(_ context.Context, req *PublishRosterRequest) *errorModels.ServerError {
	if req.LocationID == 0 {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "location_id is required"}
	}
	if req.WeekStart == "" {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "week_start is required"}
	}
	return nil
}

func (h *PublishRosterHandler) Serve(ctx context.Context, req *PublishRosterRequest) (PublishRosterResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return PublishRosterResponse{}, errorModels.PermissionDeniedServerError()
	}

	err := h.Service.PublishRoster(ctx, device.User.UserID, req.LocationID, req.WeekStart)
	if err != nil {
		logger.Ctx(ctx).Error("failed to publish roster", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return PublishRosterResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to publish roster"}
	}

	return PublishRosterResponse{Success: true}, nil
}
