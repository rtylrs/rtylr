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
	"github.com/voxire/onion/internal/platform/tenant"
)

func SetClockInPINRoutes() []string { return []string{"/hr/employees/me/pin"} }

type SetClockInPINHandler struct {
	Service          *application.Service
	EmployeeResolver EmployeeResolver
}

type SetClockInPINRequest struct {
	// EmployeeID is retained for compatibility with the legacy client. The
	// handler always resolves the employee from the authenticated user.
	EmployeeID int64  `json:"-"`
	Pin        string `json:"pin"`
}

type SetClockInPINResponse struct {
	Success bool `json:"success"`
}

func (h *SetClockInPINHandler) Validate(_ context.Context, req *SetClockInPINRequest) error {
	if req == nil {
		return onekiterrors.New(http.StatusBadRequest, "request body is required")
	}
	if len(req.Pin) < 4 || len(req.Pin) > 6 {
		return onekiterrors.New(http.StatusBadRequest, "pin must be 4-6 digits")
	}
	for _, c := range req.Pin {
		if c < '0' || c > '9' {
			return onekiterrors.New(http.StatusBadRequest, "pin must contain only digits")
		}
	}
	return nil
}

func (h *SetClockInPINHandler) Serve(ctx context.Context, req *SetClockInPINRequest) (SetClockInPINResponse, error) {
	if h == nil || h.Service == nil || h.EmployeeResolver == nil {
		return SetClockInPINResponse{}, onekiterrors.PermissionDenied()
	}
	if err := h.Validate(ctx, req); err != nil {
		return SetClockInPINResponse{}, err
	}

	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return SetClockInPINResponse{}, onekiterrors.PermissionDenied()
	}

	// Resolve both the tenant and employee from the authenticated identity. The
	// legacy body still contains employee_id, but trusting it would let a
	// caller change another employee's clock-in credential.
	businessID, err := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if err != nil {
		return SetClockInPINResponse{}, onekiterrors.PermissionDenied()
	}
	employeeID, err := h.EmployeeResolver.GetEmployeeIDByUserID(ctx, businessID, device.User.UserID)
	if err != nil || employeeID <= 0 {
		return SetClockInPINResponse{}, onekiterrors.PermissionDenied()
	}

	err = h.Service.SetClockInPIN(ctx, businessID, employeeID, req.Pin)
	if err != nil {
		if errors.Is(err, domain.ErrAttendanceNotFound) {
			return SetClockInPINResponse{}, onekiterrors.New(http.StatusNotFound, "Employee not found")
		}
		logger.Ctx(ctx).Error("failed to set clock-in PIN", logger.Any("business_id", businessID), logger.Any("employee_id", employeeID), logger.Error(err))
		return SetClockInPINResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to set PIN")
	}

	return SetClockInPINResponse{Success: true}, nil
}
