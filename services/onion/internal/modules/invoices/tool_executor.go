package invoices

import (
	"context"
	"errors"
	"time"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

const (
	listInvoicesToolName            = "finance.list_invoices"
	getInvoiceToolName              = "finance.get_invoice"
	createInvoiceToolName           = "finance.create_invoice"
	recordPaymentToolName           = "finance.record_payment"
	sendInvoiceToolName             = "finance.send_invoice"
	defaultListInvoicesToolPageSize = int64(20)
	maxListInvoicesToolPageSize     = int64(50)
)

type listInvoicesToolInput struct {
	Status   string
	Search   string
	DateFrom string
	DateTo   string
	Page     int64
	PageSize int64
}

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	switch name {
	case listInvoicesToolName:
		output, err := m.executeListInvoicesTool(ctx, input)
		return output, true, err
	case getInvoiceToolName:
		output, err := m.executeGetInvoiceTool(ctx, input)
		return output, true, err
	case createInvoiceToolName:
		output, err := m.executeCreateInvoiceTool(ctx, input)
		return output, true, err
	case recordPaymentToolName:
		output, err := m.executeRecordPaymentTool(ctx, input)
		return output, true, err
	case sendInvoiceToolName:
		output, err := m.executeSendInvoiceTool(ctx, input)
		return output, true, err
	default:
		return nil, false, nil
	}
}

func (m *Module) executeListInvoicesTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	parsed, err := parseListInvoicesToolInput(input)
	if err != nil {
		return nil, err
	}

	invoices, hasMore, err := m.service.ListInvoices(ctx, businessID, parsed.Status, parsed.Search, parsed.DateFrom, parsed.DateTo, parsed.Page, parsed.PageSize)
	if err != nil {
		return nil, err
	}

	payloads := make([]map[string]any, 0, len(invoices))
	for _, invoice := range invoices {
		payloads = append(payloads, invoiceToolPayload(invoice))
	}

	return map[string]any{
		"invoices":  payloads,
		"returned":  len(payloads),
		"page":      parsed.Page,
		"page_size": parsed.PageSize,
		"has_more":  hasMore,
		"filters": map[string]any{
			"status":    parsed.Status,
			"search":    parsed.Search,
			"date_from": parsed.DateFrom,
			"date_to":   parsed.DateTo,
		},
	}, nil
}

func (m *Module) executeGetInvoiceTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	invoiceID, err := parseGetInvoiceToolInput(input)
	if err != nil {
		return nil, err
	}

	invoice, err := m.service.GetInvoice(ctx, businessID, invoiceID)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"invoice": invoiceDetailToolPayload(*invoice),
	}, nil
}

func (m *Module) executeCreateInvoiceTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	parsed, err := parseCreateInvoiceToolInput(input, time.Now())
	if err != nil {
		return nil, err
	}

	invoice, err := m.service.CreateInvoice(ctx, businessID, parsed)
	if err != nil {
		return nil, err
	}

	items := make([]map[string]any, 0, len(invoice.Items))
	for _, item := range invoice.Items {
		items = append(items, map[string]any{
			"description":  item.Description,
			"quantity":     item.Quantity,
			"unit":         item.Unit,
			"unit_price":   item.UnitPrice,
			"tax_rate":     item.TaxRate,
			"discount_pct": item.DiscountPct,
			"total":        item.Total,
			"sort_order":   item.SortOrder,
		})
	}

	return map[string]any{
		"invoice": map[string]any{
			"id":              invoice.ID,
			"invoice_number":  invoice.InvoiceNumber,
			"client_name":     invoice.ClientName,
			"client_email":    invoice.ClientEmail,
			"client_address":  invoice.ClientAddress,
			"client_tax_id":   invoice.ClientTaxID,
			"status":          invoice.Status,
			"issue_date":      invoice.IssueDate.Format("2006-01-02"),
			"due_date":        invoice.DueDate.Format("2006-01-02"),
			"currency":        invoice.Currency,
			"subtotal":        invoice.Subtotal,
			"discount_type":   invoice.DiscountType,
			"discount_value":  invoice.DiscountValue,
			"discount_amount": invoice.DiscountAmount,
			"tax_rate":        invoice.TaxRate,
			"tax_amount":      invoice.TaxAmount,
			"total":           invoice.Total,
			"amount_due":      invoice.AmountDue,
			"notes":           invoice.Notes,
			"terms":           invoice.Terms,
			"items":           items,
		},
	}, nil
}

func (m *Module) executeRecordPaymentTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	parsed, err := parseRecordPaymentToolInput(input, time.Now())
	if err != nil {
		return nil, err
	}

	payment, err := m.service.RecordPayment(ctx, businessID, parsed)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"payment": map[string]any{
			"id":           payment.ID,
			"invoice_id":   payment.InvoiceID,
			"amount":       payment.Amount,
			"currency":     payment.Currency,
			"payment_date": payment.PaymentDate.Format("2006-01-02"),
			"method":       payment.Method,
			"reference":    payment.Reference,
			"notes":        payment.Notes,
		},
	}, nil
}
