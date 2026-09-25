package api

import (
	"context"
	"net/http"
	"time"

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
	ComboID     int64   `json:"combo_id"`
	ModifierIDs []int64 `json:"modifier_ids"`
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
