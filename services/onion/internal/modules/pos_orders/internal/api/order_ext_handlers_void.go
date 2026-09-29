package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/application"
	"github.com/voxire/onion/internal/modules/pos_orders/internal/domain"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

type VoidOrderApprovalResponse struct {
	Message string `json:"message"`
}

func (r *VoidOrderApprovalRequest) orderID(ctx context.Context) (int64, error) {
	return resolveOrderID(ctx, r.OrderID)
}

func serveVoidOrderWithPermission(ctx context.Context, req *VoidOrderApprovalRequest, service *application.Service, permission string) (VoidOrderApprovalResponse, error) {
	user, authErr := requirePOSPermission(ctx, permission)
	if authErr != nil {
		return VoidOrderApprovalResponse{}, authErr
	}
	businessID, bizErr := resolveBusinessID(ctx, user)
	if bizErr != nil {
		return VoidOrderApprovalResponse{}, bizErr
	}
	orderID, orderErr := req.orderID(ctx)
	if orderErr != nil {
		return VoidOrderApprovalResponse{}, orderErr
	}
	if err := service.VoidOrder(ctx, businessID, orderID, req.Reason); err != nil {
		return VoidOrderApprovalResponse{}, voidOrderError(err)
	}
	return VoidOrderApprovalResponse{Message: "Order voided successfully"}, nil
}

type ApproveVoidOrderHandler struct{ Service *application.Service }

func ApproveVoidOrderRoutes() []string {
	return []string{"/pos/orders/{id}/void/approve"}
}

func (h *ApproveVoidOrderHandler) Validate(ctx context.Context, req *VoidOrderApprovalRequest) error {
	_, err := req.orderID(ctx)
	return err
}

func (h *ApproveVoidOrderHandler) Serve(ctx context.Context, req *VoidOrderApprovalRequest) (VoidOrderApprovalResponse, error) {
	return serveVoidOrderWithPermission(ctx, req, h.Service, "pos:orders:void:approve")
}

type OverrideVoidOrderHandler struct{ Service *application.Service }

func OverrideVoidOrderRoutes() []string {
	return []string{"/pos/orders/{id}/void/override"}
}

func (h *OverrideVoidOrderHandler) Validate(ctx context.Context, req *VoidOrderApprovalRequest) error {
	_, err := req.orderID(ctx)
	return err
}

func (h *OverrideVoidOrderHandler) Serve(ctx context.Context, req *VoidOrderApprovalRequest) (VoidOrderApprovalResponse, error) {
	return serveVoidOrderWithPermission(ctx, req, h.Service, "pos:orders:void:override")
}

// ---- Hold Order ----

func HoldOrderRoutes() []string { return []string{"/pos/orders/hold", "/pos/orders/{id}/hold"} }

type HoldOrderHandler struct{ Service *application.Service }

type HoldOrderRequest struct {
	OrderID int64  `json:"order_id"`
	Reason  string `json:"reason"`
}

type HoldOrderResponse struct {
	Message string `json:"message"`
}

func (h *HoldOrderHandler) Validate(ctx context.Context, req *HoldOrderRequest) error {
	if req.OrderID == 0 && chi.URLParamFromCtx(ctx, "id") == "" {
		return onekiterrors.New(http.StatusBadRequest, "Order ID is required")
	}
	return nil
}

func (h *HoldOrderHandler) Serve(ctx context.Context, req *HoldOrderRequest) (HoldOrderResponse, error) {
	user, authErr := requirePOSPermission(ctx, "pos:orders:create")
	if authErr != nil {
		return HoldOrderResponse{}, authErr
	}
	businessID, bizErr := resolveBusinessID(ctx, user)
	if bizErr != nil {
		return HoldOrderResponse{}, bizErr
	}

	orderID, sErr := resolveOrderID(ctx, req.OrderID)
	if sErr != nil {
		return HoldOrderResponse{}, sErr
	}

	if err := h.Service.HoldOrder(ctx, businessID, orderID, req.Reason); err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			return HoldOrderResponse{}, onekiterrors.New(http.StatusNotFound, "Order not found")
		}
		return HoldOrderResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to hold order")
	}
	return HoldOrderResponse{Message: "Order held successfully"}, nil
}

// ---- Send to Kitchen ----

func SendToKitchenRoutes() []string {
	return []string{"/pos/orders/send-to-kitchen", "/pos/orders/{id}/send-to-kitchen"}
}

type SendToKitchenHandler struct{ Service *application.Service }

type SendToKitchenRequest struct {
	OrderID int64 `json:"order_id"`
}

type SendToKitchenResponse struct {
	Success bool `json:"success"`
}

func (h *SendToKitchenHandler) Validate(ctx context.Context, req *SendToKitchenRequest) error {
	if req.OrderID == 0 && chi.URLParamFromCtx(ctx, "id") == "" {
		return onekiterrors.New(http.StatusBadRequest, "Order ID is required")
	}
	return nil
}

func (h *SendToKitchenHandler) Serve(ctx context.Context, req *SendToKitchenRequest) (SendToKitchenResponse, error) {
	user, authErr := requirePOSPermission(ctx, "pos:orders:create")
	if authErr != nil {
		return SendToKitchenResponse{}, authErr
	}
	businessID, bizErr := resolveBusinessID(ctx, user)
	if bizErr != nil {
		return SendToKitchenResponse{}, bizErr
	}

	orderID, sErr := resolveOrderID(ctx, req.OrderID)
	if sErr != nil {
		return SendToKitchenResponse{}, sErr
	}

	if err := h.Service.SendToKitchen(ctx, businessID, orderID); err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			return SendToKitchenResponse{}, onekiterrors.New(http.StatusNotFound, "Order not found")
		}
		return SendToKitchenResponse{}, onekiterrors.New(http.StatusUnprocessableEntity, err.Error())
	}
	return SendToKitchenResponse{Success: true}, nil
}

// ---- Void Order Item ----

func VoidOrderItemRoutes() []string {
	return []string{"/pos/orders/{id}/items/{itemId}"}
}

type VoidOrderItemHandler struct{ Service *application.Service }

type VoidOrderItemRequest struct{}

type VoidOrderItemResponse struct {
	Success bool `json:"success"`
}

func (h *VoidOrderItemHandler) Validate(ctx context.Context, _ *VoidOrderItemRequest) error {
	if chi.URLParamFromCtx(ctx, "id") == "" {
		return onekiterrors.New(http.StatusBadRequest, "Order ID is required")
	}
	if chi.URLParamFromCtx(ctx, "itemId") == "" {
		return onekiterrors.New(http.StatusBadRequest, "Item ID is required")
	}
	return nil
}

func (h *VoidOrderItemHandler) Serve(ctx context.Context, _ *VoidOrderItemRequest) (VoidOrderItemResponse, error) {
	user, authErr := requirePOSPermission(ctx, "pos:orders:void")
	if authErr != nil {
		return VoidOrderItemResponse{}, authErr
	}
	businessID, bizErr := resolveBusinessID(ctx, user)
	if bizErr != nil {
		return VoidOrderItemResponse{}, bizErr
	}

	orderID, sErr := resolveOrderID(ctx, 0)
	if sErr != nil {
		return VoidOrderItemResponse{}, sErr
	}

	rawItemID := chi.URLParamFromCtx(ctx, "itemId")
	itemID, err := strconv.ParseInt(rawItemID, 10, 64)
	if err != nil || itemID <= 0 {
		return VoidOrderItemResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid item id")
	}

	if err := h.Service.VoidOrderItem(ctx, businessID, orderID, itemID); err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			return VoidOrderItemResponse{}, onekiterrors.New(http.StatusNotFound, "Order item not found")
		}
		if errors.Is(err, domain.ErrItemNotVoidable) {
			return VoidOrderItemResponse{}, onekiterrors.New(http.StatusConflict, "Order item cannot be voided in its current status")
		}
		return VoidOrderItemResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to void order item")
	}
	return VoidOrderItemResponse{Success: true}, nil
}

// ---- Transfer Table ----
