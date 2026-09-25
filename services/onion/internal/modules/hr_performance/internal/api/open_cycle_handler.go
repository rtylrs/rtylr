package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_performance/internal/application"
	"github.com/voxire/onion/internal/modules/hr_performance/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func OpenCycleRoutes() []string { return []string{"/hr/performance/cycles/{id}/open"} }

type OpenCycleHandler struct{ Service *application.Service }

type OpenCycleRequest struct {
	ID int64 `json:"id"`
}

type OpenCycleResponse struct {
	Success bool `json:"success"`
}

func (h *OpenCycleHandler) Serve(ctx context.Context, req *OpenCycleRequest) (OpenCycleResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return OpenCycleResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return OpenCycleResponse{}, onekiterrors.PermissionDenied()
	}

	err := h.Service.OpenReviewCycle(ctx, businessID, req.ID)
	if err != nil {
		if errors.Is(err, domain.ErrReviewCycleNotFound) {
			return OpenCycleResponse{}, onekiterrors.New(http.StatusNotFound, "Review cycle not found")
		}
		if errors.Is(err, domain.ErrCycleNotDraft) {
			return OpenCycleResponse{}, onekiterrors.New(http.StatusBadRequest, "Review cycle is not in draft status")
		}
		logger.Ctx(ctx).Error("failed to open review cycle", logger.Any("owner_id", device.User.UserID), logger.Any("cycle_id", req.ID), logger.Error(err))
		return OpenCycleResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to open review cycle")
	}

	return OpenCycleResponse{Success: true}, nil
}
