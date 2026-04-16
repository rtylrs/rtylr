package handlers

import (
	"context"
	"net/http"

	hrv1 "github.com/voxire/voxie/gen/go/hr/v1"

	"github.com/voxire/onion/internal/modules/hr_leave/internal/application"
	"github.com/voxire/onion/internal/modules/hr_leave/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
)

func ListLeaveRequestsRoutes() []string { return []string{"/hr/leave/requests"} }

type ListLeaveRequestsHandler struct{ Service *application.Service }

type ListLeaveRequestsRequest struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

type ListLeaveRequestsResponse struct {
	Requests []*hrv1.LeaveRequestPayload `json:"requests"`
	Total    int64                       `json:"total"`
	Limit    int32                       `json:"limit"`
	Offset   int32                       `json:"offset"`
}

func (h *ListLeaveRequestsHandler) Serve(ctx context.Context, req *ListLeaveRequestsRequest) (ListLeaveRequestsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListLeaveRequestsResponse{}, errorModels.PermissionDeniedServerError()
	}

	requests, total, err := h.Service.ListLeaveRequests(ctx, device.User.UserID, domain.ListLeaveRequestsParams{
		Limit: req.Limit, Offset: req.Offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list leave requests", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return ListLeaveRequestsResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to list leave requests"}
	}

	payloads := make([]*hrv1.LeaveRequestPayload, 0, len(requests))
	for _, r := range requests {
		payloads = append(payloads, leaveRequestToProto(r))
	}

	return ListLeaveRequestsResponse{Requests: payloads, Total: total, Limit: req.Limit, Offset: req.Offset}, nil
}
