package invoices

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/finance/invoices/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func (m *Module) executeSendInvoiceTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	invoiceID, err := parseSendInvoiceToolInput(input)
	if err != nil {
		return nil, err
	}

	if err := m.service.SendInvoice(ctx, businessID, invoiceID); err != nil {
		return nil, err
	}

	return map[string]any{
		"invoice": map[string]any{
			"id":     invoiceID,
			"status": domain.StatusSent,
		},
		"message": "Invoice sent successfully",
	}, nil
}

func parseListInvoicesToolInput(input map[string]any) (listInvoicesToolInput, error) {
	status := normalizeListInvoicesToolStatus(input["status"])
	if strings.TrimSpace(invoiceToolString(input["status"])) != "" && status == "" {
		return listInvoicesToolInput{}, errors.New("status must be one of draft, sent, viewed, partial, paid, overdue, or void")
	}

	dateFrom, err := invoiceToolDate(firstInvoiceToolValue(input, "date_from", "from", "start_date", "issue_date_from"))
	if err != nil {
		return listInvoicesToolInput{}, fmt.Errorf("invalid date_from: %w", err)
	}
	dateTo, err := invoiceToolDate(firstInvoiceToolValue(input, "date_to", "to", "end_date", "issue_date_to"))
	if err != nil {
		return listInvoicesToolInput{}, fmt.Errorf("invalid date_to: %w", err)
	}
	if err := validateListInvoicesToolDateRange(dateFrom, dateTo); err != nil {
		return listInvoicesToolInput{}, err
	}

	page, err := invoiceToolInt64(firstInvoiceToolValue(input, "page"))
	if err != nil {
		return listInvoicesToolInput{}, fmt.Errorf("invalid page: %w", err)
	}
	if page < 0 {
		return listInvoicesToolInput{}, errors.New("page must be non-negative")
	}
	if page == 0 {
		page = 1
	}

	pageSize, err := invoiceToolInt64(firstInvoiceToolValue(input, "page_size", "limit"))
	if err != nil {
		return listInvoicesToolInput{}, fmt.Errorf("invalid page_size: %w", err)
	}
	if pageSize < 0 {
		return listInvoicesToolInput{}, errors.New("page_size must be non-negative")
	}
	if pageSize == 0 {
		pageSize = defaultListInvoicesToolPageSize
	}
	if pageSize > maxListInvoicesToolPageSize {
		pageSize = maxListInvoicesToolPageSize
	}

	return listInvoicesToolInput{
		Status:   status,
		Search:   strings.TrimSpace(invoiceToolString(firstInvoiceToolValue(input, "search", "query", "q", "customer", "client_name", "customer_name"))),
		DateFrom: dateFrom,
		DateTo:   dateTo,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func parseGetInvoiceToolInput(input map[string]any) (int64, error) {
	invoiceID, err := invoiceToolInt64(firstInvoiceToolValue(input, "invoice_id", "invoice", "id"))
	if err != nil {
		return 0, fmt.Errorf("invalid invoice_id: %w", err)
	}
	if invoiceID <= 0 {
		return 0, errors.New("invoice_id is required")
	}
	return invoiceID, nil
}

func parseCreateInvoiceToolInput(input map[string]any, now time.Time) (domain.InvoiceInput, error) {
	clientName := strings.TrimSpace(invoiceToolString(firstInvoiceToolValue(input, "client_name", "customer_name", "name")))
	if clientName == "" {
		return domain.InvoiceInput{}, errors.New("client_name is required")
	}

	issueDate, err := invoiceToolDate(firstInvoiceToolValue(input, "issue_date", "invoice_date", "date"))
	if err != nil {
		return domain.InvoiceInput{}, fmt.Errorf("invalid issue_date: %w", err)
	}
	if issueDate == "" {
		issueDate = now.Format("2006-01-02")
	}

	dueDate, err := invoiceToolDate(firstInvoiceToolValue(input, "due_date", "payment_due_date"))
	if err != nil {
		return domain.InvoiceInput{}, fmt.Errorf("invalid due_date: %w", err)
	}
	if dueDate == "" {
		dueDate, err = invoiceToolDueDate(input, issueDate)
		if err != nil {
			return domain.InvoiceInput{}, err
		}
	}
	if err := validateInvoiceToolDateOrder(issueDate, dueDate); err != nil {
		return domain.InvoiceInput{}, err
	}

	currency := strings.ToUpper(strings.TrimSpace(invoiceToolString(input["currency"])))
	if currency != "" && !validInvoiceToolCurrency(currency) {
		return domain.InvoiceInput{}, errors.New("currency must be a 3-letter ISO code")
	}

	discountType := normalizeInvoiceToolDiscountType(input["discount_type"])
	if discountType == "" {
		return domain.InvoiceInput{}, errors.New("discount_type must be fixed or percentage")
	}

	discountValue, err := invoiceToolFloat64(firstInvoiceToolValue(input, "discount_value", "discount"))
	if err != nil {
		return domain.InvoiceInput{}, fmt.Errorf("invalid discount_value: %w", err)
	}
	if discountValue < 0 {
		return domain.InvoiceInput{}, errors.New("discount_value must be non-negative")
	}
	if discountType == "percentage" && discountValue > 100 {
		return domain.InvoiceInput{}, errors.New("discount_value percentage cannot exceed 100")
	}

	taxRate, err := invoiceToolFloat64(firstInvoiceToolValue(input, "tax_rate", "tax"))
	if err != nil {
		return domain.InvoiceInput{}, fmt.Errorf("invalid tax_rate: %w", err)
	}
	if taxRate < 0 || taxRate > 100 {
		return domain.InvoiceInput{}, errors.New("tax_rate must be between 0 and 100")
	}

	items, err := invoiceToolItems(input)
	if err != nil {
		return domain.InvoiceInput{}, fmt.Errorf("invalid items: %w", err)
	}
	if len(items) == 0 {
		return domain.InvoiceInput{}, errors.New("at least one item is required")
	}

	return domain.InvoiceInput{
		ClientName:    clientName,
		ClientEmail:   strings.TrimSpace(invoiceToolString(firstInvoiceToolValue(input, "client_email", "customer_email", "email"))),
		ClientAddress: strings.TrimSpace(invoiceToolString(firstInvoiceToolValue(input, "client_address", "customer_address", "address"))),
		ClientTaxID:   strings.TrimSpace(invoiceToolString(firstInvoiceToolValue(input, "client_tax_id", "customer_tax_id", "tax_id"))),
		IssueDate:     issueDate,
		DueDate:       dueDate,
		Currency:      currency,
		DiscountType:  discountType,
		DiscountValue: discountValue,
		TaxRate:       taxRate,
		Notes:         strings.TrimSpace(invoiceToolString(input["notes"])),
		Terms:         strings.TrimSpace(invoiceToolString(input["terms"])),
		Items:         items,
	}, nil
}

func parseRecordPaymentToolInput(input map[string]any, now time.Time) (domain.PaymentInput, error) {
	invoiceID, err := invoiceToolInt64(firstInvoiceToolValue(input, "invoice_id", "invoice", "id"))
	if err != nil {
		return domain.PaymentInput{}, fmt.Errorf("invalid invoice_id: %w", err)
	}
	if invoiceID <= 0 {
		return domain.PaymentInput{}, errors.New("invoice_id is required")
	}

	amount, err := invoiceToolFloat64(firstInvoiceToolValue(input, "amount", "payment_amount", "paid_amount", "total"))
	if err != nil {
		return domain.PaymentInput{}, fmt.Errorf("invalid amount: %w", err)
	}
	if amount <= 0 {
		return domain.PaymentInput{}, domain.ErrInvalidPaymentAmount
	}

	paymentDate, err := invoiceToolDate(firstInvoiceToolValue(input, "payment_date", "paid_at", "date"))
	if err != nil {
		return domain.PaymentInput{}, fmt.Errorf("invalid payment_date: %w", err)
	}
	if paymentDate == "" {
		paymentDate = now.Format("2006-01-02")
	}

	currency := strings.ToUpper(strings.TrimSpace(invoiceToolString(input["currency"])))
	if currency != "" && !validInvoiceToolCurrency(currency) {
		return domain.PaymentInput{}, errors.New("currency must be a 3-letter ISO code")
	}

	method := normalizePaymentToolMethod(input["method"])
	if method == "" {
		return domain.PaymentInput{}, errors.New("method must be cash, card, bank_transfer, or other")
	}

	return domain.PaymentInput{
		InvoiceID:   invoiceID,
		Amount:      amount,
		Currency:    currency,
		PaymentDate: paymentDate,
		Method:      method,
		Reference:   strings.TrimSpace(invoiceToolString(firstInvoiceToolValue(input, "reference", "ref", "payment_reference", "transaction_id"))),
		Notes:       strings.TrimSpace(invoiceToolString(input["notes"])),
	}, nil
}

func parseSendInvoiceToolInput(input map[string]any) (int64, error) {
	invoiceID, err := invoiceToolInt64(firstInvoiceToolValue(input, "invoice_id", "invoice", "id"))
	if err != nil {
		return 0, fmt.Errorf("invalid invoice_id: %w", err)
	}
	if invoiceID <= 0 {
		return 0, errors.New("invoice_id is required")
	}
	return invoiceID, nil
}
