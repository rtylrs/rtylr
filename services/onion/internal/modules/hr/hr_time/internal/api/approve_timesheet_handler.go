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

func ApproveTimesheetRoutes() []string { return []string{"/hr/timesheets/{id}/approve"} }

type ApproveTimesheetHandler struct{ Service *application.Service }

type ApproveTimesheetRequest struct{}

type ApproveTimesheetResponse struct {
	Timesheet *hrtimeentryv1.TimesheetPayload `json:"timesheet"`
}

func (h *ApproveTimesheetHandler) Validate(_ context.Context, _ *ApproveTimesheetRequest) error {
	return nil
}

func (h *ApproveTimesheetHandler) Serve(ctx context.Context, _ *ApproveTimesheetRequest) (ApproveTimesheetResponse, error) {
	id, parseErr := strconv.ParseInt(serverHttp.PathParam(ctx, "id"), 10, 64)
	if parseErr != nil {
		return ApproveTimesheetResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid timesheet id")
	}
	return h.approve(ctx, id)
}

// approve acts on the timesheet with the given id. The hand-registered route reads
// it from the path and the hr.time_entry.v1 contract from the request.
func (h *ApproveTimesheetHandler) approve(ctx context.Context, id int64) (ApproveTimesheetResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ApproveTimesheetResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ApproveTimesheetResponse{}, onekiterrors.PermissionDenied()
	}

	ts, err := h.Service.ApproveTimesheet(ctx, businessID, id, device.User.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrTimesheetAccessDenied) || errors.Is(err, domain.ErrTimesheetApproverNotAllowed) || errors.Is(err, domain.ErrTimesheetSelfApproval) {
			return ApproveTimesheetResponse{}, onekiterrors.PermissionDenied()
		}
		if errors.Is(err, domain.ErrTimesheetNotFound) {
			return ApproveTimesheetResponse{}, onekiterrors.New(http.StatusNotFound, "Timesheet not found")
		}
		if errors.Is(err, domain.ErrTimesheetNotPending) {
			return ApproveTimesheetResponse{}, onekiterrors.New(http.StatusConflict, "Timesheet is not awaiting approval")
		}
		if errors.Is(err, domain.ErrApprovalChainNotConfigured) {
			return ApproveTimesheetResponse{}, onekiterrors.New(http.StatusPreconditionFailed, "Approval chain not configured for timesheets")
		}
		logger.Ctx(ctx).Error("failed to approve timesheet", logger.Any("owner_id", device.User.UserID), logger.Any("timesheet_id", id), logger.Error(err))
		return ApproveTimesheetResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to approve timesheet")
	}

	return ApproveTimesheetResponse{Timesheet: timesheetToProto(*ts)}, nil
}
