package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	posv1 "github.com/voxire/voxie/gen/go/pos/v1"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/application"
	"github.com/voxire/onion/internal/modules/pos_orders/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

// ─── Shared payloads ──────────────────────────────────────────────────────────

type OrderPayload = posv1.Order

func orderItemToProto(item domain.OrderItem) *posv1.OrderItem {
	return &posv1.OrderItem{
		Id:         item.ID,
		OrderId:    item.OrderID,
		MenuItemId: item.MenuItemID,
		Name:       item.Name,
		Quantity:   int32(item.Quantity),
		UnitPrice:  item.UnitPrice,
		TotalPrice: item.Total,
		Notes:      item.Notes,
		Status:     string(item.Status),
		CreatedAt:  item.CreatedAt.Format(time.RFC3339),
	}
}

func newOrderPayload(o domain.Order) *posv1.Order {
	p := &posv1.Order{
		Id:             o.ID,
		OrderNumber:    o.OrderNumber,
		LocationId:     o.LocationID,
		SessionId:      o.SessionID,
		TableId:        o.TableID,
		CustomerId:     o.CustomerID,
		OrderType:      string(o.Type),
		Status:         string(o.Status),
		Subtotal:       o.Subtotal,
		TaxAmount:      o.TaxAmount,
		DiscountAmount: o.DiscountAmount,
		Total:          o.Total,
		Notes:          o.Notes,
		CreatedBy:      fmt.Sprintf("%d", o.StaffID),
		CreatedAt:      o.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      o.UpdatedAt.Format(time.RFC3339),
	}

	if len(o.Items) > 0 {
		items := make([]*posv1.OrderItem, 0, len(o.Items))
		for _, item := range o.Items {
			items = append(items, orderItemToProto(item))
		}
		p.Items = items
	}

	return p
}

// ─── List Orders ──────────────────────────────────────────────────────────────

func ListOrdersRoutes() []string { return []string{"/pos/orders"} }

type ListOrdersHandler struct{ Service *application.Service }

type ListOrdersRequest struct {
	SessionID int64  `json:"session_id"`
	Status    string `json:"status"`
	DateFrom  string `json:"date_from"`
	DateTo    string `json:"date_to"`
}

type ListOrdersResponse struct {
	Orders []*posv1.Order `json:"orders"`
}

func (h *ListOrdersHandler) Serve(ctx context.Context, req *ListOrdersRequest) (ListOrdersResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListOrdersResponse{}, errorModels.PermissionDeniedServerError()
	}

	ownerID := device.User.UserID

	orders, err := h.Service.ListOrders(ctx, ownerID, req.SessionID, req.Status, req.DateFrom, req.DateTo)
	if err != nil {
		return ListOrdersResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to list orders",
		}
	}

	payloads := make([]*posv1.Order, 0, len(orders))
	for _, o := range orders {
		payloads = append(payloads, newOrderPayload(o))
	}

	return ListOrdersResponse{Orders: payloads}, nil
}

// ─── Create Order ─────────────────────────────────────────────────────────────

type CreateOrderHandler struct{ Service *application.Service }

type CreateOrderItemInput struct {
	MenuItemID  int64   `json:"menu_item_id"`
	Name        string  `json:"name"`
	SKU         string  `json:"sku"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	DiscountPct float64 `json:"discount_pct"`
	TaxRate     float64 `json:"tax_rate"`
	Notes       string  `json:"notes"`
}

type CreateOrderRequest struct {
	SessionID     int64                  `json:"session_id"`
	TableID       int64                  `json:"table_id"`
	LocationID    int64                  `json:"location_id"`
	Type          string                 `json:"type"`
	CustomerID    int64                  `json:"customer_id"`
	CustomerName  string                 `json:"customer_name"`
	PaymentMethod string                 `json:"payment_method"`
	Notes         string                 `json:"notes"`
	KitchenNotes  string                 `json:"kitchen_notes"`
	Items         []CreateOrderItemInput `json:"items"`
}

type CreateOrderResponse struct {
	OrderID     int64  `json:"order_id"`
	OrderNumber string `json:"order_number"`
	Message     string `json:"message"`
}

func (h *CreateOrderHandler) Validate(_ context.Context, req *CreateOrderRequest) *errorModels.ServerError {
	if len(req.Items) == 0 {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Order must have at least one item",
		}
	}

	for _, item := range req.Items {
		if item.Name == "" {
			return &errorModels.ServerError{
				HTTPCode: http.StatusBadRequest,
				Code:     http.StatusBadRequest,
				Message:  "Each item must have a name",
			}
		}

		if item.Quantity <= 0 {
			return &errorModels.ServerError{
				HTTPCode: http.StatusBadRequest,
				Code:     http.StatusBadRequest,
				Message:  "Item quantity must be positive",
			}
		}

		if item.UnitPrice < 0 {
			return &errorModels.ServerError{
				HTTPCode: http.StatusBadRequest,
				Code:     http.StatusBadRequest,
				Message:  "Item unit price cannot be negative",
			}
		}
	}

	return nil
}

func (h *CreateOrderHandler) Serve(ctx context.Context, req *CreateOrderRequest) (CreateOrderResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateOrderResponse{}, errorModels.PermissionDeniedServerError()
	}

	ownerID := device.User.UserID
	staffID := device.User.UserID

	items := make([]domain.OrderItemInput, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, domain.OrderItemInput{
			MenuItemID:  item.MenuItemID,
			Name:        item.Name,
			SKU:         item.SKU,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			DiscountPct: item.DiscountPct,
			TaxRate:     item.TaxRate,
			Notes:       item.Notes,
		})
	}

	orderID, err := h.Service.CreateOrder(ctx, ownerID, staffID, domain.OrderInput{
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
		return CreateOrderResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to create order",
		}
	}

	order, getErr := h.Service.GetOrder(ctx, ownerID, orderID)
	if getErr != nil {
		return CreateOrderResponse{OrderID: orderID, Message: "Order created"}, nil
	}

	return CreateOrderResponse{OrderID: orderID, OrderNumber: order.OrderNumber, Message: "Order created"}, nil
}

// ─── Get Order Detail ─────────────────────────────────────────────────────────

func GetOrderDetailRoutes() []string { return []string{"/pos/orders/detail"} }

type GetOrderDetailHandler struct{ Service *application.Service }

type GetOrderDetailRequest struct {
	OrderID int64 `json:"order_id"`
}

type GetOrderDetailResponse struct {
	Order *posv1.Order `json:"order"`
}

func (h *GetOrderDetailHandler) Serve(ctx context.Context, req *GetOrderDetailRequest) (GetOrderDetailResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetOrderDetailResponse{}, errorModels.PermissionDeniedServerError()
	}

	ownerID := device.User.UserID

	order, err := h.Service.GetOrder(ctx, ownerID, req.OrderID)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			return GetOrderDetailResponse{}, &errorModels.ServerError{
				HTTPCode: http.StatusNotFound,
				Code:     http.StatusNotFound,
				Message:  "Order not found",
			}
		}

		return GetOrderDetailResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to get order",
		}
	}

	payload := newOrderPayload(*order)

	return GetOrderDetailResponse{Order: payload}, nil
}

// ─── Update Order Status ──────────────────────────────────────────────────────

func UpdateOrderStatusRoutes() []string { return []string{"/pos/orders/status"} }

type UpdateOrderStatusHandler struct{ Service *application.Service }

type UpdateOrderStatusRequest struct {
	OrderID int64  `json:"order_id"`
	Status  string `json:"status"`
}

type UpdateOrderStatusResponse struct {
	Message string `json:"message"`
}

func (h *UpdateOrderStatusHandler) Validate(_ context.Context, req *UpdateOrderStatusRequest) *errorModels.ServerError {
	if req.OrderID == 0 {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Order id is required",
		}
	}

	if req.Status == "" {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Status is required",
		}
	}

	return nil
}

func (h *UpdateOrderStatusHandler) Serve(ctx context.Context, req *UpdateOrderStatusRequest) (UpdateOrderStatusResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateOrderStatusResponse{}, errorModels.PermissionDeniedServerError()
	}

	ownerID := device.User.UserID

	if err := h.Service.UpdateOrderStatus(ctx, ownerID, req.OrderID, domain.OrderStatus(req.Status)); err != nil {
		return UpdateOrderStatusResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to update order status",
		}
	}

	return UpdateOrderStatusResponse{Message: "Order status updated"}, nil
}

// ─── Update Order Item Status ─────────────────────────────────────────────────

func UpdateOrderItemStatusRoutes() []string { return []string{"/pos/orders/item-status"} }

type UpdateOrderItemStatusHandler struct{ Service *application.Service }

type UpdateOrderItemStatusRequest struct {
	OrderID int64  `json:"order_id"`
	ItemID  int64  `json:"item_id"`
	Status  string `json:"status"`
}

type UpdateOrderItemStatusResponse struct {
	Message string `json:"message"`
}

func (h *UpdateOrderItemStatusHandler) Validate(_ context.Context, req *UpdateOrderItemStatusRequest) *errorModels.ServerError {
	if req.OrderID == 0 || req.ItemID == 0 {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Order id and item id are required",
		}
	}

	if req.Status == "" {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Status is required",
		}
	}

	return nil
}

func (h *UpdateOrderItemStatusHandler) Serve(ctx context.Context, req *UpdateOrderItemStatusRequest) (UpdateOrderItemStatusResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateOrderItemStatusResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.Service.UpdateOrderItemStatus(ctx, req.OrderID, req.ItemID, domain.ItemStatus(req.Status)); err != nil {
		return UpdateOrderItemStatusResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to update item status",
		}
	}

	return UpdateOrderItemStatusResponse{Message: "Item status updated"}, nil
}

// ─── Process Payment ──────────────────────────────────────────────────────────

func ProcessPaymentRoutes() []string { return []string{"/pos/orders/payment"} }

type ProcessPaymentHandler struct{ Service *application.Service }

type ProcessPaymentRequest struct {
	OrderID       int64   `json:"order_id"`
	PaymentMethod string  `json:"payment_method"`
	AmountPaid    float64 `json:"amount_paid"`
}

type ProcessPaymentResponse struct {
	Message string `json:"message"`
}

func (h *ProcessPaymentHandler) Validate(_ context.Context, req *ProcessPaymentRequest) *errorModels.ServerError {
	if req.OrderID == 0 {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Order id is required",
		}
	}

	if req.AmountPaid < 0 {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Amount paid cannot be negative",
		}
	}

	return nil
}

func (h *ProcessPaymentHandler) Serve(ctx context.Context, req *ProcessPaymentRequest) (ProcessPaymentResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ProcessPaymentResponse{}, errorModels.PermissionDeniedServerError()
	}

	ownerID := device.User.UserID

	if err := h.Service.ProcessPayment(ctx, ownerID, domain.PaymentInput{
		OrderID:       req.OrderID,
		PaymentMethod: domain.PaymentMethod(req.PaymentMethod),
		AmountPaid:    req.AmountPaid,
	}); err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			return ProcessPaymentResponse{}, &errorModels.ServerError{
				HTTPCode: http.StatusNotFound,
				Code:     http.StatusNotFound,
				Message:  "Order not found",
			}
		}

		return ProcessPaymentResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to process payment",
		}
	}

	return ProcessPaymentResponse{Message: "Payment processed"}, nil
}

// ─── Process Refund ───────────────────────────────────────────────────────────

func ProcessRefundRoutes() []string { return []string{"/pos/orders/refund"} }

type ProcessRefundHandler struct{ Service *application.Service }

type ProcessRefundRequest struct {
	OrderID int64   `json:"order_id"`
	Amount  float64 `json:"amount"`
	Reason  string  `json:"reason"`
}

type ProcessRefundResponse struct {
	RefundID int64  `json:"refund_id"`
	Message  string `json:"message"`
}

func (h *ProcessRefundHandler) Validate(_ context.Context, req *ProcessRefundRequest) *errorModels.ServerError {
	if req.OrderID == 0 {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Order id is required",
		}
	}

	if req.Amount <= 0 {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Refund amount must be positive",
		}
	}

	return nil
}

func (h *ProcessRefundHandler) Serve(ctx context.Context, req *ProcessRefundRequest) (ProcessRefundResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ProcessRefundResponse{}, errorModels.PermissionDeniedServerError()
	}

	ownerID := device.User.UserID
	staffID := device.User.UserID

	refundID, err := h.Service.ProcessRefund(ctx, ownerID, staffID, domain.RefundInput{
		OrderID: req.OrderID,
		Amount:  req.Amount,
		Reason:  req.Reason,
	})
	if err != nil {
		return ProcessRefundResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to process refund",
		}
	}

	return ProcessRefundResponse{RefundID: refundID, Message: "Refund processed"}, nil
}
