package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_performance/internal/application"
	"github.com/voxire/onion/internal/modules/hr_performance/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
)

func OpenCycleRoutes() []string { return []string{"/hr/performance/cycles/{id}/open"} }

type OpenCycleHandler struct{ Service *application.Service }

type OpenCycleRequest struct {
	ID int64 `json:"id"`
}

type OpenCycleResponse struct {
	Success bool `json:"success"`
}

func (h *OpenCycleHandler) Serve(ctx context.Context, req *OpenCycleRequest) (OpenCycleResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return OpenCycleResponse{}, errorModels.PermissionDeniedServerError()
	}

	err := h.Service.OpenReviewCycle(ctx, device.User.UserID, req.ID)
	if err != nil {
		if errors.Is(err, domain.ErrReviewCycleNotFound) {
			return OpenCycleResponse{}, &errorModels.ServerError{HTTPCode: http.StatusNotFound, Code: http.StatusNotFound, Message: "Review cycle not found"}
		}
		if errors.Is(err, domain.ErrCycleNotDraft) {
			return OpenCycleResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Review cycle is not in draft status"}
		}
		logger.Ctx(ctx).Error("failed to open review cycle", logger.Any("owner_id", device.User.UserID), logger.Any("cycle_id", req.ID), logger.Error(err))
		return OpenCycleResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to open review cycle"}
	}

	return OpenCycleResponse{Success: true}, nil
}
