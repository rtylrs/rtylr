package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_attendance/internal/application"
	"github.com/voxire/onion/internal/modules/hr_attendance/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	hrattendancev1 "github.com/voxire/voxie/gen/go/hr/attendance/v1"
)

func ClockOutRoutes() []string { return []string{"/hr/attendance/clock-out"} }

type ClockOutHandler struct{ Service *application.Service }

type ClockOutRequest struct {
	EmployeeID int64 `json:"employee_id"`
}

type ClockOutResponse struct {
	Record *hrattendancev1.AttendanceRecordPayload `json:"record"`
}

func (h *ClockOutHandler) Validate(_ context.Context, req *ClockOutRequest) error {
	if req.EmployeeID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "employee_id is required")
	}
	return nil
}

func (h *ClockOutHandler) Serve(ctx context.Context, req *ClockOutRequest) (ClockOutResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ClockOutResponse{}, onekiterrors.PermissionDenied()
	}

	updated, err := h.Service.ClockOut(ctx, device.User.UserID, req.EmployeeID)
	if err != nil {
		if errors.Is(err, domain.ErrNotClockedIn) {
			return ClockOutResponse{}, onekiterrors.New(http.StatusConflict, "Not currently clocked in")
		}
		logger.Ctx(ctx).Error("failed to clock out", logger.Any("owner_id", device.User.UserID), logger.Any("employee_id", req.EmployeeID), logger.Error(err))
		return ClockOutResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to clock out")
	}

	return ClockOutResponse{Record: attendanceToProto(*updated)}, nil
}
