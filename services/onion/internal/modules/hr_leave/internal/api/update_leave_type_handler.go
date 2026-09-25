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
	"github.com/voxire/onion/internal/platform/tenant"
	hrleavev1 "github.com/voxire/voxie/gen/go/hr/leave/v1"
)

func UpdateLeaveTypeRoutes() []string { return []string{"/hr/leave/types/{id}"} }

type UpdateLeaveTypeHandler struct{ Service *application.Service }

// UpdateLeaveTypeRequest intentionally keeps the generated patch fields as
// pointers. A nil field means omitted; false and zero are meaningful values.
type UpdateLeaveTypeRequest struct {
	ID                             int64   `json:"id"`
	Name                           *string `json:"name"`
	IsPaid                         *bool   `json:"is_paid"`
	RequiresApproval               *bool   `json:"requires_approval"`
	AccrualRule                    *string `json:"accrual_rule"`
	MaxCarryoverDays               *int32  `json:"max_carryover_days"`
	MinNoticeDays                  *int32  `json:"min_notice_days"`
	RequiresDocumentationAfterDays *int32  `json:"requires_documentation_after_days"`
	IsActive                       *bool   `json:"is_active"`
}

type UpdateLeaveTypeResponse struct {
	Success   bool                        `json:"success"`
	LeaveType *hrleavev1.LeaveTypePayload `json:"leave_type,omitempty"`
}

func (h *UpdateLeaveTypeHandler) Validate(_ context.Context, req *UpdateLeaveTypeRequest) error {
	if req == nil {
		return onekiterrors.New(http.StatusBadRequest, "request body is required")
	}
	if req.ID <= 0 {
		return onekiterrors.New(http.StatusBadRequest, "id must be greater than zero")
	}
	if req.Name == nil && req.IsPaid == nil && req.RequiresApproval == nil &&
		req.AccrualRule == nil && req.MaxCarryoverDays == nil && req.MinNoticeDays == nil &&
		req.RequiresDocumentationAfterDays == nil && req.IsActive == nil {
		return onekiterrors.New(http.StatusBadRequest, "at least one leave type field is required")
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return onekiterrors.New(http.StatusBadRequest, "name cannot be empty")
		}
		req.Name = &name
	}
	if req.AccrualRule != nil {
		rule := strings.ToLower(strings.TrimSpace(*req.AccrualRule))
		if rule != "" && !domain.IsValidAccrualRule(rule) {
			return onekiterrors.New(http.StatusBadRequest, "accrual_rule must be none, monthly, or annual")
		}
		req.AccrualRule = &rule
	}
	if (req.MaxCarryoverDays != nil && *req.MaxCarryoverDays < 0) ||
		(req.MinNoticeDays != nil && *req.MinNoticeDays < 0) ||
		(req.RequiresDocumentationAfterDays != nil && *req.RequiresDocumentationAfterDays < 0) {
		return onekiterrors.New(http.StatusBadRequest, "leave policy day limits must be non-negative")
	}
	return nil
}

func (h *UpdateLeaveTypeHandler) Serve(ctx context.Context, req *UpdateLeaveTypeRequest) (UpdateLeaveTypeResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return UpdateLeaveTypeResponse{}, onekiterrors.PermissionDenied()
	}
	if err := h.Validate(ctx, req); err != nil {
		return UpdateLeaveTypeResponse{}, err
	}
	if !canManageLeavePolicy(device.User) {
		return UpdateLeaveTypeResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdateLeaveTypeResponse{}, onekiterrors.PermissionDenied()
	}

	updated, err := h.Service.UpdateLeaveTypePatch(ctx, businessID, req.ID, domain.LeaveTypePatch{
		Name:                           req.Name,
		IsPaid:                         req.IsPaid,
		RequiresApproval:               req.RequiresApproval,
		AccrualRule:                    req.AccrualRule,
		MaxCarryoverDays:               req.MaxCarryoverDays,
		MinNoticeDays:                  req.MinNoticeDays,
		RequiresDocumentationAfterDays: req.RequiresDocumentationAfterDays,
		IsActive:                       req.IsActive,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrLeaveTypeNotFound):
			return UpdateLeaveTypeResponse{}, onekiterrors.New(http.StatusNotFound, "Leave type not found")
		case errors.Is(err, domain.ErrEmptyLeaveTypeUpdate),
			errors.Is(err, domain.ErrEmptyLeaveTypeName),
			errors.Is(err, domain.ErrNegativeLeavePolicyLimit),
			errors.Is(err, domain.ErrInvalidAccrualRule),
			errors.Is(err, domain.ErrNegativeAccrual):
			return UpdateLeaveTypeResponse{}, onekiterrors.New(http.StatusBadRequest, "invalid leave policy")
		default:
			logger.Ctx(ctx).Error("failed to update leave type", logger.Any("business_id", businessID), logger.Error(err))
			return UpdateLeaveTypeResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update leave type")
		}
	}

	return UpdateLeaveTypeResponse{Success: true, LeaveType: leaveTypeToProto(*updated)}, nil
}
