package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/hr_shifts/internal/application"
	"github.com/voxire/onion/internal/modules/hr_shifts/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hrshiftv1 "github.com/voxire/voxie/gen/go/hr/shift/v1"
)

func CreateShiftAssignmentRoutes() []string { return []string{"/hr/shifts/roster/assignments"} }

type CreateShiftAssignmentHandler struct{ Service *application.Service }

type CreateShiftAssignmentRequest struct {
	EmployeeID int64  `json:"employee_id"`
	ShiftID    int64  `json:"shift_id"`
	Date       string `json:"date"`
}

type CreateShiftAssignmentResponse struct {
	Assignment *hrshiftv1.ShiftAssignmentPayload `json:"assignment"`
	// Warnings carries advisory availability conflicts. The assignment was still
	// created; see Service.CreateShiftAssignment for why these do not block.
	Warnings []string `json:"warnings,omitempty"`
}

func (h *CreateShiftAssignmentHandler) Validate(_ context.Context, req *CreateShiftAssignmentRequest) error {
	if req.EmployeeID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "employee_id is required")
	}
	if req.ShiftID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "shift_id is required")
	}
	if req.Date == "" {
		return onekiterrors.New(http.StatusBadRequest, "date is required")
	}
	if _, err := time.Parse("2006-01-02", req.Date); err != nil {
		return onekiterrors.New(http.StatusBadRequest, "date must be in YYYY-MM-DD format")
	}
	return nil
}

func (h *CreateShiftAssignmentHandler) Serve(ctx context.Context, req *CreateShiftAssignmentRequest) (CreateShiftAssignmentResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CreateShiftAssignmentResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireShiftPermission(ctx); err != nil {
		return CreateShiftAssignmentResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateShiftAssignmentResponse{}, onekiterrors.PermissionDenied()
	}

	created, warnings, err := h.Service.CreateShiftAssignment(ctx, businessID, domain.ShiftAssignmentRecord{
		EmployeeID: req.EmployeeID,
		ShiftID:    req.ShiftID,
		Date:       req.Date,
	})
	if err != nil {
		if errors.Is(err, domain.ErrShiftNotFound) {
			return CreateShiftAssignmentResponse{}, onekiterrors.New(http.StatusNotFound, "Shift not found")
		}
		if errors.Is(err, domain.ErrRosterConflict) {
			return CreateShiftAssignmentResponse{}, onekiterrors.New(http.StatusConflict, "Shift assignment conflicts with approved leave or another shift")
		}
		logger.Ctx(ctx).Error("failed to create shift assignment", logger.Any("business_id", businessID), logger.Error(err))
		return CreateShiftAssignmentResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to create shift assignment")
	}

	return CreateShiftAssignmentResponse{Assignment: assignmentToProto(*created), Warnings: warnings}, nil
}
