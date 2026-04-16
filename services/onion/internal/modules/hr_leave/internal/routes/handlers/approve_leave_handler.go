package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_leave/internal/application"
	"github.com/voxire/onion/internal/modules/hr_leave/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
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

func (h *ApproveLeaveHandler) Validate(_ context.Context, req *ApproveLeaveRequest) *errorModels.ServerError {
	if req.RequestID == 0 {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "request_id is required"}
	}
	return nil
}

func (h *ApproveLeaveHandler) Serve(ctx context.Context, req *ApproveLeaveRequest) (ApproveLeaveResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ApproveLeaveResponse{}, errorModels.PermissionDeniedServerError()
	}

	var err error
	if req.Approved {
		err = h.Service.ApproveLeave(ctx, device.User.UserID, req.RequestID, device.User.UserID)
	} else {
		err = h.Service.RejectLeave(ctx, device.User.UserID, req.RequestID, req.RejectionReason)
	}

	if err != nil {
		if errors.Is(err, domain.ErrLeaveRequestNotFound) {
			return ApproveLeaveResponse{}, &errorModels.ServerError{HTTPCode: http.StatusNotFound, Code: http.StatusNotFound, Message: "Leave request not found"}
		}
		if errors.Is(err, domain.ErrApprovalChainNotConfigured) {
			return ApproveLeaveResponse{}, &errorModels.ServerError{HTTPCode: http.StatusUnprocessableEntity, Code: http.StatusUnprocessableEntity, Message: "Approval chain not configured for leave requests"}
		}
		logger.Ctx(ctx).Error("failed to process leave approval", logger.Any("owner_id", device.User.UserID), logger.Any("request_id", req.RequestID), logger.Error(err))
		return ApproveLeaveResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to process leave approval"}
	}

	return ApproveLeaveResponse{Success: true}, nil
}
