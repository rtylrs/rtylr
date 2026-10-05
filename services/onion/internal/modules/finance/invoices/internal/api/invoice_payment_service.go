package api

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/finance/invoices/internal/application"
	"github.com/voxire/onion/internal/modules/finance/invoices/internal/domain"
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

func normalizeInvoicePaymentMethod(value string) (string, bool) {
	method := strings.ToLower(strings.TrimSpace(value))
	method = strings.ReplaceAll(method, "-", "_")
	method = strings.ReplaceAll(method, " ", "_")

	switch method {
	case "", domain.PaymentMethodCash:
		return domain.PaymentMethodCash, true
	case domain.PaymentMethodCard, "credit_card", "debit_card":
		return domain.PaymentMethodCard, true
	case domain.PaymentMethodBankTransfer, "transfer", "bank", "ach", "wire", "wire_transfer":
		return domain.PaymentMethodBankTransfer, true
	case "cheque", "check":
		return "cheque", true
	case "online", "paypal", "stripe":
		return "online", true
	case domain.PaymentMethodOther:
		return domain.PaymentMethodOther, true
	default:
		return "", false
	}
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

	method, ok := normalizeInvoicePaymentMethod(req.GetMethod())
	if !ok {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid payment method")
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
		return nil, invoiceServiceError(err, "Failed to record payment")
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
