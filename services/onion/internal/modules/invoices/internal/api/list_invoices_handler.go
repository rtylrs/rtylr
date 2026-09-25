package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/invoices/internal/application"
	"github.com/voxire/onion/internal/modules/invoices/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	financeinvoicesv1 "github.com/voxire/voxie/gen/go/finance/invoices/v1"
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
	Page     int64  `json:"page"`
	PageSize int64  `json:"page_size"`
}

type ListInvoicesResponse struct {
	Invoices []*financeinvoicesv1.InvoiceSummary `json:"invoices"`
	Page     int64                               `json:"page"`
	PageSize int64                               `json:"page_size"`
	HasMore  bool                                `json:"has_more"`
}

func newInvoiceSummary(inv domain.Invoice) *financeinvoicesv1.InvoiceSummary {
	paidDate := ""
	if inv.PaidDate != nil {
		paidDate = inv.PaidDate.Format("2006-01-02")
	}

	return &financeinvoicesv1.InvoiceSummary{
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

func (h *ListInvoicesHandler) Serve(ctx context.Context, req *ListInvoicesRequest) (ListInvoicesResponse, error) {
	device := authPkg.DeviceCtx(ctx)

	if device.User == nil {
		return ListInvoicesResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireFinancePermission(ctx, "finance:invoices:view"); err != nil {
		return ListInvoicesResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListInvoicesResponse{}, onekiterrors.PermissionDenied()
	}

	invoices, hasMore, err := h.FinanceService.ListInvoices(ctx, businessID, req.Status, req.Search, req.DateFrom, req.DateTo, req.Page, req.PageSize)
	if err != nil {
		return ListInvoicesResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list invoices")
	}

	summaries := make([]*financeinvoicesv1.InvoiceSummary, 0, len(invoices))
	for _, inv := range invoices {
		summaries = append(summaries, newInvoiceSummary(inv))
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 50
	}
	return ListInvoicesResponse{Invoices: summaries, Page: page, PageSize: pageSize, HasMore: hasMore}, nil
}
