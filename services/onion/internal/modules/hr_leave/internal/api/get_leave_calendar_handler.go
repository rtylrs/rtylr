package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_leave/internal/application"
	"github.com/voxire/onion/internal/modules/hr_leave/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	hrleavev1 "github.com/voxire/voxie/gen/go/hr/leave/v1"
)

func GetLeaveCalendarRoutes() []string { return []string{"/hr/leave/calendar"} }

type GetLeaveCalendarHandler struct{ Service *application.Service }

type GetLeaveCalendarRequest struct {
	LocationID int64  `json:"location_id"`
	DateFrom   string `json:"date_from"`
	DateTo     string `json:"date_to"`
}

type GetLeaveCalendarResponse struct {
	Requests []*hrleavev1.LeaveRequestPayload `json:"requests"`
}

func (h *GetLeaveCalendarHandler) Serve(ctx context.Context, req *GetLeaveCalendarRequest) (GetLeaveCalendarResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return GetLeaveCalendarResponse{}, onekiterrors.PermissionDenied()
	}

	requests, err := h.Service.GetLeaveCalendar(ctx, device.User.UserID, domain.LeaveCalendarParams{
		LocationID: req.LocationID,
		DateFrom:   req.DateFrom,
		DateTo:     req.DateTo,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to get leave calendar", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return GetLeaveCalendarResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get leave calendar")
	}

	payloads := make([]*hrleavev1.LeaveRequestPayload, 0, len(requests))
	for _, r := range requests {
		payloads = append(payloads, leaveRequestToProto(r))
	}

	return GetLeaveCalendarResponse{Requests: payloads}, nil
}
