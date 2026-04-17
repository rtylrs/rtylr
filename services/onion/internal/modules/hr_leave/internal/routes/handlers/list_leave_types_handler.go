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

func ListLeaveTypesRoutes() []string { return []string{"/hr/leave/types"} }

type ListLeaveTypesHandler struct{ Service *application.Service }

type ListLeaveTypesRequest struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

type ListLeaveTypesResponse struct {
	Types  []*hrv1.LeaveTypePayload `json:"types"`
	Total  int64                    `json:"total"`
	Limit  int32                    `json:"limit"`
	Offset int32                    `json:"offset"`
}

func (h *ListLeaveTypesHandler) Serve(ctx context.Context, req *ListLeaveTypesRequest) (ListLeaveTypesResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListLeaveTypesResponse{}, errorModels.PermissionDeniedServerError()
	}

	types, total, err := h.Service.ListLeaveTypes(ctx, device.User.UserID, domain.ListLeaveTypesParams{
		Limit: req.Limit, Offset: req.Offset,
	})
	if err != nil {
		if err == domain.ErrLeaveNotConfigured {
			return ListLeaveTypesResponse{Types: []*hrv1.LeaveTypePayload{}, Total: 0, Limit: req.Limit, Offset: req.Offset}, nil
		}
		logger.Ctx(ctx).Error("failed to list leave types", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return ListLeaveTypesResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to list leave types"}
	}

	payloads := make([]*hrv1.LeaveTypePayload, 0, len(types))
	for _, t := range types {
		payloads = append(payloads, leaveTypeToProto(t))
	}

	return ListLeaveTypesResponse{Types: payloads, Total: total, Limit: req.Limit, Offset: req.Offset}, nil
}
