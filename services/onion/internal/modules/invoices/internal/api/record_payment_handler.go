package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/invoices/internal/application"
	"github.com/voxire/onion/internal/modules/invoices/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	financeinvoicesv1 "github.com/voxire/voxie/gen/go/finance/invoices/v1"
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
	Payment *financeinvoicesv1.RecordPaymentResponsePayload `json:"payment"`
	Message string                                          `json:"message"`
}

func (h *RecordPaymentHandler) Validate(_ context.Context, req *RecordPaymentRequest) error {
	if req.InvoiceID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "Invoice ID is required")
	}

	if req.Amount <= 0 {
		return onekiterrors.New(http.StatusBadRequest, "Amount must be greater than zero")
	}

	if req.PaymentDate != "" {
		if _, err := time.Parse("2006-01-02", req.PaymentDate); err != nil {
			return onekiterrors.New(http.StatusBadRequest, "Payment date must be in YYYY-MM-DD format")
		}
	}

	return nil
}

func (h *RecordPaymentHandler) Serve(ctx context.Context, req *RecordPaymentRequest) (RecordPaymentResponse, error) {
	device := authPkg.DeviceCtx(ctx)

	if device.User == nil {
		return RecordPaymentResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireFinancePermission(ctx, "finance:invoices:create"); err != nil {
		return RecordPaymentResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return RecordPaymentResponse{}, onekiterrors.PermissionDenied()
	}

	paymentDate := req.PaymentDate
	if paymentDate == "" {
		paymentDate = time.Now().Format("2006-01-02")
	}

	method := req.Method
	if method == "" {
		method = "cash"
	}

	input := domain.PaymentInput{
		InvoiceID:   req.InvoiceID,
		Amount:      req.Amount,
		Currency:    req.Currency,
		PaymentDate: paymentDate,
		Method:      method,
		Reference:   req.Reference,
		Notes:       req.Notes,
	}

	p, err := h.FinanceService.RecordPayment(ctx, businessID, input)
	if err != nil {
		if errors.Is(err, domain.ErrInvoiceNotFound) {
			return RecordPaymentResponse{}, onekiterrors.New(http.StatusNotFound, "Invoice not found")
		}

		if errors.Is(err, domain.ErrInvalidPaymentAmount) {
			return RecordPaymentResponse{}, onekiterrors.New(http.StatusBadRequest, "Amount must be greater than zero")
		}

		return RecordPaymentResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to record payment")
	}

	return RecordPaymentResponse{
		Payment: &financeinvoicesv1.RecordPaymentResponsePayload{
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
