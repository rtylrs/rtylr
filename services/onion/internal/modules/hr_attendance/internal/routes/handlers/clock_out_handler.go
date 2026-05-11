package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_attendance/internal/application"
	"github.com/voxire/onion/internal/modules/hr_attendance/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
	hrv1 "github.com/voxire/voxie/gen/go/hr/v1"
)

func ClockOutRoutes() []string { return []string{"/hr/attendance/clock-out"} }

type ClockOutHandler struct{ Service *application.Service }

type ClockOutRequest struct {
	EmployeeID int64 `json:"employee_id"`
}

type ClockOutResponse struct {
	Record *hrv1.AttendanceRecordPayload `json:"record"`
}

func (h *ClockOutHandler) Validate(_ context.Context, req *ClockOutRequest) *errorModels.ServerError {
	if req.EmployeeID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "employee_id is required"}
	}
	return nil
}

func (h *ClockOutHandler) Serve(ctx context.Context, req *ClockOutRequest) (ClockOutResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ClockOutResponse{}, errorModels.PermissionDeniedServerError()
	}

	updated, err := h.Service.ClockOut(ctx, device.User.UserID, req.EmployeeID)
	if err != nil {
		if errors.Is(err, domain.ErrNotClockedIn) {
			return ClockOutResponse{}, &errorModels.ServerError{HTTPCode: http.StatusConflict, Code: http.StatusConflict, Message: "Not currently clocked in"}
		}
		logger.Ctx(ctx).Error("failed to clock out", logger.Any("owner_id", device.User.UserID), logger.Any("employee_id", req.EmployeeID), logger.Error(err))
		return ClockOutResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to clock out"}
	}

	return ClockOutResponse{Record: attendanceToProto(*updated)}, nil
}
