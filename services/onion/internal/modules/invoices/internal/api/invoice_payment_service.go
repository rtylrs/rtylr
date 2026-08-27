package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/invoices/internal/application"
	"github.com/voxire/onion/internal/modules/invoices/internal/domain"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	financeinvoicespaymentsv1 "github.com/voxire/voxie/gen/go/finance/invoices/payments/v1"
	financeinvoicesv1 "github.com/voxire/voxie/gen/go/finance/invoices/v1"
)

type InvoicePaymentAPI struct {
	Service *application.Service
}

func NewInvoicePaymentAPI(service *application.Service) *InvoicePaymentAPI {
	return &InvoicePaymentAPI{Service: service}
}

func (a *InvoicePaymentAPI) RecordPayment(ctx context.Context, req *financeinvoicespaymentsv1.RecordPaymentRequest) (*financeinvoicespaymentsv1.RecordPaymentResponse, error) {
	if err := requireFinancePermission(ctx, "finance:invoices:create"); err != nil {
		return nil, err
	}
	businessID, errModel := invoiceBusinessID(ctx)
	if errModel != nil {
		return nil, errModel
	}

	if req.GetAmount() <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Amount must be greater than zero")
	}

	paymentDate := req.GetPaymentDate()
	if paymentDate == "" {
		paymentDate = time.Now().Format("2006-01-02")
	}

	method := req.GetMethod()
	if method == "" {
		method = "cash"
	}

	p, err := a.Service.RecordPayment(ctx, businessID, domain.PaymentInput{
		InvoiceID:   req.GetInvoiceId(),
		Amount:      req.GetAmount(),
		Currency:    req.GetCurrency(),
		PaymentDate: paymentDate,
		Method:      method,
		Reference:   req.GetReference(),
		Notes:       req.GetNotes(),
	})
	if err != nil {
		if errors.Is(err, domain.ErrInvoiceNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Invoice not found")
		}
		if errors.Is(err, domain.ErrInvalidPaymentAmount) {
			return nil, onekiterrors.New(http.StatusBadRequest, "Amount must be greater than zero")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to record payment")
	}

	return &financeinvoicespaymentsv1.RecordPaymentResponse{
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
