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
)

func UpdateLeaveTypeRoutes() []string { return []string{"/hr/leave/types/{id}"} }

type UpdateLeaveTypeHandler struct{ Service *application.Service }

type UpdateLeaveTypeRequest struct {
	ID                             int64   `json:"id"`
	Name                           string  `json:"name"`
	IsPaid                         bool    `json:"is_paid"`
	RequiresApproval               bool    `json:"requires_approval"`
	AccrualRule                    string  `json:"accrual_rule"`
	AccrualAmount                  float64 `json:"accrual_amount"`
	MaxCarryoverDays               int32   `json:"max_carryover_days"`
	MinNoticeDays                  int32   `json:"min_notice_days"`
	RequiresDocumentationAfterDays int32   `json:"requires_documentation_after_days"`
	IsActive                       bool    `json:"is_active"`
}

type UpdateLeaveTypeResponse struct {
	Success bool `json:"success"`
}

func (h *UpdateLeaveTypeHandler) Validate(_ context.Context, req *UpdateLeaveTypeRequest) error {
	if req.ID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "id is required")
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return onekiterrors.New(http.StatusBadRequest, "name is required")
	}
	return nil
}

func (h *UpdateLeaveTypeHandler) Serve(ctx context.Context, req *UpdateLeaveTypeRequest) (UpdateLeaveTypeResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return UpdateLeaveTypeResponse{}, onekiterrors.PermissionDenied()
	}

	err := h.Service.UpdateLeaveType(ctx, device.User.UserID, domain.LeaveTypeRecord{
		ID:                             req.ID,
		Name:                           req.Name,
		IsPaid:                         req.IsPaid,
		RequiresApproval:               req.RequiresApproval,
		AccrualRule:                    req.AccrualRule,
		AccrualAmount:                  req.AccrualAmount,
		MaxCarryoverDays:               req.MaxCarryoverDays,
		MinNoticeDays:                  req.MinNoticeDays,
		RequiresDocumentationAfterDays: req.RequiresDocumentationAfterDays,
		IsActive:                       req.IsActive,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to update leave type", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return UpdateLeaveTypeResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update leave type")
	}

	return UpdateLeaveTypeResponse{Success: true}, nil
}
