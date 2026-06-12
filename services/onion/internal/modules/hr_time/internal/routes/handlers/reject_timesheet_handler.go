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
	"github.com/voxire/onion/internal/platform/logger"
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

func (h *RejectTimesheetHandler) Validate(_ context.Context, _ *RejectTimesheetRequest) *errorModels.ServerError {
	return nil
}

func (h *RejectTimesheetHandler) Serve(ctx context.Context, req *RejectTimesheetRequest) (RejectTimesheetResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return RejectTimesheetResponse{}, errorModels.PermissionDeniedServerError()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return RejectTimesheetResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid timesheet id"}
	}

	ts, err := h.Service.RejectTimesheet(ctx, device.User.UserID, id, req.RejectionReason)
	if err != nil {
		if errors.Is(err, domain.ErrTimesheetNotFound) {
			return RejectTimesheetResponse{}, &errorModels.ServerError{HTTPCode: http.StatusNotFound, Code: http.StatusNotFound, Message: "Timesheet not found"}
		}
		logger.Ctx(ctx).Error("failed to reject timesheet", logger.Any("owner_id", device.User.UserID), logger.Any("timesheet_id", id), logger.Error(err))
		return RejectTimesheetResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to reject timesheet"}
	}

	return RejectTimesheetResponse{Timesheet: timesheetToProto(*ts)}, nil
}
