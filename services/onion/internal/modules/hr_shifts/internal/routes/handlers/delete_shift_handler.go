package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_shifts/internal/application"
	"github.com/voxire/onion/internal/modules/hr_shifts/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func DeleteShiftRoutes() []string { return []string{"/hr/shifts"} }

type DeleteShiftHandler struct{ Service *application.Service }

type DeleteShiftRequest struct {
	ID int64 `json:"id"`
}

type DeleteShiftResponse struct {
	Success bool `json:"success"`
}

func (h *DeleteShiftHandler) Validate(_ context.Context, req *DeleteShiftRequest) error {
	if req.ID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "id is required")
	}
	return nil
}

func (h *DeleteShiftHandler) Serve(ctx context.Context, req *DeleteShiftRequest) (DeleteShiftResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return DeleteShiftResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireShiftPermission(ctx); err != nil {
		return DeleteShiftResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return DeleteShiftResponse{}, onekiterrors.PermissionDenied()
	}

	if err := h.Service.DeleteShift(ctx, businessID, req.ID); err != nil {
		if errors.Is(err, domain.ErrShiftNotFound) {
			return DeleteShiftResponse{}, onekiterrors.New(http.StatusNotFound, "Shift not found")
		}
		logger.Ctx(ctx).Error("failed to delete shift",
			logger.Any("business_id", businessID), logger.Any("shift_id", req.ID), logger.Error(err))
		return DeleteShiftResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to delete shift")
	}

	return DeleteShiftResponse{Success: true}, nil
}
