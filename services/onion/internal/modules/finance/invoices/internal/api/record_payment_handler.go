package api

import (
	"github.com/voxire/onion/internal/modules/finance/invoices/internal/application"
	financeinvoicesv1 "github.com/voxire/voxie/gen/go/finance/invoices/v1"
)

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
