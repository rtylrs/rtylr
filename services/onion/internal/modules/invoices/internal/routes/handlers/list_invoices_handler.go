package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/invoices/internal/application"
	"github.com/voxire/onion/internal/modules/invoices/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func ListInvoicesRoutes() []string {
	return []string{"/finance/invoices"}
}

type ListInvoicesHandler struct {
	FinanceService *application.Service
}

type ListInvoicesRequest struct {
	Status   string `json:"status"`
	Search   string `json:"search"`
	DateFrom string `json:"dateFrom"`
	DateTo   string `json:"dateTo"`
}

type InvoiceSummary struct {
	ID            int64      `json:"id"`
	InvoiceNumber string     `json:"invoiceNumber"`
	ClientName    string     `json:"clientName"`
	Status        string     `json:"status"`
	IssueDate     time.Time  `json:"issueDate"`
	DueDate       time.Time  `json:"dueDate"`
	PaidDate      *time.Time `json:"paidDate"`
	Currency      string     `json:"currency"`
	Total         float64    `json:"total"`
	AmountPaid    float64    `json:"amountPaid"`
	AmountDue     float64    `json:"amountDue"`
	CreatedAt     time.Time  `json:"createdAt"`
}

type ListInvoicesResponse struct {
	Invoices []InvoiceSummary `json:"invoices"`
}

func newInvoiceSummary(inv domain.Invoice) InvoiceSummary {
	return InvoiceSummary{
		ID:            inv.ID,
		InvoiceNumber: inv.InvoiceNumber,
		ClientName:    inv.ClientName,
		Status:        inv.Status,
		IssueDate:     inv.IssueDate,
		DueDate:       inv.DueDate,
		PaidDate:      inv.PaidDate,
		Currency:      inv.Currency,
		Total:         inv.Total,
		AmountPaid:    inv.AmountPaid,
		AmountDue:     inv.AmountDue,
		CreatedAt:     inv.CreatedAt,
	}
}

func (h *ListInvoicesHandler) Serve(ctx context.Context, req *ListInvoicesRequest) (ListInvoicesResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)

	invoices, err := h.FinanceService.ListInvoices(ctx, device.User.UserID, req.Status, req.Search, req.DateFrom, req.DateTo)
	if err != nil {
		return ListInvoicesResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to list invoices",
		}
	}

	summaries := make([]InvoiceSummary, 0, len(invoices))
	for _, inv := range invoices {
		summaries = append(summaries, newInvoiceSummary(inv))
	}

	return ListInvoicesResponse{Invoices: summaries}, nil
}
