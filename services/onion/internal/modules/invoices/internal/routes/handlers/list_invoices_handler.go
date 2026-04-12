package handlers

import (
	"context"
	"net/http"

	finv1 "github.com/voxire/voxie/gen/go/finance/v1"

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
	DateFrom string `json:"date_from"`
	DateTo   string `json:"date_to"`
}

type ListInvoicesResponse struct {
	Invoices []*finv1.InvoiceSummary `json:"invoices"`
}

func newInvoiceSummary(inv domain.Invoice) *finv1.InvoiceSummary {
	paidDate := ""
	if inv.PaidDate != nil {
		paidDate = inv.PaidDate.Format("2006-01-02")
	}

	return &finv1.InvoiceSummary{
		Id:            inv.ID,
		InvoiceNumber: inv.InvoiceNumber,
		ClientName:    inv.ClientName,
		Status:        inv.Status,
		IssueDate:     inv.IssueDate.Format("2006-01-02"),
		DueDate:       inv.DueDate.Format("2006-01-02"),
		PaidDate:      paidDate,
		Currency:      inv.Currency,
		Total:         inv.Total,
		AmountPaid:    inv.AmountPaid,
		AmountDue:     inv.AmountDue,
		CreatedAt:     inv.CreatedAt.Format("2006-01-02T15:04:05Z"),
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

	summaries := make([]*finv1.InvoiceSummary, 0, len(invoices))
	for _, inv := range invoices {
		summaries = append(summaries, newInvoiceSummary(inv))
	}

	return ListInvoicesResponse{Invoices: summaries}, nil
}
