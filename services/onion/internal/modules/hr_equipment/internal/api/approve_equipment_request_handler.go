package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr_equipment/internal/application"
	"github.com/voxire/onion/internal/modules/hr_equipment/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func ApproveEquipmentRequestRoutes() []string {
	return []string{"/hr/equipment/requests/{id}/approve"}
}

type ApproveEquipmentRequestHandler struct{ Service *application.Service }

type ApproveEquipmentRequestRequest struct {
	Approve bool   `json:"approve"`
	Reason  string `json:"reason"`
}

type ApproveEquipmentRequestResponse struct {
	Success bool `json:"success"`
}

func (h *ApproveEquipmentRequestHandler) Validate(_ context.Context, req *ApproveEquipmentRequestRequest) error {
	if !req.Approve && req.Reason == "" {
		return onekiterrors.New(http.StatusBadRequest, "reason is required when rejecting")
	}
	return nil
}

func (h *ApproveEquipmentRequestHandler) Serve(ctx context.Context, req *ApproveEquipmentRequestRequest) (ApproveEquipmentRequestResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ApproveEquipmentRequestResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ApproveEquipmentRequestResponse{}, onekiterrors.PermissionDenied()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	requestID, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return ApproveEquipmentRequestResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid request id")
	}

	err := h.Service.ApproveEquipmentRequest(ctx, businessID, requestID, device.User.UserID, req.Approve, req.Reason)
	if err != nil {
		if errors.Is(err, domain.ErrEquipmentRequestNotFound) {
			return ApproveEquipmentRequestResponse{}, onekiterrors.New(http.StatusNotFound, "Equipment request not found")
		}
		logger.Ctx(ctx).Error("failed to approve equipment request", logger.Any("owner_id", device.User.UserID), logger.Any("request_id", requestID), logger.Error(err))
		return ApproveEquipmentRequestResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to process equipment request")
	}

	return ApproveEquipmentRequestResponse{Success: true}, nil
}
