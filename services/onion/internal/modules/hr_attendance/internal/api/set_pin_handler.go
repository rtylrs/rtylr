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

type SetClockInPINHandler struct{ Service *application.Service }

type SetClockInPINRequest struct {
	EmployeeID int64  `json:"employee_id"`
	Pin        string `json:"pin"`
}

type SetClockInPINResponse struct {
	Success bool `json:"success"`
}

func (h *SetClockInPINHandler) Validate(_ context.Context, req *SetClockInPINRequest) error {
	if req.EmployeeID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "employee_id is required")
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
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return SetClockInPINResponse{}, onekiterrors.PermissionDenied()
	}

	// Resolve the tenant scope from the JWT. Passing the raw user id as the
	// owner argument scoped the UPDATE by a value that is not
	// hr_employees.business_id, so any employee whose user id is not itself a
	// business id matched zero rows (#938). The sibling clock-in/clock-out
	// handlers resolve the same way.
	businessID, err := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if err != nil {
		return SetClockInPINResponse{}, onekiterrors.PermissionDenied()
	}

	err = h.Service.SetClockInPIN(ctx, businessID, req.EmployeeID, req.Pin)
	if err != nil {
		if errors.Is(err, domain.ErrAttendanceNotFound) {
			return SetClockInPINResponse{}, onekiterrors.New(http.StatusNotFound, "Employee not found")
		}
		logger.Ctx(ctx).Error("failed to set clock-in PIN", logger.Any("business_id", businessID), logger.Any("employee_id", req.EmployeeID), logger.Error(err))
		return SetClockInPINResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to set PIN")
	}

	return SetClockInPINResponse{Success: true}, nil
}
