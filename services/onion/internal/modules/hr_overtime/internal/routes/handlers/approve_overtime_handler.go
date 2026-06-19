package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_overtime/internal/application"
	"github.com/voxire/onion/internal/modules/hr_overtime/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/tenant"
)

func ApproveOvertimeRoutes() []string { return []string{"/hr/overtime/requests/{id}/approve"} }

type ApproveOvertimeHandler struct{ Service *application.Service }

type ApproveOvertimeRequestReq struct {
	RequestID         int64   `json:"request_id"`
	Approved          bool    `json:"approved"`
	MultiplierApplied float64 `json:"multiplier_applied"`
}

type ApproveOvertimeRequestRes struct {
	Success bool `json:"success"`
}

func (h *ApproveOvertimeHandler) Validate(_ context.Context, req *ApproveOvertimeRequestReq) *errorModels.ServerError {
	if req.RequestID == 0 {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "request_id is required"}
	}
	return nil
}

func (h *ApproveOvertimeHandler) Serve(ctx context.Context, req *ApproveOvertimeRequestReq) (ApproveOvertimeRequestRes, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ApproveOvertimeRequestRes{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ApproveOvertimeRequestRes{}, errorModels.PermissionDeniedServerError()
	}

	var err error
	if req.Approved {
		err = h.Service.ApproveOvertime(ctx, businessID, req.RequestID, device.User.UserID, req.MultiplierApplied)
	} else {
		err = h.Service.RejectOvertime(ctx, businessID, req.RequestID)
	}

	if err != nil {
		if errors.Is(err, domain.ErrOvertimeRequestNotFound) {
			return ApproveOvertimeRequestRes{}, &errorModels.ServerError{HTTPCode: http.StatusNotFound, Code: http.StatusNotFound, Message: "Overtime request not found"}
		}
		if errors.Is(err, domain.ErrApprovalChainNotConfigured) {
			return ApproveOvertimeRequestRes{}, &errorModels.ServerError{HTTPCode: http.StatusUnprocessableEntity, Code: http.StatusUnprocessableEntity, Message: "Approval chain not configured for overtime requests"}
		}
		logger.Ctx(ctx).Error("failed to process overtime approval", logger.Any("business_id", businessID), logger.Any("request_id", req.RequestID), logger.Error(err))
		return ApproveOvertimeRequestRes{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to process overtime approval"}
	}

	return ApproveOvertimeRequestRes{Success: true}, nil
}
