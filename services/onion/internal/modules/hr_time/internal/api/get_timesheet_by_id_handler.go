package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr_time/internal/application"
	"github.com/voxire/onion/internal/modules/hr_time/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hrtimeentryv1 "github.com/voxire/voxie/gen/go/hr/time_entry/v1"
)

func GetTimesheetByIDRoutes() []string { return []string{"/hr/timesheets/{id}"} }

type GetTimesheetByIDHandler struct{ Service *application.Service }

type GetTimesheetByIDRequest struct{}

type GetTimesheetByIDResponse struct {
	Timesheet *hrtimeentryv1.TimesheetPayload `json:"timesheet"`
}

func (h *GetTimesheetByIDHandler) Serve(ctx context.Context, _ *GetTimesheetByIDRequest) (GetTimesheetByIDResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return GetTimesheetByIDResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetTimesheetByIDResponse{}, onekiterrors.PermissionDenied()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return GetTimesheetByIDResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid timesheet id")
	}

	ts, err := h.Service.GetTimesheetForCaller(ctx, businessID, device.User.UserID, id)
	if err != nil {
		if errors.Is(err, domain.ErrTimesheetAccessDenied) {
			return GetTimesheetByIDResponse{}, onekiterrors.PermissionDenied()
		}
		if errors.Is(err, domain.ErrTimesheetNotFound) {
			return GetTimesheetByIDResponse{}, onekiterrors.New(http.StatusNotFound, "Timesheet not found")
		}
		logger.Ctx(ctx).Error("failed to get timesheet by id", logger.Any("owner_id", device.User.UserID), logger.Any("timesheet_id", id), logger.Error(err))
		return GetTimesheetByIDResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get timesheet")
	}

	return GetTimesheetByIDResponse{Timesheet: timesheetToProto(*ts)}, nil
}
