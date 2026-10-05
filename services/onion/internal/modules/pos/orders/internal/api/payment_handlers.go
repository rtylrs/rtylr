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

// ─── Pay Order via Paddle ────────────────────────────────────────────────────

func PayOrderRoutes() []string { return []string{"/pos/orders/{id}/pay"} }

type PayOrderHandler struct{ Service *application.Service }

type PayOrderRequest struct {
	CustomerEmail string `json:"customer_email"`
	Currency      string `json:"currency"`
}

type PayOrderResponse struct {
	TransactionID string `json:"transaction_id"`
	CheckoutURL   string `json:"checkout_url"`
}

func paddlePaymentError(err error) error {
	switch {
	case errors.Is(err, domain.ErrOrderNotFound):
		return onekiterrors.New(http.StatusNotFound, "Order not found")
	case errors.Is(err, domain.ErrOrderAlreadyPaid), err.Error() == "order already paid", err.Error() == domain.ErrOrderAlreadyPaid.Error():
		return onekiterrors.New(http.StatusConflict, "Order is already paid")
	case err.Error() == "order total must be positive":
		return onekiterrors.New(http.StatusBadRequest, "Order total must be positive")
	default:
		return onekiterrors.New(http.StatusInternalServerError, "Failed to create payment")
	}
}

func (h *PayOrderHandler) Serve(ctx context.Context, req *PayOrderRequest) (PayOrderResponse, error) {
	user, authErr := requirePOSPermission(ctx, "pos:orders:create")
	if authErr != nil {
		return PayOrderResponse{}, authErr
	}
	businessID, bizErr := resolveBusinessID(ctx, user)
	if bizErr != nil {
		return PayOrderResponse{}, bizErr
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	orderID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || orderID <= 0 {
		return PayOrderResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid order id")
	}

	result, payErr := h.Service.CreatePaddlePayment(ctx, businessID, orderID, req.CustomerEmail, req.Currency)
	if payErr != nil {
		return PayOrderResponse{}, paddlePaymentError(payErr)
	}

	return PayOrderResponse{
		TransactionID: result.TransactionID,
		CheckoutURL:   result.CheckoutURL,
	}, nil
}
