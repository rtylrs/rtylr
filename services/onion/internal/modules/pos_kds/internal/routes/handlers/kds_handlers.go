package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/pos_kds/internal/application"
	"github.com/voxire/onion/internal/modules/pos_kds/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// --- List Active Orders for Kitchen ---

func ListKDSOrdersRoutes() []string { return []string{"/pos/kds", "/pos/kds/orders"} }

type (
	ListKDSOrdersHandler struct{ Service *application.Service }
	ListKDSOrdersRequest struct {
		Station string `json:"station" schema:"station"`
	}
)

type ListKDSOrdersResponse struct {
	Orders []domain.KDSOrder `json:"orders"`
}

func (h *ListKDSOrdersHandler) Serve(ctx context.Context, req *ListKDSOrdersRequest) (ListKDSOrdersResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListKDSOrdersResponse{}, errorModels.PermissionDeniedServerError()
	}
	ownerID := tenant.ResolveOwnerID(ctx, device.User.BusinessID, device.User.UserID)
	orders, err := h.Service.ListActiveOrdersByStation(ctx, ownerID, req.Station)
	if err != nil {
		return ListKDSOrdersResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list KDS orders"}
	}
	return ListKDSOrdersResponse{Orders: orders}, nil
}

// --- Update Item Status by item id only ---

func UpdateKDSItemByIDRoutes() []string { return []string{"/pos/kds/items/{itemId}"} }

type UpdateKDSItemByIDHandler struct{ Service *application.Service }

type UpdateKDSItemByIDRequest struct {
	Status string `json:"status"`
}

type UpdateKDSItemByIDResponse struct {
	Message string `json:"message"`
}

func (h *UpdateKDSItemByIDHandler) Validate(_ context.Context, req *UpdateKDSItemByIDRequest) *errorModels.ServerError {
	req.Status = strings.TrimSpace(req.Status)
	if req.Status == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Status is required"}
	}
	return nil
}

func (h *UpdateKDSItemByIDHandler) Serve(ctx context.Context, req *UpdateKDSItemByIDRequest) (UpdateKDSItemByIDResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateKDSItemByIDResponse{}, errorModels.PermissionDeniedServerError()
	}
	ownerID := tenant.ResolveOwnerID(ctx, device.User.BusinessID, device.User.UserID)

	itemIDStr := chi.URLParamFromCtx(ctx, "itemId")
	itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil || itemID <= 0 {
		return UpdateKDSItemByIDResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid item ID"}
	}

	if updateErr := h.Service.UpdateItemStatusByItemID(ctx, ownerID, itemID, req.Status); updateErr != nil {
		return UpdateKDSItemByIDResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to update item status"}
	}
	return UpdateKDSItemByIDResponse{Message: "Item status updated"}, nil
}

// --- Update Item Status (mark ready/done) ---

func UpdateKDSItemRoutes() []string { return []string{"/pos/kds/{orderId}/items/{itemId}"} }

type UpdateKDSItemHandler struct{ Service *application.Service }

type UpdateKDSItemRequest struct {
	Status string `json:"status"`
}

type UpdateKDSItemResponse struct {
	Message string `json:"message"`
}

func (h *UpdateKDSItemHandler) Validate(_ context.Context, req *UpdateKDSItemRequest) *errorModels.ServerError {
	req.Status = strings.TrimSpace(req.Status)
	if req.Status == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Status is required"}
	}
	return nil
}

func (h *UpdateKDSItemHandler) Serve(ctx context.Context, req *UpdateKDSItemRequest) (UpdateKDSItemResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateKDSItemResponse{}, errorModels.PermissionDeniedServerError()
	}
	ownerID := tenant.ResolveOwnerID(ctx, device.User.BusinessID, device.User.UserID)

	orderIDStr := chi.URLParamFromCtx(ctx, "orderId")
	orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
	if err != nil || orderID <= 0 {
		return UpdateKDSItemResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid order ID"}
	}

	itemIDStr := chi.URLParamFromCtx(ctx, "itemId")
	itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil || itemID <= 0 {
		return UpdateKDSItemResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid item ID"}
	}

	if updateErr := h.Service.UpdateItemStatus(ctx, domain.UpdateItemStatusInput{
		OwnerID: ownerID,
		OrderID: orderID,
		ItemID:  itemID,
		Status:  req.Status,
	}); updateErr != nil {
		return UpdateKDSItemResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to update item status"}
	}
	return UpdateKDSItemResponse{Message: "Item status updated"}, nil
}
