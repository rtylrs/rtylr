package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr/hr_attendance/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	hrattendancev1 "github.com/voxire/voxie/gen/go/hr/attendance/v1"
)

type clockOutService interface {
	ClockOut(ctx context.Context, businessID, employeeID int64) (*domain.AttendanceRecordRecord, error)
}

type ClockOutHandler struct {
	Service clockOutService
	Resolve func(ctx context.Context) (businessID, employeeID int64, err error)
}

type ClockOutRequest struct {
	EmployeeID int64 `json:"employee_id"`
}

type ClockOutResponse struct {
	Record *hrattendancev1.AttendanceRecordPayload `json:"record"`
}

func (h *ClockOutHandler) Validate(_ context.Context, req *ClockOutRequest) error {
	if req == nil || req.EmployeeID <= 0 {
		return onekiterrors.New(http.StatusBadRequest, "employee_id is required")
	}
	return nil
}

func (h *ClockOutHandler) Serve(ctx context.Context, req *ClockOutRequest) (ClockOutResponse, error) {
	if h.Resolve == nil || h.Service == nil {
		return ClockOutResponse{}, onekiterrors.PermissionDenied()
	}
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ClockOutResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, employeeID, err := h.Resolve(ctx)
	if err != nil || businessID <= 0 || employeeID <= 0 {
		return ClockOutResponse{}, onekiterrors.PermissionDenied()
	}
	if req == nil || req.EmployeeID != employeeID {
		return ClockOutResponse{}, onekiterrors.PermissionDenied()
	}

	updated, err := h.Service.ClockOut(ctx, businessID, employeeID)
	if err != nil {
		if errors.Is(err, domain.ErrNotClockedIn) {
			return ClockOutResponse{}, onekiterrors.New(http.StatusConflict, "Not currently clocked in")
		}
		logger.Ctx(ctx).Error("failed to clock out", logger.Any("owner_id", businessID), logger.Any("employee_id", employeeID), logger.Error(err))
		return ClockOutResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to clock out")
	}

	return ClockOutResponse{Record: attendanceToProto(*updated)}, nil
}
