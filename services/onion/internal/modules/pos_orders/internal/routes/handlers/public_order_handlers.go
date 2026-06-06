package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/application"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

// ---- Create Public Order (no auth) ----

func CreatePublicOrderRoutes() []string {
	return []string{"/pos/public/orders"}
}

type PublicOrderItemInput struct {
	MenuItemID int64   `json:"menu_item_id"`
	Name       string  `json:"name"`
	Quantity   float64 `json:"quantity"`
	UnitPrice  float64 `json:"unit_price"`
	Notes      string  `json:"notes"`
}

type CreatePublicOrderRequest struct {
	OwnerID       int64                  `json:"owner_id"`
	CustomerName  string                 `json:"customer_name"`
	CustomerPhone string                 `json:"customer_phone"`
	CustomerEmail string                 `json:"customer_email"`
	Notes         string                 `json:"notes"`
	Items         []PublicOrderItemInput `json:"items"`
}

type CreatePublicOrderResponse struct {
	OrderID     int64  `json:"order_id"`
	OrderNumber string `json:"order_number"`
	Status      string `json:"status"`
}

type CreatePublicOrderHandler struct{ Service *application.Service }

func (h *CreatePublicOrderHandler) Validate(_ context.Context, req *CreatePublicOrderRequest) *errorModels.ServerError {
	if req.OwnerID == 0 {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "owner_id is required"}
	}
	if len(req.Items) == 0 {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "order must have at least one item"}
	}
	for _, item := range req.Items {
		if item.MenuItemID <= 0 {
			return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "each item must have a menu_item_id"}
		}
		if item.Quantity <= 0 {
			return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "item quantity must be positive"}
		}
	}
	return nil
}

func (h *CreatePublicOrderHandler) Serve(ctx context.Context, req *CreatePublicOrderRequest) (CreatePublicOrderResponse, *errorModels.ServerError) {
	items := make([]application.PublicOrderItemInput, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, application.PublicOrderItemInput{
			MenuItemID: it.MenuItemID,
			Name:       it.Name,
			Quantity:   it.Quantity,
			UnitPrice:  it.UnitPrice,
			Notes:      it.Notes,
		})
	}

	orderID, orderNumber, err := h.Service.CreatePublicOrder(ctx, req.OwnerID, application.PublicOrderInput{
		CustomerName:  req.CustomerName,
		CustomerPhone: req.CustomerPhone,
		CustomerEmail: req.CustomerEmail,
		Notes:         req.Notes,
		Items:         items,
	})
	if err != nil {
		return CreatePublicOrderResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  fmt.Sprintf("Failed to create order: %v", err),
		}
	}

	return CreatePublicOrderResponse{
		OrderID:     orderID,
		OrderNumber: orderNumber,
		Status:      "pending",
	}, nil
}

// ---- Get Public Order Status (no auth) ----

func GetPublicOrderStatusRoutes() []string {
	return []string{"/pos/public/orders/{id}/status"}
}

type GetPublicOrderStatusRequest struct{}

type GetPublicOrderStatusResponse struct {
	OrderID     int64  `json:"order_id"`
	OrderNumber string `json:"order_number"`
	Status      string `json:"status"`
}

type GetPublicOrderStatusHandler struct{ Service *application.Service }

func (h *GetPublicOrderStatusHandler) Serve(ctx context.Context, _ *GetPublicOrderStatusRequest) (GetPublicOrderStatusResponse, *errorModels.ServerError) {
	rawID := chi.URLParamFromCtx(ctx, "id")
	orderID, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil || orderID <= 0 {
		return GetPublicOrderStatusResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Invalid order id",
		}
	}

	orderNumber, status, err := h.Service.GetPublicOrderStatus(ctx, orderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return GetPublicOrderStatusResponse{}, &errorModels.ServerError{
				HTTPCode: http.StatusNotFound,
				Code:     http.StatusNotFound,
				Message:  "Order not found",
			}
		}
		return GetPublicOrderStatusResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to fetch order status",
		}
	}

	return GetPublicOrderStatusResponse{
		OrderID:     orderID,
		OrderNumber: orderNumber,
		Status:      status,
	}, nil
}
