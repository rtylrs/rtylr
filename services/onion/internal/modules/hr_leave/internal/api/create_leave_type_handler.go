package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/hr_leave/internal/application"
	"github.com/voxire/onion/internal/modules/hr_leave/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	hrleavev1 "github.com/voxire/voxie/gen/go/hr/leave/v1"
)

func CreateLeaveTypeRoutes() []string { return []string{"/hr/leave/types"} }

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
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return onekiterrors.New(http.StatusBadRequest, "name is required")
	}
	return nil
}

func (h *CreateLeaveTypeHandler) Serve(ctx context.Context, req *CreateLeaveTypeRequest) (CreateLeaveTypeResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CreateLeaveTypeResponse{}, onekiterrors.PermissionDenied()
	}

	lt, err := h.Service.CreateLeaveType(ctx, device.User.UserID, domain.LeaveTypeRecord{
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
		logger.Ctx(ctx).Error("failed to create leave type", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return CreateLeaveTypeResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to create leave type")
	}

	return CreateLeaveTypeResponse{LeaveType: leaveTypeToProto(*lt)}, nil
}
