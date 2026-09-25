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

func (h *CreateOrderHandler) Validate(_ context.Context, req *CreateOrderRequest) error {
	if len(req.Items) == 0 {
		return onekiterrors.New(http.StatusBadRequest, "Order must have at least one item")
	}

	if req.Type != "" && !domain.IsOrderType(domain.OrderType(req.Type)) {
		return onekiterrors.New(http.StatusBadRequest, "Unsupported order type")
	}

	for _, item := range req.Items {
		if item.Quantity <= 0 {
			return onekiterrors.New(http.StatusBadRequest, "Item quantity must be positive")
		}

		if item.MenuItemID <= 0 && item.ComboID <= 0 {
			return onekiterrors.New(http.StatusBadRequest, "Each item must reference a menu item or combo")
		}
		if item.MenuItemID > 0 && item.ComboID > 0 {
			return onekiterrors.New(http.StatusBadRequest, "Each item cannot reference both a menu item and combo")
		}
		for _, modifierID := range item.ModifierIDs {
			if modifierID <= 0 {
				return onekiterrors.New(http.StatusBadRequest, "Modifier ids must be positive")
			}
		}
	}

	return nil
}

func (h *CreateOrderHandler) Serve(ctx context.Context, req *CreateOrderRequest) (CreateOrderResponse, error) {
	user, authErr := requirePOSPermission(ctx, "pos:orders:create")
	if authErr != nil {
		return CreateOrderResponse{}, authErr
	}

	businessID, bizErr := resolveBusinessID(ctx, user)
	if bizErr != nil {
		return CreateOrderResponse{}, bizErr
	}
	staffID := user.UserID

	items := make([]domain.OrderItemInput, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, domain.OrderItemInput{
			MenuItemID:  item.MenuItemID,
			ComboID:     item.ComboID,
			ModifierIDs: append([]int64(nil), item.ModifierIDs...),
			Name:        item.Name,
			SKU:         item.SKU,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			DiscountPct: item.DiscountPct,
			TaxRate:     item.TaxRate,
			Notes:       item.Notes,
		})
	}

	orderID, err := h.Service.CreateOrder(ctx, businessID, businessID, staffID, domain.OrderInput{
		SessionID:     req.SessionID,
		TableID:       req.TableID,
		LocationID:    req.LocationID,
		Type:          domain.OrderType(req.Type),
		CustomerID:    req.CustomerID,
		CustomerName:  req.CustomerName,
		PaymentMethod: domain.PaymentMethod(req.PaymentMethod),
		Notes:         req.Notes,
		KitchenNotes:  req.KitchenNotes,
		Items:         items,
	})
	if err != nil {
		if errors.Is(err, domain.ErrPricedItemNotFound) {
			return CreateOrderResponse{}, onekiterrors.New(http.StatusBadRequest, "Item price could not be resolved")
		}
		if errors.Is(err, domain.ErrDiscountNotAllowed) {
			return CreateOrderResponse{}, onekiterrors.New(http.StatusForbidden, "You cannot apply discounts")
		}
		if errors.Is(err, domain.ErrDiscountExceedsLimit) {
			return CreateOrderResponse{}, onekiterrors.New(http.StatusBadRequest, "Discount exceeds your allowed limit")
		}
		return CreateOrderResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to create order")
	}

	order, getErr := h.Service.GetOrder(ctx, businessID, orderID)
	if getErr != nil {
		return CreateOrderResponse{OrderID: orderID, Message: "Order created"}, nil
	}

	return CreateOrderResponse{OrderID: orderID, OrderNumber: order.OrderNumber, Message: "Order created"}, nil
}

// ─── Get Order Detail ─────────────────────────────────────────────────────────

func GetOrderDetailRoutes() []string { return []string{"/pos/orders/detail", "/pos/orders/{id}"} }

type GetOrderDetailHandler struct{ Service *application.Service }

type GetOrderDetailRequest struct {
	OrderID int64 `json:"order_id"`
}

type GetOrderDetailResponse struct {
	Order *OrderPayload `json:"order"`
}

func (h *GetOrderDetailHandler) Serve(ctx context.Context, req *GetOrderDetailRequest) (GetOrderDetailResponse, error) {
	user, authErr := requirePOSPermission(ctx, "pos:orders:view")
	if authErr != nil {
		return GetOrderDetailResponse{}, authErr
	}

	businessID, bizErr := resolveBusinessID(ctx, user)
	if bizErr != nil {
		return GetOrderDetailResponse{}, bizErr
	}

	orderID := req.OrderID
	if raw := chi.URLParamFromCtx(ctx, "id"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || parsed <= 0 {
			return GetOrderDetailResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid order id")
		}
		orderID = parsed
	}

	order, err := h.Service.GetOrder(ctx, businessID, orderID)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			return GetOrderDetailResponse{}, onekiterrors.New(http.StatusNotFound, "Order not found")
		}

		return GetOrderDetailResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get order")
	}

	payload := newOrderPayload(*order)

	return GetOrderDetailResponse{Order: &payload}, nil
}

// ─── Update Order Status ──────────────────────────────────────────────────────

func UpdateOrderStatusRoutes() []string {
	return []string{"/pos/orders/status", "/pos/orders/{id}/status"}
}

type UpdateOrderStatusHandler struct{ Service *application.Service }

type UpdateOrderStatusRequest struct {
	OrderID int64  `json:"order_id"`
	Status  string `json:"status"`
}

type UpdateOrderStatusResponse struct {
	Message string `json:"message"`
}

func (h *UpdateOrderStatusHandler) Validate(ctx context.Context, req *UpdateOrderStatusRequest) error {
	if req.OrderID == 0 && chi.URLParamFromCtx(ctx, "id") == "" {
		return onekiterrors.New(http.StatusBadRequest, "Order id is required")
	}

	if req.Status == "" {
		return onekiterrors.New(http.StatusBadRequest, "Status is required")
	}

	return nil
}

func (h *UpdateOrderStatusHandler) Serve(ctx context.Context, req *UpdateOrderStatusRequest) (UpdateOrderStatusResponse, error) {
	user, authErr := requirePOSPermission(ctx, "pos:orders:create")
	if authErr != nil {
		return UpdateOrderStatusResponse{}, authErr
	}

	businessID, bizErr := resolveBusinessID(ctx, user)
	if bizErr != nil {
		return UpdateOrderStatusResponse{}, bizErr
	}

	orderID := req.OrderID
	if raw := chi.URLParamFromCtx(ctx, "id"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || parsed <= 0 {
			return UpdateOrderStatusResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid order id")
		}
		orderID = parsed
	}

	if err := h.Service.UpdateOrderStatus(ctx, businessID, orderID, domain.OrderStatus(req.Status)); err != nil {
		switch {
		case errors.Is(err, domain.ErrOrderNotFound):
			return UpdateOrderStatusResponse{}, onekiterrors.New(http.StatusNotFound, "Order not found")
		case errors.Is(err, domain.ErrInvalidOrderStatus):
			return UpdateOrderStatusResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid order status")
		case errors.Is(err, domain.ErrOrderStatusRequiresWorkflow):
			return UpdateOrderStatusResponse{}, onekiterrors.New(http.StatusConflict, "Order status must be changed through its dedicated flow")
		default:
			return UpdateOrderStatusResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update order status")
		}
	}

	return UpdateOrderStatusResponse{Message: "Order status updated"}, nil
}

// ─── Update Order Item Status ─────────────────────────────────────────────────
