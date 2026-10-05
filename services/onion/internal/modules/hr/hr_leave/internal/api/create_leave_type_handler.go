package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/hr/hr_leave/internal/application"
	"github.com/voxire/onion/internal/modules/hr/hr_leave/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hrleavev1 "github.com/voxire/voxie/gen/go/hr/leave/v1"
)

type CreateLeaveTypeHandler struct{ Service *application.Service }

type CreateLeaveTypeRequest struct {
	Name                           string  `json:"name"`
	IsPaid                         bool    `json:"is_paid"`
	RequiresApproval               bool    `json:"requires_approval"`
	AccrualRule                    string  `json:"accrual_rule"`
	AccrualAmount                  float64 `json:"accrual_amount"`
	MaxCarryoverDays               int32   `json:"max_carryover_days"`
	MinNoticeDays                  int32   `json:"min_notice_days"`
	RequiresDocumentationAfterDays int32   `json:"requires_documentation_after_days"`
}

type CreateLeaveTypeResponse struct {
	LeaveType *hrleavev1.LeaveTypePayload `json:"leave_type"`
}

func (h *CreateLeaveTypeHandler) Validate(_ context.Context, req *CreateLeaveTypeRequest) error {
	if req == nil {
		return onekiterrors.New(http.StatusBadRequest, "request body is required")
	}
	req.Name = strings.TrimSpace(req.Name)
	req.AccrualRule = strings.ToLower(strings.TrimSpace(req.AccrualRule))
	if err := domain.ValidateLeaveType(domain.LeaveTypeRecord{
		Name:          req.Name,
		AccrualRule:   req.AccrualRule,
		AccrualAmount: req.AccrualAmount,
	}); err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidAccrualRule):
			return onekiterrors.New(http.StatusBadRequest, "accrual_rule must be none, monthly, or annual")
		case errors.Is(err, domain.ErrNegativeAccrual):
			return onekiterrors.New(http.StatusBadRequest, "accrual_amount must be non-negative")
		default:
			return onekiterrors.New(http.StatusBadRequest, "name is required")
		}
	}
	if req.MaxCarryoverDays < 0 || req.MinNoticeDays < 0 || req.RequiresDocumentationAfterDays < 0 {
		return onekiterrors.New(http.StatusBadRequest, "leave policy day limits must be non-negative")
	}
	return nil
}

func canManageLeavePolicy(user *authPkg.AuthUser) bool {
	if authPkg.IsBusinessAdmin(user) {
		return true
	}
	if !authPkg.HasAnyAppScope(user, authorization.AppHR) {
		return false
	}
	return authPkg.HasAppPermission(user, authorization.AppHR, "hr:time_off:approve") ||
		authPkg.HasAppPermission(user, authorization.AppHR, "hr:employees:manage")
}

func (h *CreateLeaveTypeHandler) Serve(ctx context.Context, req *CreateLeaveTypeRequest) (CreateLeaveTypeResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CreateLeaveTypeResponse{}, onekiterrors.PermissionDenied()
	}
	if err := h.Validate(ctx, req); err != nil {
		return CreateLeaveTypeResponse{}, err
	}
	// Leave policies are tenant configuration, not the employee request
	// permission. Keep the elevated check at the direct handler boundary as
	// well as the global route gate so non-HTTP/direct callers cannot bypass it.
	if !canManageLeavePolicy(device.User) {
		return CreateLeaveTypeResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateLeaveTypeResponse{}, onekiterrors.PermissionDenied()
	}

	lt, err := h.Service.CreateLeaveType(ctx, businessID, domain.LeaveTypeRecord{
		Name:                           req.Name,
		IsPaid:                         req.IsPaid,
		RequiresApproval:               req.RequiresApproval,
		AccrualRule:                    req.AccrualRule,
		AccrualAmount:                  req.AccrualAmount,
		MaxCarryoverDays:               req.MaxCarryoverDays,
		MinNoticeDays:                  req.MinNoticeDays,
		RequiresDocumentationAfterDays: req.RequiresDocumentationAfterDays,
	})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidAccrualRule) || errors.Is(err, domain.ErrNegativeAccrual) ||
			errors.Is(err, domain.ErrNegativeLeavePolicyLimit) || errors.Is(err, domain.ErrEmptyLeaveTypeName) {
			return CreateLeaveTypeResponse{}, onekiterrors.New(http.StatusBadRequest, "invalid leave policy")
		}
		logger.Ctx(ctx).Error("failed to create leave type", logger.Any("business_id", businessID), logger.Error(err))
		return CreateLeaveTypeResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to create leave type")
	}

	return CreateLeaveTypeResponse{LeaveType: leaveTypeToProto(*lt)}, nil
}
