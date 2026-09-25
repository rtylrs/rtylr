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
	user, authErr := requirePOSPermission(ctx, "pos:orders:create")
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
		if errors.Is(err, domain.ErrOrderCannotBeSplit) {
			return SplitOrderResponse{}, onekiterrors.New(http.StatusConflict, "Order cannot be split in its current status")
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
	user, authErr := requirePOSPermission(ctx, "pos:orders:discount")
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
		if errors.Is(err, domain.ErrOrderNotFound) {
			return VoidOrderResponse{}, onekiterrors.New(http.StatusNotFound, "Order not found")
		}
		return VoidOrderResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to void order")
	}
	return VoidOrderResponse{Message: "Order voided successfully"}, nil
}

type VoidOrderApprovalRequest struct {
	OrderID int64  `json:"order_id"`
	Reason  string `json:"reason"`
}
