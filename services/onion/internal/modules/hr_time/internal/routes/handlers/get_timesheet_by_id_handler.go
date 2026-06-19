package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr_time/internal/application"
	"github.com/voxire/onion/internal/modules/hr_time/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
	"github.com/voxire/onion/internal/platform/logger"
	hrtimeentryv1 "github.com/voxire/voxie/gen/go/hr/time_entry/v1"
)

func GetTimesheetByIDRoutes() []string { return []string{"/hr/timesheets/{id}"} }

type GetTimesheetByIDHandler struct{ Service *application.Service }

type GetTimesheetByIDRequest struct{}

type GetTimesheetByIDResponse struct {
	Timesheet *hrtimeentryv1.TimesheetPayload `json:"timesheet"`
}

func (h *GetTimesheetByIDHandler) Serve(ctx context.Context, _ *GetTimesheetByIDRequest) (GetTimesheetByIDResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return GetTimesheetByIDResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetTimesheetByIDResponse{}, errorModels.PermissionDeniedServerError()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return GetTimesheetByIDResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid timesheet id"}
	}

	ts, err := h.Service.GetTimesheet(ctx, businessID, id)
	if err != nil {
		if errors.Is(err, domain.ErrTimesheetNotFound) {
			return GetTimesheetByIDResponse{}, &errorModels.ServerError{HTTPCode: http.StatusNotFound, Code: http.StatusNotFound, Message: "Timesheet not found"}
		}
		logger.Ctx(ctx).Error("failed to get timesheet by id", logger.Any("owner_id", device.User.UserID), logger.Any("timesheet_id", id), logger.Error(err))
		return GetTimesheetByIDResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to get timesheet"}
	}

	return GetTimesheetByIDResponse{Timesheet: timesheetToProto(*ts)}, nil
}
