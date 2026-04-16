package handlers

import (
	"context"
	"errors"
	"net/http"

	hrv1 "github.com/voxire/voxie/gen/go/hr/v1"

	"github.com/voxire/onion/internal/modules/hr_leave/internal/application"
	"github.com/voxire/onion/internal/modules/hr_leave/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
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
	Request *hrv1.LeaveRequestPayload `json:"request"`
}

func (h *CreateLeaveRequestHandler) Validate(_ context.Context, req *CreateLeaveRequestRequest) *errorModels.ServerError {
	if req.EmployeeID == 0 {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "employee_id is required"}
	}
	if req.LeaveTypeID == 0 {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "leave_type_id is required"}
	}
	if req.StartDate == "" || req.EndDate == "" {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "start_date and end_date are required"}
	}
	if req.TotalDays <= 0 {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "total_days must be greater than 0"}
	}
	return nil
}

func (h *CreateLeaveRequestHandler) Serve(ctx context.Context, req *CreateLeaveRequestRequest) (CreateLeaveRequestResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CreateLeaveRequestResponse{}, errorModels.PermissionDeniedServerError()
	}

	created, err := h.Service.CreateLeaveRequest(ctx, device.User.UserID, domain.LeaveRequestRecord{
		EmployeeID:  req.EmployeeID,
		LeaveTypeID: req.LeaveTypeID,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		TotalDays:   req.TotalDays,
		Notes:       req.Notes,
	})
	if err != nil {
		if errors.Is(err, domain.ErrLeaveNotConfigured) {
			return CreateLeaveRequestResponse{}, &errorModels.ServerError{HTTPCode: http.StatusUnprocessableEntity, Code: http.StatusUnprocessableEntity, Message: "Leave types are not configured. Please configure leave types first."}
		}
		if errors.Is(err, domain.ErrLeaveTypeNotFound) {
			return CreateLeaveRequestResponse{}, &errorModels.ServerError{HTTPCode: http.StatusNotFound, Code: http.StatusNotFound, Message: "Leave type not found"}
		}
		if errors.Is(err, domain.ErrInsufficientLeaveBalance) {
			return CreateLeaveRequestResponse{}, &errorModels.ServerError{HTTPCode: http.StatusConflict, Code: http.StatusConflict, Message: "Insufficient leave balance"}
		}
		if errors.Is(err, domain.ErrLeaveConflict) {
			return CreateLeaveRequestResponse{}, &errorModels.ServerError{HTTPCode: http.StatusConflict, Code: http.StatusConflict, Message: "Leave request overlaps with existing leave"}
		}
		if errors.Is(err, domain.ErrInvalidDateRange) {
			return CreateLeaveRequestResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid date range"}
		}
		if errors.Is(err, domain.ErrInsufficientNotice) {
			return CreateLeaveRequestResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Insufficient notice period for this leave type"}
		}
		logger.Ctx(ctx).Error("failed to create leave request", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return CreateLeaveRequestResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to create leave request"}
	}

	return CreateLeaveRequestResponse{Request: leaveRequestToProto(*created)}, nil
}
