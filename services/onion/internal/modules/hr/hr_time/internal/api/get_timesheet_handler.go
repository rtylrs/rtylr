package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/hr/hr_time/internal/application"
	"github.com/voxire/onion/internal/modules/hr/hr_time/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hrtimeentryv1 "github.com/voxire/voxie/gen/go/hr/time_entry/v1"
)

func GetTimesheetRoutes() []string { return []string{"/hr/timesheets/current"} }

type GetTimesheetHandler struct{ Service *application.Service }

type GetTimesheetRequest struct {
	EmployeeID  int64  `json:"employee_id"`
	PeriodStart string `json:"period_start"`
	PeriodEnd   string `json:"period_end"`
}

type GetTimesheetResponse struct {
	Timesheet *hrtimeentryv1.TimesheetPayload `json:"timesheet"`
}

func (h *GetTimesheetHandler) Validate(_ context.Context, req *GetTimesheetRequest) error {
	if req == nil || req.EmployeeID <= 0 {
		return onekiterrors.New(http.StatusBadRequest, "employee_id must be greater than 0")
	}
	if _, err := time.Parse("2006-01-02", req.PeriodStart); err != nil {
		return onekiterrors.New(http.StatusBadRequest, "period_start must be in YYYY-MM-DD format")
	}
	if _, err := time.Parse("2006-01-02", req.PeriodEnd); err != nil {
		return onekiterrors.New(http.StatusBadRequest, "period_end must be in YYYY-MM-DD format")
	}
	return nil
}

func (h *GetTimesheetHandler) Serve(ctx context.Context, req *GetTimesheetRequest) (GetTimesheetResponse, error) {
	if err := h.Validate(ctx, req); err != nil {
		return GetTimesheetResponse{}, err
	}

	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return GetTimesheetResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetTimesheetResponse{}, onekiterrors.PermissionDenied()
	}

	ts, err := h.Service.GetOrCreateTimesheetForCaller(ctx, businessID, device.User.UserID, req.EmployeeID, req.PeriodStart, req.PeriodEnd)
	if err != nil {
		if errors.Is(err, domain.ErrTimesheetAccessDenied) {
			return GetTimesheetResponse{}, onekiterrors.PermissionDenied()
		}
		logger.Ctx(ctx).Error("failed to get current timesheet", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return GetTimesheetResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get timesheet")
	}

	return GetTimesheetResponse{Timesheet: timesheetToProto(*ts)}, nil
}
