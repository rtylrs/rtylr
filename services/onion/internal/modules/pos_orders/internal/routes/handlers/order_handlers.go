package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/application"
	"github.com/voxire/onion/internal/modules/pos_orders/internal/domain"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

// ─── Shared payloads ──────────────────────────────────────────────────────────

type OrderItemPayload struct {
	ID             int64     `json:"id"`
	OrderID        int64     `json:"order_id"`
	MenuItemID     int64     `json:"menu_item_id"`
	Name           string    `json:"name"`
	SKU            string    `json:"sku"`
	Quantity       float64   `json:"quantity"`
	UnitPrice      float64   `json:"unit_price"`
	DiscountPct    float64   `json:"discount_pct"`
	DiscountAmount float64   `json:"discount_amount"`
	TaxRate        float64   `json:"tax_rate"`
	Total          float64   `json:"total"`
	Notes          string    `json:"notes"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

type OrderPayload struct {
	ID             int64              `json:"id"`
	OwnerID        int64              `json:"owner_id"`
	SessionID      int64              `json:"session_id"`
	TableID        int64              `json:"table_id"`
	LocationID     int64              `json:"location_id"`
	OrderNumber    string             `json:"order_number"`
	Type           string             `json:"type"`
	Status         string             `json:"status"`
	CustomerID     int64              `json:"customer_id"`
	CustomerName   string             `json:"customer_name"`
	StaffID        int64              `json:"staff_id"`
	Subtotal       float64            `json:"subtotal"`
	DiscountAmount float64            `json:"discount_amount"`
	TaxAmount      float64            `json:"tax_amount"`
	Total          float64            `json:"total"`
	PaymentMethod  string             `json:"payment_method"`
	PaymentStatus  string             `json:"payment_status"`
	AmountPaid     float64            `json:"amount_paid"`
	ChangeDue      float64            `json:"change_due"`
	Notes          string             `json:"notes"`
	KitchenNotes   string             `json:"kitchen_notes"`
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
	PaidAt         *time.Time         `json:"paid_at"`
	Items          []OrderItemPayload `json:"items,omitempty"`
}

func newOrderPayload(o domain.Order) OrderPayload {
	p := OrderPayload{
		ID:             o.ID,
		OwnerID:        o.OwnerID,
		SessionID:      o.SessionID,
		TableID:        o.TableID,
		LocationID:     o.LocationID,
		OrderNumber:    o.OrderNumber,
		Type:           string(o.Type),
		Status:         string(o.Status),
		CustomerID:     o.CustomerID,
		CustomerName:   o.CustomerName,
		StaffID:        o.StaffID,
		Subtotal:       o.Subtotal,
		DiscountAmount: o.DiscountAmount,
		TaxAmount:      o.TaxAmount,
		Total:          o.Total,
		PaymentMethod:  string(o.PaymentMethod),
		PaymentStatus:  string(o.PaymentStatus),
		AmountPaid:     o.AmountPaid,
		ChangeDue:      o.ChangeDue,
		Notes:          o.Notes,
		KitchenNotes:   o.KitchenNotes,
		CreatedAt:      o.CreatedAt,
		UpdatedAt:      o.UpdatedAt,
		PaidAt:         o.PaidAt,
	}

	if len(o.Items) > 0 {
		items := make([]OrderItemPayload, 0, len(o.Items))
		for _, item := range o.Items {
			items = append(items, OrderItemPayload{
				ID:             item.ID,
				OrderID:        item.OrderID,
				MenuItemID:     item.MenuItemID,
				Name:           item.Name,
				SKU:            item.SKU,
				Quantity:       item.Quantity,
				UnitPrice:      item.UnitPrice,
				DiscountPct:    item.DiscountPct,
				DiscountAmount: item.DiscountAmount,
				TaxRate:        item.TaxRate,
				Total:          item.Total,
				Notes:          item.Notes,
				Status:         string(item.Status),
				CreatedAt:      item.CreatedAt,
			})
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
	Search    string `json:"search"`
	Page      int64  `json:"page"`
	PageSize  int64  `json:"page_size"`
}

type ListOrdersResponse struct {
	Orders   []OrderPayload `json:"orders"`
	Total    int64          `json:"total"`
	HasMore  bool           `json:"has_more"`
	Page     int64          `json:"page"`
	PageSize int64          `json:"page_size"`
}

func (h *ListOrdersHandler) Validate(_ context.Context, req *ListOrdersRequest) error {
	const layout = "2006-01-02"
	if req.DateFrom != "" {
		if _, err := time.Parse(layout, req.DateFrom); err != nil {
			return onekiterrors.New(http.StatusBadRequest, "date_from must be in YYYY-MM-DD format")
		}
	}
	if req.DateTo != "" {
		if _, err := time.Parse(layout, req.DateTo); err != nil {
			return onekiterrors.New(http.StatusBadRequest, "date_to must be in YYYY-MM-DD format")
		}
	}
	return nil
}

func (h *ListOrdersHandler) Serve(ctx context.Context, req *ListOrdersRequest) (ListOrdersResponse, error) {
	user, authErr := requirePOSPermission(ctx, "pos:orders:view")
	if authErr != nil {
		return ListOrdersResponse{}, authErr
	}

	businessID, bizErr := resolveBusinessID(ctx, user)
	if bizErr != nil {
		return ListOrdersResponse{}, bizErr
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 50
	}
	if pageSize > 200 {
		pageSize = 200
	}

	orders, total, hasMore, err := h.Service.ListOrdersPage(ctx, businessID, req.SessionID, req.Status, req.DateFrom, req.DateTo, req.Search, page, pageSize)
	if err != nil {
		return ListOrdersResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list orders")
	}

	payloads := make([]OrderPayload, 0, len(orders))
	for _, o := range orders {
		payloads = append(payloads, newOrderPayload(o))
	}

	return ListOrdersResponse{Orders: payloads, Total: total, HasMore: hasMore, Page: page, PageSize: pageSize}, nil
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

func (h *CreateOrderHandler) Validate(_ context.Context, req *CreateOrderRequest) error {
	if len(req.Items) == 0 {
		return onekiterrors.New(http.StatusBadRequest, "Order must have at least one item")
	}

	if req.Type != "" && !domain.IsOrderType(domain.OrderType(req.Type)) {
		return onekiterrors.New(http.StatusBadRequest, "Unsupported order type")
	}

	for _, item := range req.Items {
		if item.Name == "" {
			return onekiterrors.New(http.StatusBadRequest, "Each item must have a name")
		}

		if item.Quantity <= 0 {
			return onekiterrors.New(http.StatusBadRequest, "Item quantity must be positive")
		}

		if item.UnitPrice < 0 {
			return onekiterrors.New(http.StatusBadRequest, "Item unit price cannot be negative")
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
		return UpdateOrderStatusResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update order status")
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

func (h *UpdateOrderItemStatusHandler) Validate(_ context.Context, req *UpdateOrderItemStatusRequest) error {
	if req.OrderID == 0 || req.ItemID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "Order id and item id are required")
	}

	if req.Status == "" {
		return onekiterrors.New(http.StatusBadRequest, "Status is required")
	}

	return nil
}

func (h *UpdateOrderItemStatusHandler) Serve(ctx context.Context, req *UpdateOrderItemStatusRequest) (UpdateOrderItemStatusResponse, error) {
	user, authErr := requirePOSPermission(ctx, "pos:orders:create")
	if authErr != nil {
		return UpdateOrderItemStatusResponse{}, authErr
	}

	businessID, bizErr := resolveBusinessID(ctx, user)
	if bizErr != nil {
		return UpdateOrderItemStatusResponse{}, bizErr
	}
	if err := h.Service.UpdateOrderItemStatus(ctx, businessID, req.OrderID, req.ItemID, domain.ItemStatus(req.Status)); err != nil {
		return UpdateOrderItemStatusResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update item status")
	}

	return UpdateOrderItemStatusResponse{Message: "Item status updated"}, nil
}

// ─── Process Payment ──────────────────────────────────────────────────────────

func ProcessPaymentRoutes() []string {
	return []string{"/pos/orders/payment", "/pos/orders/{id}/payment"}
}

type ProcessPaymentHandler struct{ Service *application.Service }

type ProcessPaymentRequest struct {
	OrderID       int64   `json:"order_id"`
	PaymentMethod string  `json:"payment_method"`
	AmountPaid    float64 `json:"amount_paid"`
}

type ProcessPaymentResponse struct {
	Message string `json:"message"`
}

func (h *ProcessPaymentHandler) Validate(ctx context.Context, req *ProcessPaymentRequest) error {
	if req.OrderID == 0 && chi.URLParamFromCtx(ctx, "id") == "" {
		return onekiterrors.New(http.StatusBadRequest, "Order id is required")
	}

	if req.AmountPaid < 0 {
		return onekiterrors.New(http.StatusBadRequest, "Amount paid cannot be negative")
	}

	return nil
}

func (h *ProcessPaymentHandler) Serve(ctx context.Context, req *ProcessPaymentRequest) (ProcessPaymentResponse, error) {
	user, authErr := requirePOSPermission(ctx, "pos:orders:create")
	if authErr != nil {
		return ProcessPaymentResponse{}, authErr
	}

	businessID, bizErr := resolveBusinessID(ctx, user)
	if bizErr != nil {
		return ProcessPaymentResponse{}, bizErr
	}

	orderID := req.OrderID
	if raw := chi.URLParamFromCtx(ctx, "id"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || parsed <= 0 {
			return ProcessPaymentResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid order id")
		}
		orderID = parsed
	}

	if err := h.Service.ProcessPayment(ctx, businessID, domain.PaymentInput{
		OrderID:       orderID,
		PaymentMethod: domain.PaymentMethod(req.PaymentMethod),
		AmountPaid:    req.AmountPaid,
	}); err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			return ProcessPaymentResponse{}, onekiterrors.New(http.StatusNotFound, "Order not found")
		}
		if errors.Is(err, domain.ErrOrderAlreadyPaid) {
			return ProcessPaymentResponse{}, onekiterrors.New(http.StatusConflict, "Order is already paid")
		}
		if errors.Is(err, domain.ErrPaymentInsufficient) {
			return ProcessPaymentResponse{}, onekiterrors.New(http.StatusBadRequest, "Amount paid is less than order total")
		}

		return ProcessPaymentResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to process payment")
	}

	return ProcessPaymentResponse{Message: "Payment processed"}, nil
}

// ─── Process Refund ───────────────────────────────────────────────────────────

func ProcessRefundRoutes() []string { return []string{"/pos/orders/refund", "/pos/orders/{id}/refund"} }

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

func (h *ProcessRefundHandler) Validate(ctx context.Context, req *ProcessRefundRequest) error {
	if req.OrderID == 0 && chi.URLParamFromCtx(ctx, "id") == "" {
		return onekiterrors.New(http.StatusBadRequest, "Order id is required")
	}

	if req.Amount <= 0 {
		return onekiterrors.New(http.StatusBadRequest, "Refund amount must be positive")
	}

	return nil
}

func (h *ProcessRefundHandler) Serve(ctx context.Context, req *ProcessRefundRequest) (ProcessRefundResponse, error) {
	user, authErr := requirePOSPermission(ctx, "pos:orders:refund")
	if authErr != nil {
		return ProcessRefundResponse{}, authErr
	}

	businessID, bizErr := resolveBusinessID(ctx, user)
	if bizErr != nil {
		return ProcessRefundResponse{}, bizErr
	}
	staffID := user.UserID

	orderID := req.OrderID
	if raw := chi.URLParamFromCtx(ctx, "id"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || parsed <= 0 {
			return ProcessRefundResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid order id")
		}
		orderID = parsed
	}

	refundID, err := h.Service.ProcessRefund(ctx, businessID, staffID, domain.RefundInput{
		OrderID: orderID,
		Amount:  req.Amount,
		Reason:  req.Reason,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrOrderNotFound):
			return ProcessRefundResponse{}, onekiterrors.New(http.StatusNotFound, "Order not found")
		case errors.Is(err, domain.ErrOrderNotPaid):
			return ProcessRefundResponse{}, onekiterrors.New(http.StatusConflict, "Order has not been paid")
		case errors.Is(err, domain.ErrRefundExceedsTotal):
			return ProcessRefundResponse{}, onekiterrors.New(http.StatusBadRequest, "Refund exceeds order total")
		default:
			return ProcessRefundResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to process refund")
		}
	}

	return ProcessRefundResponse{RefundID: refundID, Message: "Refund processed"}, nil
}

type ApproveRefundOrderHandler struct{ Service *application.Service }

func ApproveRefundOrderRoutes() []string {
	return []string{"/pos/orders/{id}/refund/approve"}
}

func (h *ApproveRefundOrderHandler) Validate(ctx context.Context, req *ProcessRefundRequest) error {
	if req.OrderID == 0 && chi.URLParamFromCtx(ctx, "id") == "" {
		return onekiterrors.New(http.StatusBadRequest, "Order id is required")
	}
	return nil
}

func (h *ApproveRefundOrderHandler) Serve(ctx context.Context, req *ProcessRefundRequest) (ProcessRefundResponse, error) {
	user, authErr := requirePOSPermission(ctx, "pos:orders:refund:approve")
	if authErr != nil {
		return ProcessRefundResponse{}, authErr
	}
	businessID, bizErr := resolveBusinessID(ctx, user)
	if bizErr != nil {
		return ProcessRefundResponse{}, bizErr
	}
	orderID := req.OrderID
	if raw := chi.URLParamFromCtx(ctx, "id"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || parsed <= 0 {
			return ProcessRefundResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid order id")
		}
		orderID = parsed
	}
	refundID, err := h.Service.ProcessRefund(ctx, businessID, user.UserID, domain.RefundInput{
		OrderID: orderID,
		Amount:  req.Amount,
		Reason:  req.Reason,
	})
	if err != nil {
		return ProcessRefundResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to process refund")
	}
	return ProcessRefundResponse{RefundID: refundID, Message: "Refund processed"}, nil
}

type OverrideRefundOrderHandler struct{ Service *application.Service }

func OverrideRefundOrderRoutes() []string {
	return []string{"/pos/orders/{id}/refund/override"}
}

func (h *OverrideRefundOrderHandler) Validate(ctx context.Context, req *ProcessRefundRequest) error {
	return (&ApproveRefundOrderHandler{}).Validate(ctx, req)
}

func (h *OverrideRefundOrderHandler) Serve(ctx context.Context, req *ProcessRefundRequest) (ProcessRefundResponse, error) {
	user, authErr := requirePOSPermission(ctx, "pos:orders:refund:override")
	if authErr != nil {
		return ProcessRefundResponse{}, authErr
	}
	businessID, bizErr := resolveBusinessID(ctx, user)
	if bizErr != nil {
		return ProcessRefundResponse{}, bizErr
	}
	orderID := req.OrderID
	if raw := chi.URLParamFromCtx(ctx, "id"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || parsed <= 0 {
			return ProcessRefundResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid order id")
		}
		orderID = parsed
	}
	refundID, err := h.Service.ProcessRefund(ctx, businessID, user.UserID, domain.RefundInput{
		OrderID: orderID,
		Amount:  req.Amount,
		Reason:  req.Reason,
	})
	if err != nil {
		return ProcessRefundResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to process refund")
	}
	return ProcessRefundResponse{RefundID: refundID, Message: "Refund processed"}, nil
}
