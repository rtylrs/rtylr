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

func ApproveTimesheetRoutes() []string { return []string{"/hr/timesheets/{id}/approve"} }

type ApproveTimesheetHandler struct{ Service *application.Service }

type ApproveTimesheetRequest struct{}

type ApproveTimesheetResponse struct {
	Timesheet *hrtimeentryv1.TimesheetPayload `json:"timesheet"`
}

func (h *ApproveTimesheetHandler) Validate(_ context.Context, _ *ApproveTimesheetRequest) *errorModels.ServerError {
	return nil
}

func (h *ApproveTimesheetHandler) Serve(ctx context.Context, _ *ApproveTimesheetRequest) (ApproveTimesheetResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ApproveTimesheetResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ApproveTimesheetResponse{}, errorModels.PermissionDeniedServerError()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return ApproveTimesheetResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid timesheet id"}
	}

	ts, err := h.Service.ApproveTimesheet(ctx, businessID, id, device.User.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrTimesheetNotFound) {
			return ApproveTimesheetResponse{}, &errorModels.ServerError{HTTPCode: http.StatusNotFound, Code: http.StatusNotFound, Message: "Timesheet not found"}
		}
		if errors.Is(err, domain.ErrApprovalChainNotConfigured) {
			return ApproveTimesheetResponse{}, &errorModels.ServerError{HTTPCode: http.StatusPreconditionFailed, Code: http.StatusPreconditionFailed, Message: "Approval chain not configured for timesheets"}
		}
		logger.Ctx(ctx).Error("failed to approve timesheet", logger.Any("owner_id", device.User.UserID), logger.Any("timesheet_id", id), logger.Error(err))
		return ApproveTimesheetResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to approve timesheet"}
	}

	return ApproveTimesheetResponse{Timesheet: timesheetToProto(*ts)}, nil
}
