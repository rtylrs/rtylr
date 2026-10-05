package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/pos/orders/internal/application"
	"github.com/voxire/onion/internal/modules/pos/orders/internal/domain"
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

type SplitOrderHandler struct{ Service *application.Service }

type SplitOrderRequest struct {
	OrderID int64     `json:"order_id"`
	Splits  [][]int64 `json:"splits"`
}

type SplitOrderResponse struct {
	OrderIDs []int64 `json:"order_ids"`
	Message  string  `json:"message"`
}

// ---- Apply Discount ----

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

// ---- Void Order ----

type VoidOrderHandler struct{ Service *application.Service }

// voidOrderError maps a void refusal onto an HTTP status. Shared by all three
// void entry points so a typed refusal cannot be rendered as a 500 by one of
// them: a void the repository refuses because the order already took money or
// is no longer editable is a conflict with the order's state, and the till
// needs to see that as 409 rather than "Failed to void order".
func voidOrderError(err error) error {
	switch {
	case errors.Is(err, domain.ErrOrderNotFound):
		return onekiterrors.New(http.StatusNotFound, "Order not found")
	case errors.Is(err, domain.ErrOrderAlreadyPaid):
		return onekiterrors.New(http.StatusConflict, "Order is already paid")
	case errors.Is(err, domain.ErrOrderNotEditable):
		return onekiterrors.New(http.StatusConflict, "Order cannot be voided in its current status")
	default:
		return onekiterrors.New(http.StatusInternalServerError, "Failed to void order")
	}
}

type VoidOrderRequest struct {
	OrderID int64  `json:"order_id"`
	Reason  string `json:"reason"`
}

type VoidOrderResponse struct {
	Message string `json:"message"`
}

func (h *VoidOrderHandler) Serve(ctx context.Context, req *VoidOrderRequest) (VoidOrderResponse, error) {
	user, authErr := requirePOSPermission(ctx, "pos:orders:void")
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
		return VoidOrderResponse{}, voidOrderError(err)
	}
	return VoidOrderResponse{Message: "Order voided successfully"}, nil
}

type VoidOrderApprovalRequest struct {
	OrderID int64  `json:"order_id"`
	Reason  string `json:"reason"`
}
