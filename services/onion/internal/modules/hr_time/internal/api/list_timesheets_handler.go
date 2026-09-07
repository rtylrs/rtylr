package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_time/internal/application"
	"github.com/voxire/onion/internal/modules/hr_time/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hrtimeentryv1 "github.com/voxire/voxie/gen/go/hr/time_entry/v1"
)

func ListTimesheetsRoutes() []string { return []string{"/hr/timesheets"} }

type ListTimesheetsHandler struct{ Service *application.Service }

type ListTimesheetsRequest struct {
	EmployeeID int64 `json:"employee_id"`
	Limit      int32 `json:"limit"`
	Offset     int32 `json:"offset"`
}

type ListTimesheetsResponse struct {
	Timesheets []*hrtimeentryv1.TimesheetPayload `json:"timesheets"`
	Total      int64                             `json:"total"`
	Limit      int32                             `json:"limit"`
	Offset     int32                             `json:"offset"`
}

func (h *ListTimesheetsHandler) Serve(ctx context.Context, req *ListTimesheetsRequest) (ListTimesheetsResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListTimesheetsResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListTimesheetsResponse{}, onekiterrors.PermissionDenied()
	}

	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	timesheets, total, err := h.Service.ListTimesheets(ctx, businessID, domain.ListTimesheetsParams{
		EmployeeID: req.EmployeeID,
		Limit:      req.Limit,
		Offset:     req.Offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list timesheets", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return ListTimesheetsResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list timesheets")
	}

	payloads := make([]*hrtimeentryv1.TimesheetPayload, 0, len(timesheets))
	for _, ts := range timesheets {
		payloads = append(payloads, timesheetToProto(ts))
	}

	return ListTimesheetsResponse{Timesheets: payloads, Total: total, Limit: req.Limit, Offset: req.Offset}, nil
}
