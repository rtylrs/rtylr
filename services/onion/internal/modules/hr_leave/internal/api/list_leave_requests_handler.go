package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_leave/internal/application"
	"github.com/voxire/onion/internal/modules/hr_leave/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	hrleavev1 "github.com/voxire/voxie/gen/go/hr/leave/v1"
)

func ListLeaveRequestsRoutes() []string { return []string{"/hr/leave/requests"} }

type ListLeaveRequestsHandler struct{ Service *application.Service }

type ListLeaveRequestsRequest struct {
	EmployeeID int64  `json:"employee_id"`
	Status     string `json:"status"`
	Limit      int32  `json:"limit"`
	Offset     int32  `json:"offset"`
}

type ListLeaveRequestsResponse struct {
	Requests []*hrleavev1.LeaveRequestPayload `json:"requests"`
	Total    int64                            `json:"total"`
	Limit    int32                            `json:"limit"`
	Offset   int32                            `json:"offset"`
}

func (h *ListLeaveRequestsHandler) Serve(ctx context.Context, req *ListLeaveRequestsRequest) (ListLeaveRequestsResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListLeaveRequestsResponse{}, onekiterrors.PermissionDenied()
	}

	requests, total, err := h.Service.ListLeaveRequests(ctx, device.User.UserID, device.User.UserID, domain.ListLeaveRequestsParams{
		EmployeeID: req.EmployeeID,
		Status:     req.Status,
		Limit:      req.Limit,
		Offset:     req.Offset,
	})
	if err != nil {
		if errors.Is(err, domain.ErrLeaveAccessDenied) {
			return ListLeaveRequestsResponse{}, onekiterrors.PermissionDenied()
		}
		logger.Ctx(ctx).Error("failed to list leave requests", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return ListLeaveRequestsResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list leave requests")
	}

	payloads := make([]*hrleavev1.LeaveRequestPayload, 0, len(requests))
	for _, r := range requests {
		payloads = append(payloads, leaveRequestToProto(r))
	}

	return ListLeaveRequestsResponse{Requests: payloads, Total: total, Limit: req.Limit, Offset: req.Offset}, nil
}
