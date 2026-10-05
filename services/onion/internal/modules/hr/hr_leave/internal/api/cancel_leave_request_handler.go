package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr/hr_leave/internal/application"
	"github.com/voxire/onion/internal/modules/hr/hr_leave/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

type CancelLeaveRequestHandler struct{ Service *application.Service }

type CancelLeaveRequestRequest struct {
	RequestID int64 `json:"request_id"`
}

type CancelLeaveRequestResponse struct {
	Success bool `json:"success"`
}

func (h *CancelLeaveRequestHandler) Validate(_ context.Context, req *CancelLeaveRequestRequest) error {
	if req.RequestID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "request_id is required")
	}
	return nil
}

func (h *CancelLeaveRequestHandler) Serve(ctx context.Context, req *CancelLeaveRequestRequest) (CancelLeaveRequestResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CancelLeaveRequestResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, businessErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if businessErr != nil {
		return CancelLeaveRequestResponse{}, onekiterrors.PermissionDenied()
	}

	err := h.Service.CancelLeaveRequest(ctx, businessID, device.User.UserID, req.RequestID)
	if err != nil {
		if errors.Is(err, domain.ErrLeaveRequestNotFound) {
			return CancelLeaveRequestResponse{}, onekiterrors.New(http.StatusNotFound, "Leave request not found")
		}
		if errors.Is(err, domain.ErrLeaveRequestNotPending) {
			return CancelLeaveRequestResponse{}, onekiterrors.New(http.StatusConflict, "Leave request has already been processed")
		}
		if errors.Is(err, domain.ErrLeaveAccessDenied) {
			return CancelLeaveRequestResponse{}, onekiterrors.PermissionDenied()
		}
		logger.Ctx(ctx).Error("failed to cancel leave request", logger.Any("owner_id", device.User.UserID), logger.Any("request_id", req.RequestID), logger.Error(err))
		return CancelLeaveRequestResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to cancel leave request")
	}

	return CancelLeaveRequestResponse{Success: true}, nil
}
