package api

import (
	"github.com/voxire/onion/internal/modules/finance/invoices/internal/application"
	financeinvoicesv1 "github.com/voxire/voxie/gen/go/finance/invoices/v1"
)

type GetInvoiceHandler struct {
	FinanceService *application.Service
}

type GetInvoiceRequest struct {
	ID int64 `json:"id"`
}

type GetInvoiceResponse struct {
	Invoice  *financeinvoicesv1.InvoiceDetail `json:"invoice"`
	Payments []*financeinvoicesv1.Payment     `json:"payments"`
}
