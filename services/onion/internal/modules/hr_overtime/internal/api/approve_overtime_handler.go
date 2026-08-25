package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_overtime/internal/application"
	"github.com/voxire/onion/internal/modules/hr_overtime/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
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

func (h *ApproveOvertimeHandler) Validate(_ context.Context, req *ApproveOvertimeRequestReq) error {
	if req.RequestID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "request_id is required")
	}
	return nil
}

func (h *ApproveOvertimeHandler) Serve(ctx context.Context, req *ApproveOvertimeRequestReq) (ApproveOvertimeRequestRes, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ApproveOvertimeRequestRes{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ApproveOvertimeRequestRes{}, onekiterrors.PermissionDenied()
	}

	var err error
	if req.Approved {
		err = h.Service.ApproveOvertime(ctx, businessID, req.RequestID, device.User.UserID, req.MultiplierApplied)
	} else {
		err = h.Service.RejectOvertime(ctx, businessID, req.RequestID, device.User.UserID)
	}

	if err != nil {
		if errors.Is(err, domain.ErrOvertimeRequestNotFound) {
			return ApproveOvertimeRequestRes{}, onekiterrors.New(http.StatusNotFound, "Overtime request not found")
		}
		if errors.Is(err, domain.ErrApprovalChainNotConfigured) {
			return ApproveOvertimeRequestRes{}, onekiterrors.New(http.StatusUnprocessableEntity, "Approval chain not configured for overtime requests")
		}
		if errors.Is(err, domain.ErrOvertimeRequestNotPending) {
			return ApproveOvertimeRequestRes{}, onekiterrors.New(http.StatusConflict, "Overtime request has already been processed")
		}
		if errors.Is(err, domain.ErrOvertimeAccessDenied) || errors.Is(err, domain.ErrOvertimeApproverNotAllowed) || errors.Is(err, domain.ErrOvertimeSelfApproval) {
			return ApproveOvertimeRequestRes{}, onekiterrors.PermissionDenied()
		}
		if errors.Is(err, domain.ErrOvertimeConfigNotFound) {
			return ApproveOvertimeRequestRes{}, onekiterrors.New(http.StatusUnprocessableEntity, "Overtime configuration is required before approval")
		}
		logger.Ctx(ctx).Error("failed to process overtime approval", logger.Any("business_id", businessID), logger.Any("request_id", req.RequestID), logger.Error(err))
		return ApproveOvertimeRequestRes{}, onekiterrors.New(http.StatusInternalServerError, "Failed to process overtime approval")
	}

	return ApproveOvertimeRequestRes{Success: true}, nil
}
