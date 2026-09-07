package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_time/internal/application"
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

func (h *GetTimesheetHandler) Serve(ctx context.Context, req *GetTimesheetRequest) (GetTimesheetResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return GetTimesheetResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetTimesheetResponse{}, onekiterrors.PermissionDenied()
	}

	ts, err := h.Service.GetOrCreateTimesheet(ctx, businessID, req.EmployeeID, req.PeriodStart, req.PeriodEnd)
	if err != nil {
		logger.Ctx(ctx).Error("failed to get current timesheet", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return GetTimesheetResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get timesheet")
	}

	return GetTimesheetResponse{Timesheet: timesheetToProto(*ts)}, nil
}
