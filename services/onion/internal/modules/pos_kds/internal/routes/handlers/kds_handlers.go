package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"

	"github.com/voxire/onion/internal/modules/pos_kds/internal/application"
	"github.com/voxire/onion/internal/modules/pos_kds/internal/domain"
)

// --- List Active Orders for Kitchen ---

func ListKDSOrdersRoutes() []string { return []string{"/pos/kds"} }

type ListKDSOrdersHandler struct{ Service *application.Service }
type ListKDSOrdersRequest struct{}
type ListKDSOrdersResponse struct {
	Orders []domain.KDSOrder `json:"orders"`
}

func (h *ListKDSOrdersHandler) Serve(ctx context.Context, _ *ListKDSOrdersRequest) (ListKDSOrdersResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListKDSOrdersResponse{}, errorModels.PermissionDeniedServerError()
	}
	orders, err := h.Service.ListActiveOrders(ctx, device.User.UserID)
	if err != nil {
		return ListKDSOrdersResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list KDS orders"}
	}
	return ListKDSOrdersResponse{Orders: orders}, nil
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

	orderIDStr := chi.URLParamFromCtx(ctx, "orderId")
	orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
	if err != nil {
		return UpdateKDSItemResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid order ID"}
	}

	itemIDStr := chi.URLParamFromCtx(ctx, "itemId")
	itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		return UpdateKDSItemResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid item ID"}
	}

	if updateErr := h.Service.UpdateItemStatus(ctx, domain.UpdateItemStatusInput{
		OrderID: orderID,
		ItemID:  itemID,
		Status:  req.Status,
	}); updateErr != nil {
		return UpdateKDSItemResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to update item status"}
	}
	return UpdateKDSItemResponse{Message: "Item status updated"}, nil
}
