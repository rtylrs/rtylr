package api

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/hr_leave/internal/application"
	"github.com/voxire/onion/internal/modules/hr_leave/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	hrleavev1 "github.com/voxire/voxie/gen/go/hr/leave/v1"
)

func CreateLeaveRequestRoutes() []string { return []string{"/hr/leave/requests"} }

type CreateLeaveRequestHandler struct{ Service *application.Service }

type CreateLeaveRequestRequest struct {
	EmployeeID  int64   `json:"employee_id"`
	LeaveTypeID int64   `json:"leave_type_id"`
	StartDate   string  `json:"start_date"`
	EndDate     string  `json:"end_date"`
	TotalDays   float64 `json:"total_days"`
	Notes       string  `json:"notes"`
}

type CreateLeaveRequestResponse struct {
	Request *hrleavev1.LeaveRequestPayload `json:"request"`
}

func (h *CreateLeaveRequestHandler) Validate(_ context.Context, req *CreateLeaveRequestRequest) error {
	if req.EmployeeID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "employee_id is required")
	}
	if req.LeaveTypeID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "leave_type_id is required")
	}
	req.StartDate = strings.TrimSpace(req.StartDate)
	req.EndDate = strings.TrimSpace(req.EndDate)
	if req.StartDate == "" || req.EndDate == "" {
		return onekiterrors.New(http.StatusBadRequest, "start_date and end_date are required")
	}
	if _, err := time.Parse("2006-01-02", req.StartDate); err != nil {
		return onekiterrors.New(http.StatusBadRequest, "start_date must be in YYYY-MM-DD format")
	}
	if _, err := time.Parse("2006-01-02", req.EndDate); err != nil {
		return onekiterrors.New(http.StatusBadRequest, "end_date must be in YYYY-MM-DD format")
	}
	// Keep rejecting missing/invalid wire values for a clear API contract, but
	// the service still recomputes the authoritative total from the dates.
	if req.TotalDays <= 0 {
		return onekiterrors.New(http.StatusBadRequest, "total_days must be greater than 0")
	}
	return nil
}

func (h *CreateLeaveRequestHandler) Serve(ctx context.Context, req *CreateLeaveRequestRequest) (CreateLeaveRequestResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CreateLeaveRequestResponse{}, onekiterrors.PermissionDenied()
	}

	created, err := h.Service.CreateLeaveRequest(ctx, device.User.UserID, device.User.UserID, domain.LeaveRequestRecord{
		EmployeeID:  req.EmployeeID,
		LeaveTypeID: req.LeaveTypeID,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		TotalDays:   req.TotalDays,
		Notes:       req.Notes,
	})
	if err != nil {
		if errors.Is(err, domain.ErrLeaveAccessDenied) {
			return CreateLeaveRequestResponse{}, onekiterrors.PermissionDenied()
		}
		if errors.Is(err, domain.ErrLeaveNotConfigured) {
			return CreateLeaveRequestResponse{}, onekiterrors.New(http.StatusUnprocessableEntity, "Leave types are not configured. Please configure leave types first.")
		}
		if errors.Is(err, domain.ErrLeaveTypeNotFound) {
			return CreateLeaveRequestResponse{}, onekiterrors.New(http.StatusNotFound, "Leave type not found")
		}
		if errors.Is(err, domain.ErrInsufficientLeaveBalance) {
			return CreateLeaveRequestResponse{}, onekiterrors.New(http.StatusConflict, "Insufficient leave balance")
		}
		if errors.Is(err, domain.ErrLeaveBalanceNotFound) {
			return CreateLeaveRequestResponse{}, onekiterrors.New(http.StatusConflict, "Leave balance is not configured for this employee")
		}
		if errors.Is(err, domain.ErrLeaveConflict) {
			return CreateLeaveRequestResponse{}, onekiterrors.New(http.StatusConflict, "Leave request overlaps with existing leave")
		}
		if errors.Is(err, domain.ErrInvalidDateRange) {
			return CreateLeaveRequestResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid date range")
		}
		if errors.Is(err, domain.ErrInsufficientNotice) {
			return CreateLeaveRequestResponse{}, onekiterrors.New(http.StatusBadRequest, "Insufficient notice period for this leave type")
		}
		logger.Ctx(ctx).Error("failed to create leave request", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return CreateLeaveRequestResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to create leave request")
	}

	return CreateLeaveRequestResponse{Request: leaveRequestToProto(*created)}, nil
}
