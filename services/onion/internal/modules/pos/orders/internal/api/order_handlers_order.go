package api

import (
	"context"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
	"net/http"
	"strconv"

	"github.com/voxire/onion/internal/modules/pos/orders/internal/application"
	"github.com/voxire/onion/internal/modules/pos/orders/internal/domain"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

type UpdateOrderItemStatusHandler struct{ Service *application.Service }

type UpdateOrderItemStatusRequest struct {
	OrderID int64  `json:"order_id"`
	ItemID  int64  `json:"item_id"`
	Status  string `json:"status"`
}

type UpdateOrderItemStatusResponse struct {
	Message string `json:"message"`
}

// ─── Process Payment ──────────────────────────────────────────────────────────

type ProcessPaymentHandler struct{ Service *application.Service }

type ProcessPaymentRequest struct {
	OrderID       int64   `json:"order_id"`
	PaymentMethod string  `json:"payment_method"`
	AmountPaid    float64 `json:"amount_paid"`
}

type ProcessPaymentResponse struct {
	Message string `json:"message"`
}

// ─── Process Refund ───────────────────────────────────────────────────────────

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

type ApproveRefundOrderHandler struct{ Service *application.Service }

func ApproveRefundOrderRoutes() []string {
	return []string{"/pos/orders/{id}/refund/approve"}
}

func (h *ApproveRefundOrderHandler) Validate(ctx context.Context, req *ProcessRefundRequest) error {
	if req.OrderID == 0 && serverHttp.PathParam(ctx, "id") == "" {
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
	if raw := serverHttp.PathParam(ctx, "id"); raw != "" {
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
	if raw := serverHttp.PathParam(ctx, "id"); raw != "" {
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
