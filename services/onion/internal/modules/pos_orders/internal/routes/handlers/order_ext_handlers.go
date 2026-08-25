package handlers

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

// resolveOrderID returns the order id from either the chi URL param or the request body.
// The URL param takes precedence when present.
func resolveOrderID(ctx context.Context, bodyID int64) (int64, error) {
	if raw := chi.URLParamFromCtx(ctx, "id"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || parsed <= 0 {
			return 0, onekiterrors.New(http.StatusBadRequest, "Invalid order id")
		}
		return parsed, nil
	}
	if bodyID == 0 {
		return 0, onekiterrors.New(http.StatusBadRequest, "Order id is required")
	}
	return bodyID, nil
}

// ---- Split Order ----

func SplitOrderRoutes() []string { return []string{"/pos/orders/split", "/pos/orders/{id}/split"} }

type SplitOrderHandler struct{ Service *application.Service }

type SplitOrderRequest struct {
	OrderID int64     `json:"order_id"`
	Splits  [][]int64 `json:"splits"`
}

type SplitOrderResponse struct {
	OrderIDs []int64 `json:"order_ids"`
	Message  string  `json:"message"`
}

func (h *SplitOrderHandler) Validate(ctx context.Context, req *SplitOrderRequest) error {
	if req.OrderID == 0 && chi.URLParamFromCtx(ctx, "id") == "" {
		return onekiterrors.New(http.StatusBadRequest, "Order ID is required")
	}
	if len(req.Splits) < 2 {
		return onekiterrors.New(http.StatusBadRequest, "At least two splits are required")
	}
	return nil
}

func (h *SplitOrderHandler) Serve(ctx context.Context, req *SplitOrderRequest) (SplitOrderResponse, error) {
	user, authErr := authenticatedPOSUser(ctx)
	if authErr != nil {
		return SplitOrderResponse{}, authErr
	}
	businessID, bizErr := resolveBusinessID(ctx, user)
	if bizErr != nil {
		return SplitOrderResponse{}, bizErr
	}

	orderID, sErr := resolveOrderID(ctx, req.OrderID)
	if sErr != nil {
		return SplitOrderResponse{}, sErr
	}

	orderIDs, err := h.Service.SplitOrder(ctx, businessID, orderID, req.Splits)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			return SplitOrderResponse{}, onekiterrors.New(http.StatusNotFound, "Order not found")
		}
		return SplitOrderResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to split order")
	}
	return SplitOrderResponse{OrderIDs: orderIDs, Message: "Order split successfully"}, nil
}

// ---- Apply Discount ----

func DiscountOrderRoutes() []string {
	return []string{"/pos/orders/discount", "/pos/orders/{id}/discount"}
}

type DiscountOrderHandler struct{ Service *application.Service }

type DiscountOrderRequest struct {
	OrderID       int64   `json:"order_id"`
	DiscountType  string  `json:"discount_type"`
	DiscountValue float64 `json:"discount_value"`
	Reason        string  `json:"reason"`
}

type DiscountOrderResponse struct {
	Message string `json:"message"`
}

func (h *DiscountOrderHandler) Validate(ctx context.Context, req *DiscountOrderRequest) error {
	if req.OrderID == 0 && chi.URLParamFromCtx(ctx, "id") == "" {
		return onekiterrors.New(http.StatusBadRequest, "Order ID is required")
	}
	if req.DiscountValue <= 0 {
		return onekiterrors.New(http.StatusBadRequest, "Discount value must be positive")
	}
	return nil
}

func (h *DiscountOrderHandler) Serve(ctx context.Context, req *DiscountOrderRequest) (DiscountOrderResponse, error) {
	user, authErr := authenticatedPOSUser(ctx)
	if authErr != nil {
		return DiscountOrderResponse{}, authErr
	}
	businessID, bizErr := resolveBusinessID(ctx, user)
	if bizErr != nil {
		return DiscountOrderResponse{}, bizErr
	}

	orderID, sErr := resolveOrderID(ctx, req.OrderID)
	if sErr != nil {
		return DiscountOrderResponse{}, sErr
	}

	if err := h.Service.ApplyDiscount(ctx, businessID, user.UserID, domain.DiscountInput{
		// Capabilities are resolved server-side from the authenticated POS user.
		// Client-provided discount values never bypass that check.
		OrderID:       orderID,
		DiscountType:  req.DiscountType,
		DiscountValue: req.DiscountValue,
		Reason:        req.Reason,
	}); err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			return DiscountOrderResponse{}, onekiterrors.New(http.StatusNotFound, "Order not found")
		}
		if errors.Is(err, domain.ErrOrderCannotBeDiscounted) {
			return DiscountOrderResponse{}, onekiterrors.New(http.StatusConflict, "Paid or refunded orders cannot be discounted")
		}
		if errors.Is(err, domain.ErrDiscountNotAllowed) {
			return DiscountOrderResponse{}, onekiterrors.New(http.StatusForbidden, "You cannot apply discounts")
		}
		if errors.Is(err, domain.ErrDiscountExceedsLimit) {
			return DiscountOrderResponse{}, onekiterrors.New(http.StatusBadRequest, "Discount exceeds your allowed limit")
		}
		return DiscountOrderResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to apply discount")
	}
	return DiscountOrderResponse{Message: "Discount applied successfully"}, nil
}

// ---- Void Order ----

func VoidOrderRoutes() []string { return []string{"/pos/orders/void", "/pos/orders/{id}/void"} }

type VoidOrderHandler struct{ Service *application.Service }

type VoidOrderRequest struct {
	OrderID int64  `json:"order_id"`
	Reason  string `json:"reason"`
}

type VoidOrderResponse struct {
	Message string `json:"message"`
}

func (h *VoidOrderHandler) Validate(ctx context.Context, req *VoidOrderRequest) error {
	if req.OrderID == 0 && chi.URLParamFromCtx(ctx, "id") == "" {
		return onekiterrors.New(http.StatusBadRequest, "Order ID is required")
	}
	if req.Reason == "" {
		return onekiterrors.New(http.StatusBadRequest, "Reason is required")
	}
	return nil
}

func (h *VoidOrderHandler) Serve(ctx context.Context, req *VoidOrderRequest) (VoidOrderResponse, error) {
	user, authErr := authenticatedPOSUser(ctx)
	if authErr != nil {
		return VoidOrderResponse{}, authErr
	}
	businessID, bizErr := resolveBusinessID(ctx, user)
	if bizErr != nil {
		return VoidOrderResponse{}, bizErr
	}

	orderID, sErr := resolveOrderID(ctx, req.OrderID)
	if sErr != nil {
		return VoidOrderResponse{}, sErr
	}

	if err := h.Service.VoidOrder(ctx, businessID, orderID, req.Reason); err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			return VoidOrderResponse{}, onekiterrors.New(http.StatusNotFound, "Order not found")
		}
		return VoidOrderResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to void order")
	}
	return VoidOrderResponse{Message: "Order voided successfully"}, nil
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
	user, authErr := authenticatedPOSUser(ctx)
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
	user, authErr := authenticatedPOSUser(ctx)
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
	user, authErr := authenticatedPOSUser(ctx)
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
		return VoidOrderItemResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to void order item")
	}
	return VoidOrderItemResponse{Success: true}, nil
}

// ---- Transfer Table ----

func TransferTableRoutes() []string {
	return []string{"/pos/orders/transfer-table", "/pos/orders/{id}/transfer-table"}
}

type TransferTableHandler struct{ Service *application.Service }

type TransferTableRequest struct {
	OrderID    int64 `json:"order_id"`
	NewTableID int64 `json:"new_table_id"`
}

type TransferTableResponse struct {
	Success bool `json:"success"`
}

func (h *TransferTableHandler) Validate(ctx context.Context, req *TransferTableRequest) error {
	if req.OrderID == 0 && chi.URLParamFromCtx(ctx, "id") == "" {
		return onekiterrors.New(http.StatusBadRequest, "Order ID is required")
	}
	if req.NewTableID <= 0 {
		return onekiterrors.New(http.StatusBadRequest, "New table ID is required")
	}
	return nil
}

func (h *TransferTableHandler) Serve(ctx context.Context, req *TransferTableRequest) (TransferTableResponse, error) {
	user, authErr := authenticatedPOSUser(ctx)
	if authErr != nil {
		return TransferTableResponse{}, authErr
	}
	businessID, bizErr := resolveBusinessID(ctx, user)
	if bizErr != nil {
		return TransferTableResponse{}, bizErr
	}

	orderID, sErr := resolveOrderID(ctx, req.OrderID)
	if sErr != nil {
		return TransferTableResponse{}, sErr
	}

	if err := h.Service.TransferTable(ctx, businessID, orderID, req.NewTableID); err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			return TransferTableResponse{}, onekiterrors.New(http.StatusNotFound, "Order not found")
		}
		if errors.Is(err, domain.ErrOrderCannotBeTransferred) {
			return TransferTableResponse{}, onekiterrors.New(http.StatusUnprocessableEntity, err.Error())
		}
		return TransferTableResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to transfer table")
	}
	return TransferTableResponse{Success: true}, nil
}

// ---- List Held Orders ----

func ListHeldOrdersRoutes() []string { return []string{"/pos/orders/held"} }

type ListHeldOrdersHandler struct{ Service *application.Service }

type ListHeldOrdersRequest struct{}

type ListHeldOrdersResponse struct {
	Orders []OrderPayload `json:"orders"`
}

func (h *ListHeldOrdersHandler) Serve(ctx context.Context, _ *ListHeldOrdersRequest) (ListHeldOrdersResponse, error) {
	user, authErr := authenticatedPOSUser(ctx)
	if authErr != nil {
		return ListHeldOrdersResponse{}, authErr
	}
	businessID, bizErr := resolveBusinessID(ctx, user)
	if bizErr != nil {
		return ListHeldOrdersResponse{}, bizErr
	}

	orders, err := h.Service.ListHeldOrders(ctx, businessID)
	if err != nil {
		return ListHeldOrdersResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list held orders")
	}

	payloads := make([]OrderPayload, 0, len(orders))
	for _, o := range orders {
		payloads = append(payloads, newOrderPayload(o))
	}
	return ListHeldOrdersResponse{Orders: payloads}, nil
}
