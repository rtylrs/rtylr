package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_equipment/internal/application"
	"github.com/voxire/onion/internal/modules/hr_equipment/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hrequipmentv1 "github.com/voxire/voxie/gen/go/hr/equipment/v1"
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
	Requests []*hrequipmentv1.EquipmentRequestPayload `json:"requests"`
	Total    int64                                    `json:"total"`
	Limit    int32                                    `json:"limit"`
	Offset   int32                                    `json:"offset"`
}

func (h *ListEquipmentRequestsHandler) Serve(ctx context.Context, req *ListEquipmentRequestsRequest) (ListEquipmentRequestsResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListEquipmentRequestsResponse{}, onekiterrors.PermissionDenied()
	}
	if req == nil {
		req = &ListEquipmentRequestsRequest{}
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListEquipmentRequestsResponse{}, onekiterrors.PermissionDenied()
	}

	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	requests, total, err := h.Service.ListEquipmentRequests(ctx, businessID, domain.ListEquipmentRequestsParams{
		EmployeeID: req.EmployeeID,
		Status:     req.Status,
		Limit:      req.Limit,
		Offset:     req.Offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list equipment requests", logger.Any("business_id", businessID), logger.Error(err))
		return ListEquipmentRequestsResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list equipment requests")
	}

	payloads := make([]*hrequipmentv1.EquipmentRequestPayload, 0, len(requests))
	for _, r := range requests {
		payloads = append(payloads, equipmentRequestToProto(r))
	}

	return ListEquipmentRequestsResponse{Requests: payloads, Total: total, Limit: req.Limit, Offset: req.Offset}, nil
}
