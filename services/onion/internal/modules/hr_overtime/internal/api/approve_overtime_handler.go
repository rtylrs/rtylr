package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/hr_overtime/internal/application"
	"github.com/voxire/onion/internal/modules/hr_overtime/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func ApproveOvertimeRoutes() []string { return []string{"/hr/overtime/requests/{id}/approve"} }

const overtimeCacheScope = "hr_overtime"

func invalidateOvertimeCache(ctx context.Context, businessID int64) {
	// The write has committed even if the request context is canceled while
	// the response is being returned; do not let that cancellation preserve a
	// stale list entry.
	cacher.InvalidateScope(context.WithoutCancel(ctx), cacheKeys.ScopedPrefix(overtimeCacheScope, businessID))
}

type ApproveOvertimeHandler struct{ Service *application.Service }

type ApproveOvertimeRequestReq struct {
	RequestID         int64   `json:"request_id"`
	Approved          bool    `json:"approved"`
	MultiplierApplied float64 `json:"multiplier_applied"`
	RejectionReason   string  `json:"rejection_reason,omitempty"`
}

type ApproveOvertimeRequestRes struct {
	Success bool `json:"success"`
}

func (h *ApproveOvertimeHandler) Validate(_ context.Context, req *ApproveOvertimeRequestReq) error {
	if req == nil || req.RequestID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "request_id is required")
	}
	if err := domain.ValidateOvertimeDecision(req.Approved, req.RejectionReason); err != nil {
		return onekiterrors.New(http.StatusBadRequest, err.Error())
	}
	req.RejectionReason = strings.TrimSpace(req.RejectionReason)
	return nil
}

func (h *ApproveOvertimeHandler) Serve(ctx context.Context, req *ApproveOvertimeRequestReq) (ApproveOvertimeRequestRes, error) {
	if err := h.Validate(ctx, req); err != nil {
		return ApproveOvertimeRequestRes{}, err
	}

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
		err = h.Service.RejectOvertime(ctx, businessID, req.RequestID, device.User.UserID, req.RejectionReason)
	}
	if err == nil {
		// The service invalidates the scope after final approval/rejection,
		// but an intermediate approval returns before that invalidation. Keep
		// the list response fresh for every successful decision made through
		// this HTTP boundary.
		invalidateOvertimeCache(ctx, businessID)
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
