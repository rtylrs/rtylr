package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/pos_receipts/internal/application"
	"github.com/voxire/onion/internal/modules/pos_receipts/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

type ReceiptPayload struct {
	ID            int64   `json:"id"`
	OrderNumber   string  `json:"order_number"`
	CustomerName  string  `json:"customer_name"`
	Total         float64 `json:"total"`
	PaymentMethod string  `json:"payment_method"`
	PaidAt        string  `json:"paid_at"`
}

func newReceiptPayload(r domain.Receipt) ReceiptPayload {
	return ReceiptPayload{
		ID:            r.ID,
		OrderNumber:   r.OrderNumber,
		CustomerName:  r.CustomerName,
		Total:         r.Total,
		PaymentMethod: r.PaymentMethod,
		PaidAt:        r.PaidAt.Format("2006-01-02T15:04:05Z"),
	}
}

// --- List ---

func ListReceiptsRoutes() []string { return []string{"/pos/receipts"} }

type ListReceiptsHandler struct{ Service *application.Service }

type ListReceiptsRequest struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type ListReceiptsResponse struct {
	Receipts []ReceiptPayload `json:"receipts"`
}

func (h *ListReceiptsHandler) Validate(_ context.Context, req *ListReceiptsRequest) *errorModels.ServerError {
	const dateLayout = "2006-01-02"
	if req.From != "" {
		if _, err := time.Parse(dateLayout, req.From); err != nil {
			return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "from must be a valid date in YYYY-MM-DD format"}
		}
	}
	if req.To != "" {
		if _, err := time.Parse(dateLayout, req.To); err != nil {
			return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "to must be a valid date in YYYY-MM-DD format"}
		}
	}
	return nil
}

// --- Email Receipt ---

func EmailReceiptRoutes() []string { return []string{"/pos/receipts/{id}/email"} }

type (
	EmailReceiptHandler struct{ Service *application.Service }
	EmailReceiptRequest struct {
		Email string `json:"email"`
	}
)

type EmailReceiptResponse struct {
	Message string `json:"message"`
}

func (h *EmailReceiptHandler) Validate(_ context.Context, req *EmailReceiptRequest) *errorModels.ServerError {
	email := strings.TrimSpace(req.Email)
	if email == "" || !strings.Contains(email, "@") {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "A valid email is required"}
	}
	return nil
}

func (h *EmailReceiptHandler) Serve(ctx context.Context, req *EmailReceiptRequest) (EmailReceiptResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return EmailReceiptResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return EmailReceiptResponse{}, errorModels.PermissionDeniedServerError()
	}
	idStr := chi.URLParamFromCtx(ctx, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return EmailReceiptResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid receipt ID"}
	}
	if sendErr := h.Service.EmailReceipt(ctx, businessID, id, strings.TrimSpace(req.Email)); sendErr != nil {
		if errors.Is(sendErr, domain.ErrReceiptNotFound) {
			return EmailReceiptResponse{}, &errorModels.ServerError{Code: http.StatusNotFound, Message: "Receipt not found"}
		}
		return EmailReceiptResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to send receipt email"}
	}
	return EmailReceiptResponse{Message: "Receipt sent"}, nil
}

func (h *ListReceiptsHandler) Serve(ctx context.Context, req *ListReceiptsRequest) (ListReceiptsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListReceiptsResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListReceiptsResponse{}, errorModels.PermissionDeniedServerError()
	}

	receipts, err := h.Service.List(ctx, businessID, req.From, req.To)
	if err != nil {
		return ListReceiptsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list receipts"}
	}
	payloads := make([]ReceiptPayload, len(receipts))
	for i, rc := range receipts {
		payloads[i] = newReceiptPayload(rc)
	}
	return ListReceiptsResponse{Receipts: payloads}, nil
}
