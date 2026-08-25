package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/hr_leave/internal/application"
	"github.com/voxire/onion/internal/modules/hr_leave/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

func ApproveLeaveRoutes() []string { return []string{"/hr/leave/requests/{id}/approve"} }

type ApproveLeaveHandler struct{ Service *application.Service }

type ApproveLeaveRequest struct {
	RequestID       int64  `json:"request_id"`
	Approved        bool   `json:"approved"`
	RejectionReason string `json:"rejection_reason"`
}

type ApproveLeaveResponse struct {
	Success bool `json:"success"`
}

func (h *ApproveLeaveHandler) Validate(_ context.Context, req *ApproveLeaveRequest) error {
	if req.RequestID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "request_id is required")
	}
	if !req.Approved && strings.TrimSpace(req.RejectionReason) == "" {
		return onekiterrors.New(http.StatusBadRequest, "rejection_reason is required when rejecting a leave request")
	}
	return nil
}

func (h *ApproveLeaveHandler) Serve(ctx context.Context, req *ApproveLeaveRequest) (ApproveLeaveResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ApproveLeaveResponse{}, onekiterrors.PermissionDenied()
	}

	var err error
	if req.Approved {
		err = h.Service.ApproveLeave(ctx, device.User.UserID, req.RequestID, device.User.UserID)
	} else {
		err = h.Service.RejectLeave(ctx, device.User.UserID, req.RequestID, device.User.UserID, req.RejectionReason)
	}

	if err != nil {
		if errors.Is(err, domain.ErrLeaveRequestNotFound) {
			return ApproveLeaveResponse{}, onekiterrors.New(http.StatusNotFound, "Leave request not found")
		}
		if errors.Is(err, domain.ErrLeaveRequestNotPending) {
			return ApproveLeaveResponse{}, onekiterrors.New(http.StatusConflict, "Leave request has already been processed")
		}
		if errors.Is(err, domain.ErrApprovalChainNotConfigured) {
			return ApproveLeaveResponse{}, onekiterrors.New(http.StatusUnprocessableEntity, "Approval chain not configured for leave requests")
		}
		if errors.Is(err, domain.ErrLeaveAccessDenied) || errors.Is(err, domain.ErrLeaveApproverNotAllowed) || errors.Is(err, domain.ErrLeaveSelfApproval) {
			return ApproveLeaveResponse{}, onekiterrors.PermissionDenied()
		}
		logger.Ctx(ctx).Error("failed to process leave approval", logger.Any("owner_id", device.User.UserID), logger.Any("request_id", req.RequestID), logger.Error(err))
		return ApproveLeaveResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to process leave approval")
	}

	return ApproveLeaveResponse{Success: true}, nil
}
