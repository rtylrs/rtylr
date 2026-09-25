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
		switch {
		case errors.Is(err, domain.ErrOrderNotFound):
			return UpdateOrderItemStatusResponse{}, onekiterrors.New(http.StatusNotFound, "Order item not found")
		case errors.Is(err, domain.ErrInvalidItemStatus):
			return UpdateOrderItemStatusResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid item status")
		default:
			return UpdateOrderItemStatusResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update item status")
		}
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
