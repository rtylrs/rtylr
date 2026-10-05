package api

import (
	"github.com/voxire/onion/internal/modules/finance/invoices/internal/application"
	financeinvoicesv1 "github.com/voxire/voxie/gen/go/finance/invoices/v1"
)

type UpdateInvoiceHandler struct {
	FinanceService *application.Service
}

type UpdateInvoiceRequest struct {
	ID            int64                `json:"id"`
	ClientName    string               `json:"client_name"`
	ClientEmail   string               `json:"client_email"`
	ClientAddress string               `json:"client_address"`
	ClientTaxID   string               `json:"client_tax_id"`
	IssueDate     string               `json:"issue_date"`
	DueDate       string               `json:"due_date"`
	Currency      string               `json:"currency"`
	DiscountType  string               `json:"discount_type"`
	DiscountValue float64              `json:"discount_value"`
	TaxRate       float64              `json:"tax_rate"`
	Notes         string               `json:"notes"`
	Terms         string               `json:"terms"`
	Items         []InvoiceItemPayload `json:"items"`
}

type UpdateInvoiceResponse struct {
	Invoice *financeinvoicesv1.InvoiceDetail `json:"invoice"`
}
