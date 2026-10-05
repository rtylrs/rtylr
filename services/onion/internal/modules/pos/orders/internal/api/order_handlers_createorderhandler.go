package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/pos/orders/internal/application"
	"github.com/voxire/onion/internal/modules/pos/orders/internal/domain"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

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

type GetOrderDetailHandler struct{ Service *application.Service }

type GetOrderDetailRequest struct {
	OrderID int64 `json:"order_id"`
}

type GetOrderDetailResponse struct {
	Order *OrderPayload `json:"order"`
}

// ─── Update Order Status ──────────────────────────────────────────────────────

type UpdateOrderStatusHandler struct{ Service *application.Service }

type UpdateOrderStatusRequest struct {
	OrderID int64  `json:"order_id"`
	Status  string `json:"status"`
}

type UpdateOrderStatusResponse struct {
	Message string `json:"message"`
}

// ─── Update Order Item Status ─────────────────────────────────────────────────
