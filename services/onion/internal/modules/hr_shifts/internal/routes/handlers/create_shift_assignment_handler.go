package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/hr_shifts/internal/application"
	"github.com/voxire/onion/internal/modules/hr_shifts/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
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
}

func (h *CreateShiftAssignmentHandler) Validate(_ context.Context, req *CreateShiftAssignmentRequest) *errorModels.ServerError {
	if req.EmployeeID == 0 {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "employee_id is required"}
	}
	if req.ShiftID == 0 {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "shift_id is required"}
	}
	if req.Date == "" {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "date is required"}
	}
	if _, err := time.Parse("2006-01-02", req.Date); err != nil {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "date must be in YYYY-MM-DD format"}
	}
	return nil
}

func (h *CreateShiftAssignmentHandler) Serve(ctx context.Context, req *CreateShiftAssignmentRequest) (CreateShiftAssignmentResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CreateShiftAssignmentResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateShiftAssignmentResponse{}, errorModels.PermissionDeniedServerError()
	}

	created, err := h.Service.CreateShiftAssignment(ctx, businessID, domain.ShiftAssignmentRecord{
		EmployeeID: req.EmployeeID,
		ShiftID:    req.ShiftID,
		Date:       req.Date,
	})
	if err != nil {
		if errors.Is(err, domain.ErrShiftNotFound) {
			return CreateShiftAssignmentResponse{}, &errorModels.ServerError{HTTPCode: http.StatusNotFound, Code: http.StatusNotFound, Message: "Shift not found"}
		}
		if errors.Is(err, domain.ErrRosterConflict) {
			return CreateShiftAssignmentResponse{}, &errorModels.ServerError{HTTPCode: http.StatusConflict, Code: http.StatusConflict, Message: "Shift assignment conflicts with approved leave or another shift"}
		}
		logger.Ctx(ctx).Error("failed to create shift assignment", logger.Any("business_id", businessID), logger.Error(err))
		return CreateShiftAssignmentResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to create shift assignment"}
	}

	return CreateShiftAssignmentResponse{Assignment: assignmentToProto(*created)}, nil
}
