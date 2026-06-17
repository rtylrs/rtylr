package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/application"
	"github.com/voxire/onion/internal/modules/pos_orders/internal/domain"
	errorModels "github.com/voxire/onion/internal/platform/errors"
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

func (h *PayOrderHandler) Serve(ctx context.Context, req *PayOrderRequest) (PayOrderResponse, *errorModels.ServerError) {
	user, authErr := authenticatedPOSUser(ctx)
	if authErr != nil {
		return PayOrderResponse{}, authErr
	}
	ownerID := resolvedOwnerID(ctx, user)

	idStr := chi.URLParamFromCtx(ctx, "id")
	orderID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || orderID <= 0 {
		return PayOrderResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Invalid order id",
		}
	}

	result, payErr := h.Service.CreatePaddlePayment(ctx, ownerID, orderID, req.CustomerEmail, req.Currency)
	if payErr != nil {
		if errors.Is(payErr, domain.ErrOrderNotFound) {
			return PayOrderResponse{}, &errorModels.ServerError{
				HTTPCode: http.StatusNotFound,
				Code:     http.StatusNotFound,
				Message:  "Order not found",
			}
		}
		return PayOrderResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to create payment",
		}
	}

	return PayOrderResponse{
		TransactionID: result.TransactionID,
		CheckoutURL:   result.CheckoutURL,
	}, nil
}
