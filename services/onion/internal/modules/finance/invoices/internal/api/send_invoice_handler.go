package api

import (
	"github.com/voxire/onion/internal/modules/finance/invoices/internal/application"
)

type SendInvoiceHandler struct {
	FinanceService *application.Service
}

type SendInvoiceRequest struct {
	ID int64 `json:"id"`
}

type SendInvoiceResponse struct {
	Message string `json:"message"`
}
