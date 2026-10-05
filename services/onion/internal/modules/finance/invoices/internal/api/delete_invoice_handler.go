package api

import (
	"github.com/voxire/onion/internal/modules/finance/invoices/internal/application"
)

type DeleteInvoiceHandler struct {
	FinanceService *application.Service
}

type DeleteInvoiceRequest struct {
	ID int64 `json:"id"`
}

type DeleteInvoiceResponse struct {
	Message string `json:"message"`
}
