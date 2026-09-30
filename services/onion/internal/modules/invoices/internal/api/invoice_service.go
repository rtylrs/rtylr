package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/invoices/internal/application"
	"github.com/voxire/onion/internal/modules/invoices/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	financeinvoicesv1 "github.com/voxire/voxie/gen/go/finance/invoices/v1"
)

type InvoicesAPI struct {
	Service *application.Service
}

func NewInvoicesAPI(service *application.Service) *InvoicesAPI {
	return &InvoicesAPI{Service: service}
}

func (a *InvoicesAPI) ListInvoices(ctx context.Context, req *financeinvoicesv1.ListInvoicesRequest) (*financeinvoicesv1.ListInvoicesResponse, error) {
	if err := requireFinancePermission(ctx, "finance:invoices:view"); err != nil {
		return nil, err
	}
	businessID, errModel := invoiceBusinessID(ctx)
	if errModel != nil {
		return nil, errModel
	}

	page := req.GetPage()
	if page < 1 {
		page = 1
	}
	pageSize := req.GetPageSize()
	if pageSize < 1 {
		pageSize = 50
	}

	var (
		invoices []domain.Invoice
		hasMore  bool
		err      error
	)
	// CRM->FIN-02's read surface: per-customer revenue starts here. The
	// business+customer index serves this shape (schemas zz migration).
	if req.GetCustomerId() > 0 {
		invoices, hasMore, err = a.Service.ListInvoicesForCustomer(ctx, businessID, req.GetCustomerId(), page, pageSize, req.GetStatus(), req.GetSearch(), req.GetDateFrom(), req.GetDateTo())
	} else {
		invoices, hasMore, err = a.Service.ListInvoices(ctx, businessID, req.GetStatus(), req.GetSearch(), req.GetDateFrom(), req.GetDateTo(), page, pageSize)
	}
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list invoices")
	}

	summaries := make([]*financeinvoicesv1.InvoiceSummary, 0, len(invoices))
	for _, inv := range invoices {
		summaries = append(summaries, newInvoiceSummary(inv))
	}

	return &financeinvoicesv1.ListInvoicesResponse{
		Invoices: summaries,
		Page:     page,
		PageSize: pageSize,
		HasMore:  hasMore,
	}, nil
}

func (a *InvoicesAPI) ListBillableTimeEntries(ctx context.Context, req *financeinvoicesv1.ListBillableTimeEntriesRequest) (*financeinvoicesv1.ListBillableTimeEntriesResponse, error) {
	if err := requireFinancePermission(ctx, "finance:invoices:view"); err != nil {
		return nil, err
	}
	businessID, errModel := invoiceBusinessID(ctx)
	if errModel != nil {
		return nil, errModel
	}

	page := req.GetPage()
	if page < 1 {
		page = 1
	}
	pageSize := req.GetPageSize()
	if pageSize < 1 {
		pageSize = 50
	}

	entries, total, totalMinutes, hasMore, err := a.Service.ListUnbilledBillableTimeEntries(ctx, businessID, page, pageSize)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list billable time entries")
	}

	payloads := make([]*financeinvoicesv1.BillableTimeEntry, 0, len(entries))
	for _, entry := range entries {
		payload := &financeinvoicesv1.BillableTimeEntry{
			Id:              entry.ID,
			EmployeeId:      entry.EmployeeID,
			Date:            entry.Date,
			DurationMinutes: entry.DurationMinutes,
			Description:     entry.Description,
			IsBillable:      entry.IsBillable,
			Status:          entry.Status,
			Source:          entry.Source,
		}
		if entry.ProjectID != nil {
			projectID := *entry.ProjectID
			payload.ProjectId = &projectID
		}
		if entry.TaskID != nil {
			taskID := *entry.TaskID
			payload.TaskId = &taskID
		}
		payloads = append(payloads, payload)
	}

	return &financeinvoicesv1.ListBillableTimeEntriesResponse{
		Entries:      payloads,
		Total:        total,
		TotalMinutes: totalMinutes,
		Page:         page,
		PageSize:     pageSize,
		HasMore:      hasMore,
	}, nil
}

func (a *InvoicesAPI) CreateInvoice(ctx context.Context, req *financeinvoicesv1.CreateInvoiceRequest) (*financeinvoicesv1.CreateInvoiceResponse, error) {
	if err := requireFinancePermission(ctx, "finance:invoices:create"); err != nil {
		return nil, err
	}
	businessID, errModel := invoiceBusinessID(ctx)
	if errModel != nil {
		return nil, errModel
	}

	inv, err := a.Service.CreateInvoice(ctx, businessID, domain.InvoiceInput{
		CustomerID:    req.GetCustomerId(),
		ClientName:    req.GetClientName(),
		ClientEmail:   req.GetClientEmail(),
		ClientAddress: req.GetClientAddress(),
		ClientTaxID:   req.GetClientTaxId(),
		IssueDate:     req.GetIssueDate(),
		DueDate:       req.GetDueDate(),
		Currency:      req.GetCurrency(),
		DiscountType:  req.GetDiscountType(),
		DiscountValue: req.GetDiscountValue(),
		TaxRate:       req.GetTaxRate(),
		Notes:         req.GetNotes(),
		Terms:         req.GetTerms(),
		Items:         toInvoiceItemInputs(req.GetItems()),
	})
	if err != nil {
		return nil, invoiceServiceError(err, "Failed to create invoice")
	}
	if inv.CreatedAt.IsZero() || inv.UpdatedAt.IsZero() {
		inv, err = a.Service.GetInvoice(ctx, businessID, inv.ID)
		if err != nil {
			return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to load created invoice")
		}
	}

	return &financeinvoicesv1.CreateInvoiceResponse{Invoice: newInvoiceDetailPayload(inv)}, nil
}

func (a *InvoicesAPI) GetInvoice(ctx context.Context, req *financeinvoicesv1.GetInvoiceRequest) (*financeinvoicesv1.GetInvoiceResponse, error) {
	if err := requireFinancePermission(ctx, "finance:invoices:view"); err != nil {
		return nil, err
	}
	businessID, errModel := invoiceBusinessID(ctx)
	if errModel != nil {
		return nil, errModel
	}

	inv, err := a.Service.GetInvoice(ctx, businessID, req.GetId())
	if err != nil {
		if errors.Is(err, domain.ErrInvoiceNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Invoice not found")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get invoice")
	}

	payments := make([]*financeinvoicesv1.Payment, 0, len(inv.Payments))
	for _, p := range inv.Payments {
		payments = append(payments, &financeinvoicesv1.Payment{
			Id:          p.ID,
			Amount:      p.Amount,
			Currency:    p.Currency,
			PaymentDate: p.PaymentDate.Format("2006-01-02"),
			Method:      p.Method,
			Reference:   p.Reference,
			Notes:       p.Notes,
		})
	}

	return &financeinvoicesv1.GetInvoiceResponse{
		Invoice:  newInvoiceDetailPayload(inv),
		Payments: payments,
	}, nil
}

func (a *InvoicesAPI) UpdateInvoice(ctx context.Context, req *financeinvoicesv1.UpdateInvoiceRequest) (*financeinvoicesv1.UpdateInvoiceResponse, error) {
	if err := requireFinancePermission(ctx, "finance:invoices:create"); err != nil {
		return nil, err
	}
	businessID, errModel := invoiceBusinessID(ctx)
	if errModel != nil {
		return nil, errModel
	}

	body := req.GetBody()
	if body == nil {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invoice body is required")
	}

	inv, err := a.Service.UpdateInvoice(ctx, businessID, req.GetId(), domain.InvoiceUpdateInput{
		ClientName:    body.GetClientName(),
		ClientEmail:   body.GetClientEmail(),
		ClientAddress: body.GetClientAddress(),
		ClientTaxID:   body.GetClientTaxId(),
		IssueDate:     body.GetIssueDate(),
		DueDate:       body.GetDueDate(),
		Currency:      body.GetCurrency(),
		DiscountType:  body.GetDiscountType(),
		DiscountValue: body.GetDiscountValue(),
		TaxRate:       body.GetTaxRate(),
		Notes:         body.GetNotes(),
		Terms:         body.GetTerms(),
		Items:         toInvoiceItemInputs(body.GetItems()),
	})
	if err != nil {
		return nil, invoiceServiceError(err, "Failed to update invoice")
	}

	return &financeinvoicesv1.UpdateInvoiceResponse{Invoice: newInvoiceDetailPayload(inv)}, nil
}

func (a *InvoicesAPI) UpdateInvoiceStatus(ctx context.Context, req *financeinvoicesv1.UpdateInvoiceStatusRequest) (*financeinvoicesv1.UpdateInvoiceStatusResponse, error) {
	if err := requireFinancePermission(ctx, "finance:invoices:create"); err != nil {
		return nil, err
	}
	businessID, errModel := invoiceBusinessID(ctx)
	if errModel != nil {
		return nil, errModel
	}

	body := req.GetBody()
	if body == nil || body.GetStatus() == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "Status is required")
	}
	// The valid-status set and the transition rules live in the domain and are
	// applied by the service, so this endpoint and every other caller of it
	// reject the same statuses.
	if err := a.Service.UpdateInvoiceStatus(ctx, businessID, req.GetId(), body.GetStatus()); err != nil {
		return nil, invoiceServiceError(err, "Failed to update invoice status")
	}

	return &financeinvoicesv1.UpdateInvoiceStatusResponse{Message: "Invoice status updated successfully"}, nil
}

func (a *InvoicesAPI) DeleteInvoice(ctx context.Context, req *financeinvoicesv1.DeleteInvoiceRequest) (*financeinvoicesv1.DeleteInvoiceResponse, error) {
	if err := requireFinancePermission(ctx, "finance:invoices:create"); err != nil {
		return nil, err
	}
	businessID, errModel := invoiceBusinessID(ctx)
	if errModel != nil {
		return nil, errModel
	}

	inv, err := a.Service.GetInvoice(ctx, businessID, req.GetId())
	if err != nil {
		if errors.Is(err, domain.ErrInvoiceNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Invoice not found")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to load invoice")
	}
	if inv.Status != domain.StatusDraft {
		return nil, onekiterrors.New(http.StatusConflict, "Only draft invoices can be deleted")
	}

	if err := a.Service.DeleteInvoice(ctx, businessID, req.GetId()); err != nil {
		if errors.Is(err, domain.ErrInvoiceNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Invoice not found")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to delete invoice")
	}

	remaining, err := a.Service.GetInvoice(ctx, businessID, req.GetId())
	if errors.Is(err, domain.ErrInvoiceNotFound) {
		return &financeinvoicesv1.DeleteInvoiceResponse{Message: "Invoice deleted successfully"}, nil
	}
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to verify invoice deletion")
	}
	if remaining.Status != domain.StatusDraft {
		return nil, onekiterrors.New(http.StatusConflict, "Invoice status changed before deletion")
	}

	return nil, onekiterrors.New(http.StatusConflict, "Invoice could not be deleted")
}

func (a *InvoicesAPI) SendInvoice(ctx context.Context, req *financeinvoicesv1.SendInvoiceRequest) (*financeinvoicesv1.SendInvoiceResponse, error) {
	if err := requireFinancePermission(ctx, "finance:invoices:send"); err != nil {
		return nil, err
	}
	businessID, errModel := invoiceBusinessID(ctx)
	if errModel != nil {
		return nil, errModel
	}

	inv, err := a.Service.GetInvoice(ctx, businessID, req.GetId())
	if err != nil {
		if errors.Is(err, domain.ErrInvoiceNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Invoice not found")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to load invoice")
	}
	if domain.IsTerminal(inv.Status) {
		return nil, onekiterrors.New(http.StatusConflict, "Terminal invoices cannot be sent")
	}

	if err := a.Service.SendInvoice(ctx, businessID, req.GetId()); err != nil {
		if errors.Is(err, domain.ErrInvoiceNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Invoice not found")
		}
		if errors.Is(err, domain.ErrInvoiceNotDraft) {
			return nil, onekiterrors.New(http.StatusConflict, "Terminal invoices cannot be sent")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to send invoice")
	}

	return &financeinvoicesv1.SendInvoiceResponse{Message: "Invoice sent successfully"}, nil
}

func toInvoiceItemInputs(items []*financeinvoicesv1.InvoiceItem) []domain.InvoiceItemInput {
	out := make([]domain.InvoiceItemInput, 0, len(items))
	for _, item := range items {
		out = append(out, domain.InvoiceItemInput{
			Description: item.GetDescription(),
			Quantity:    item.GetQuantity(),
			Unit:        item.GetUnit(),
			UnitPrice:   item.GetUnitPrice(),
			TaxRate:     item.GetTaxRate(),
			DiscountPct: item.GetDiscountPct(),
			SortOrder:   int(item.GetSortOrder()),
		})
	}
	return out
}

func invoiceBusinessID(ctx context.Context) (int64, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return 0, onekiterrors.PermissionDenied()
	}
	businessID, err := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if err != nil {
		return 0, onekiterrors.PermissionDenied()
	}
	return businessID, nil
}
