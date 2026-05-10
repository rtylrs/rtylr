package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/invoices/internal/application"
	"github.com/voxire/onion/internal/modules/invoices/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	finv1 "github.com/voxire/voxie/gen/go/finance/v1"
)

func RecordPaymentRoutes() []string {
	return []string{"/finance/invoices/payment"}
}

type RecordPaymentHandler struct {
	FinanceService *application.Service
}

type RecordPaymentRequest struct {
	InvoiceID   int64   `json:"invoice_id"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	PaymentDate string  `json:"payment_date"`
	Method      string  `json:"method"`
	Reference   string  `json:"reference"`
	Notes       string  `json:"notes"`
}

type RecordPaymentResponse struct {
	Payment *finv1.RecordPaymentResponsePayload `json:"payment"`
	Message string                              `json:"message"`
}

func (h *RecordPaymentHandler) Validate(_ context.Context, req *RecordPaymentRequest) *errorModels.ServerError {
	if req.InvoiceID == 0 {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Invoice ID is required",
		}
	}

	if req.Amount <= 0 {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Amount must be greater than zero",
		}
	}

	if req.PaymentDate == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Payment date is required",
		}
	}

	if req.Method == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Payment method is required",
		}
	}

	return nil
}

func (h *RecordPaymentHandler) Serve(ctx context.Context, req *RecordPaymentRequest) (RecordPaymentResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)

	input := domain.PaymentInput{
		InvoiceID:   req.InvoiceID,
		Amount:      req.Amount,
		Currency:    req.Currency,
		PaymentDate: req.PaymentDate,
		Method:      req.Method,
		Reference:   req.Reference,
		Notes:       req.Notes,
	}

	p, err := h.FinanceService.RecordPayment(ctx, device.User.UserID, input)
	if err != nil {
		if errors.Is(err, domain.ErrInvoiceNotFound) {
			return RecordPaymentResponse{}, &errorModels.ServerError{
				Code:    http.StatusNotFound,
				Message: "Invoice not found",
			}
		}

		return RecordPaymentResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to record payment",
		}
	}

	return RecordPaymentResponse{
		Payment: &finv1.RecordPaymentResponsePayload{
			Id:          p.ID,
			InvoiceId:   p.InvoiceID,
			Amount:      p.Amount,
			Currency:    p.Currency,
			PaymentDate: p.PaymentDate.Format("2006-01-02"),
			Method:      p.Method,
			Reference:   p.Reference,
			Notes:       p.Notes,
		},
		Message: "Payment recorded successfully",
	}, nil
}
