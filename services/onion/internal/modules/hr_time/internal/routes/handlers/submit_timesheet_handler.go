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
	hrv1 "github.com/voxire/voxie/gen/go/hr/v1"
)

func SubmitTimesheetRoutes() []string { return []string{"/hr/timesheets/{id}/submit"} }

type SubmitTimesheetHandler struct{ Service *application.Service }

type SubmitTimesheetRequest struct{}

type SubmitTimesheetResponse struct {
	Timesheet *hrv1.TimesheetPayload `json:"timesheet"`
}

func (h *SubmitTimesheetHandler) Validate(_ context.Context, _ *SubmitTimesheetRequest) *errorModels.ServerError {
	return nil
}

func (h *SubmitTimesheetHandler) Serve(ctx context.Context, _ *SubmitTimesheetRequest) (SubmitTimesheetResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return SubmitTimesheetResponse{}, errorModels.PermissionDeniedServerError()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return SubmitTimesheetResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid timesheet id"}
	}

	ts, err := h.Service.SubmitTimesheet(ctx, device.User.UserID, id)
	if err != nil {
		if errors.Is(err, domain.ErrTimesheetAlreadySubmitted) {
			return SubmitTimesheetResponse{}, &errorModels.ServerError{HTTPCode: http.StatusConflict, Code: http.StatusConflict, Message: "Timesheet already submitted"}
		}
		if errors.Is(err, domain.ErrTimesheetNotFound) {
			return SubmitTimesheetResponse{}, &errorModels.ServerError{HTTPCode: http.StatusNotFound, Code: http.StatusNotFound, Message: "Timesheet not found"}
		}
		logger.Ctx(ctx).Error("failed to submit timesheet", logger.Any("owner_id", device.User.UserID), logger.Any("timesheet_id", id), logger.Error(err))
		return SubmitTimesheetResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to submit timesheet"}
	}

	return SubmitTimesheetResponse{Timesheet: timesheetToProto(*ts)}, nil
}
