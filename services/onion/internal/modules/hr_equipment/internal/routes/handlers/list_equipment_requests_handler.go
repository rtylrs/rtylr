package handlers

import (
	"context"
	"net/http"

	hrv1 "github.com/voxire/voxie/gen/go/hr/v1"
	"github.com/voxire/onion/internal/modules/hr_equipment/internal/application"
	"github.com/voxire/onion/internal/modules/hr_equipment/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
)

func ListEquipmentRequestsRoutes() []string { return []string{"/hr/equipment/requests"} }

type ListEquipmentRequestsHandler struct{ Service *application.Service }

type ListEquipmentRequestsRequest struct {
	EmployeeID int64  `json:"employee_id"`
	Status     string `json:"status"`
	Limit      int32  `json:"limit"`
	Offset     int32  `json:"offset"`
}

type ListEquipmentRequestsResponse struct {
	Requests []*hrv1.EquipmentRequestPayload `json:"requests"`
	Total    int64                           `json:"total"`
	Limit    int32                           `json:"limit"`
	Offset   int32                           `json:"offset"`
}

func (h *ListEquipmentRequestsHandler) Serve(ctx context.Context, req *ListEquipmentRequestsRequest) (ListEquipmentRequestsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListEquipmentRequestsResponse{}, errorModels.PermissionDeniedServerError()
	}

	requests, total, err := h.Service.ListEquipmentRequests(ctx, device.User.UserID, domain.ListEquipmentRequestsParams{
		EmployeeID: req.EmployeeID,
		Status:     req.Status,
		Limit:      req.Limit,
		Offset:     req.Offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list equipment requests", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return ListEquipmentRequestsResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to list equipment requests"}
	}

	payloads := make([]*hrv1.EquipmentRequestPayload, 0, len(requests))
	for _, r := range requests {
		payloads = append(payloads, equipmentRequestToProto(r))
	}

	return ListEquipmentRequestsResponse{Requests: payloads, Total: total, Limit: req.Limit, Offset: req.Offset}, nil
}
