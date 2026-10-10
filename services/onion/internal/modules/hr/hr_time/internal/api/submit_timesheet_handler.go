package api

import (
	"context"
	"errors"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
	"net/http"
	"strconv"

	"github.com/voxire/onion/internal/modules/hr/hr_time/internal/application"
	"github.com/voxire/onion/internal/modules/hr/hr_time/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hrtimeentryv1 "github.com/voxire/voxie/gen/go/hr/time_entry/v1"
)

func SubmitTimesheetRoutes() []string { return []string{"/hr/timesheets/{id}/submit"} }

type SubmitTimesheetHandler struct{ Service *application.Service }

type SubmitTimesheetRequest struct{}

type SubmitTimesheetResponse struct {
	Timesheet *hrtimeentryv1.TimesheetPayload `json:"timesheet"`
}

func (h *SubmitTimesheetHandler) Validate(_ context.Context, _ *SubmitTimesheetRequest) error {
	return nil
}

func (h *SubmitTimesheetHandler) Serve(ctx context.Context, _ *SubmitTimesheetRequest) (SubmitTimesheetResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return SubmitTimesheetResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return SubmitTimesheetResponse{}, onekiterrors.PermissionDenied()
	}

	idStr := serverHttp.PathParam(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return SubmitTimesheetResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid timesheet id")
	}

	ts, err := h.Service.SubmitTimesheet(ctx, businessID, id)
	if err != nil {
		if errors.Is(err, domain.ErrTimesheetAlreadySubmitted) {
			return SubmitTimesheetResponse{}, onekiterrors.New(http.StatusConflict, "Timesheet already submitted")
		}
		if errors.Is(err, domain.ErrTimesheetNotFound) {
			return SubmitTimesheetResponse{}, onekiterrors.New(http.StatusNotFound, "Timesheet not found")
		}
		logger.Ctx(ctx).Error("failed to submit timesheet", logger.Any("owner_id", device.User.UserID), logger.Any("timesheet_id", id), logger.Error(err))
		return SubmitTimesheetResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to submit timesheet")
	}

	return SubmitTimesheetResponse{Timesheet: timesheetToProto(*ts)}, nil
}
