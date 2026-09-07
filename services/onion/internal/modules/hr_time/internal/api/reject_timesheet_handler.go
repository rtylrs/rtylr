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

func RejectTimesheetRoutes() []string { return []string{"/hr/timesheets/{id}/reject"} }

type RejectTimesheetHandler struct{ Service *application.Service }

type RejectTimesheetRequest struct {
	RejectionReason string `json:"rejection_reason"`
}

type RejectTimesheetResponse struct {
	Timesheet *hrtimeentryv1.TimesheetPayload `json:"timesheet"`
}

func (h *RejectTimesheetHandler) Validate(_ context.Context, _ *RejectTimesheetRequest) error {
	return nil
}

func (h *RejectTimesheetHandler) Serve(ctx context.Context, req *RejectTimesheetRequest) (RejectTimesheetResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return RejectTimesheetResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return RejectTimesheetResponse{}, onekiterrors.PermissionDenied()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return RejectTimesheetResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid timesheet id")
	}

	ts, err := h.Service.RejectTimesheet(ctx, businessID, id, req.RejectionReason)
	if err != nil {
		if errors.Is(err, domain.ErrTimesheetNotFound) {
			return RejectTimesheetResponse{}, onekiterrors.New(http.StatusNotFound, "Timesheet not found")
		}
		logger.Ctx(ctx).Error("failed to reject timesheet", logger.Any("owner_id", device.User.UserID), logger.Any("timesheet_id", id), logger.Error(err))
		return RejectTimesheetResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to reject timesheet")
	}

	return RejectTimesheetResponse{Timesheet: timesheetToProto(*ts)}, nil
}
