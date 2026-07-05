package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/hr_equipment/internal/application"
	"github.com/voxire/onion/internal/modules/hr_equipment/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hrequipmentv1 "github.com/voxire/voxie/gen/go/hr/equipment/v1"
)

func CreateEquipmentRequestRoutes() []string { return []string{"/hr/equipment/requests"} }

type CreateEquipmentRequestHandler struct{ Service *application.Service }

type CreateEquipmentRequestRequest struct {
	EmployeeID   int64  `json:"employee_id"`
	ItemName     string `json:"item_name"`
	Quantity     int32  `json:"quantity"`
	Reason       string `json:"reason"`
	NeededByDate string `json:"needed_by_date"`
}

type CreateEquipmentRequestResponse struct {
	Request *hrequipmentv1.EquipmentRequestPayload `json:"request"`
}

func (h *CreateEquipmentRequestHandler) Validate(_ context.Context, req *CreateEquipmentRequestRequest) error {
	if req.EmployeeID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "employee_id is required")
	}
	req.ItemName = strings.TrimSpace(req.ItemName)
	if req.ItemName == "" {
		return onekiterrors.New(http.StatusBadRequest, "item_name is required")
	}
	if req.Quantity <= 0 {
		return onekiterrors.New(http.StatusBadRequest, "quantity must be greater than 0")
	}
	return nil
}

func (h *CreateEquipmentRequestHandler) Serve(ctx context.Context, req *CreateEquipmentRequestRequest) (CreateEquipmentRequestResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CreateEquipmentRequestResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateEquipmentRequestResponse{}, onekiterrors.PermissionDenied()
	}

	created, err := h.Service.CreateEquipmentRequest(ctx, businessID, domain.CreateEquipmentRequestParams{
		EmployeeID:   req.EmployeeID,
		ItemName:     req.ItemName,
		Quantity:     req.Quantity,
		Reason:       req.Reason,
		NeededByDate: req.NeededByDate,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to create equipment request", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return CreateEquipmentRequestResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to create equipment request")
	}

	return CreateEquipmentRequestResponse{Request: equipmentRequestToProto(*created)}, nil
}
