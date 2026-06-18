package invoices

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/invoices/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
)

const (
	listInvoicesToolName            = "finance.list_invoices"
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

	parsed, err := parseListInvoicesToolInput(input)
	if err != nil {
		return nil, err
	}

	invoices, hasMore, err := m.service.ListInvoices(ctx, authUser.UserID, parsed.Status, parsed.Search, parsed.DateFrom, parsed.DateTo, parsed.Page, parsed.PageSize)
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

func (m *Module) executeCreateInvoiceTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	parsed, err := parseCreateInvoiceToolInput(input, time.Now())
	if err != nil {
		return nil, err
	}

	invoice, err := m.service.CreateInvoice(ctx, authUser.UserID, parsed)
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

	parsed, err := parseRecordPaymentToolInput(input, time.Now())
	if err != nil {
		return nil, err
	}

	payment, err := m.service.RecordPayment(ctx, authUser.UserID, parsed)
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

func (m *Module) executeSendInvoiceTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	invoiceID, err := parseSendInvoiceToolInput(input)
	if err != nil {
		return nil, err
	}

	if err := m.service.SendInvoice(ctx, authUser.UserID, invoiceID); err != nil {
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

func invoiceToolDueDate(input map[string]any, issueDate string) (string, error) {
	parsedIssueDate, err := time.Parse("2006-01-02", issueDate)
	if err != nil {
		return "", fmt.Errorf("invalid issue_date: %w", err)
	}

	daysValue := firstInvoiceToolValue(input, "due_in_days", "payment_terms_days", "terms_days", "net_terms_days")
	days := int64(7)
	if daysValue != nil {
		days, err = invoiceToolInt64(daysValue)
		if err != nil {
			return "", fmt.Errorf("invalid due_in_days: %w", err)
		}
	}
	if days < 0 {
		return "", errors.New("due_in_days must be non-negative")
	}
	return parsedIssueDate.AddDate(0, 0, int(days)).Format("2006-01-02"), nil
}

func validateInvoiceToolDateOrder(issueDate, dueDate string) error {
	parsedIssueDate, err := time.Parse("2006-01-02", issueDate)
	if err != nil {
		return fmt.Errorf("invalid issue_date: %w", err)
	}
	parsedDueDate, err := time.Parse("2006-01-02", dueDate)
	if err != nil {
		return fmt.Errorf("invalid due_date: %w", err)
	}
	if parsedDueDate.Before(parsedIssueDate) {
		return domain.ErrInvalidInvoiceDate
	}
	return nil
}

func validateListInvoicesToolDateRange(dateFrom, dateTo string) error {
	if dateFrom == "" || dateTo == "" {
		return nil
	}
	parsedFrom, err := time.Parse("2006-01-02", dateFrom)
	if err != nil {
		return fmt.Errorf("invalid date_from: %w", err)
	}
	parsedTo, err := time.Parse("2006-01-02", dateTo)
	if err != nil {
		return fmt.Errorf("invalid date_to: %w", err)
	}
	if parsedTo.Before(parsedFrom) {
		return errors.New("date_to must be on or after date_from")
	}
	return nil
}

func invoiceToolItems(input map[string]any) ([]domain.InvoiceItemInput, error) {
	value := firstInvoiceToolValue(input, "items", "line_items")
	if value == nil {
		return invoiceToolFallbackItem(input)
	}

	switch v := value.(type) {
	case []any:
		return invoiceToolItemsFromSlice(v, input)
	case []map[string]any:
		values := make([]any, 0, len(v))
		for _, item := range v {
			values = append(values, item)
		}
		return invoiceToolItemsFromSlice(values, input)
	case []domain.InvoiceItemInput:
		return validateInvoiceToolItems(v)
	case map[string]any:
		return invoiceToolItemsFromSlice([]any{v}, input)
	case string:
		raw := strings.TrimSpace(v)
		if raw == "" {
			return invoiceToolFallbackItem(input)
		}
		if json.Valid([]byte(raw)) {
			var decoded any
			decoder := json.NewDecoder(strings.NewReader(raw))
			decoder.UseNumber()
			if err := decoder.Decode(&decoded); err != nil {
				return nil, err
			}
			return invoiceToolItemsWithValue(decoded, input)
		}
		return invoiceToolFallbackItemWithDescription(input, raw)
	default:
		return nil, fmt.Errorf("unsupported value %T", value)
	}
}

func invoiceToolItemsWithValue(value any, input map[string]any) ([]domain.InvoiceItemInput, error) {
	switch v := value.(type) {
	case []any:
		return invoiceToolItemsFromSlice(v, input)
	case map[string]any:
		return invoiceToolItemsFromSlice([]any{v}, input)
	default:
		return nil, fmt.Errorf("unsupported JSON value %T", value)
	}
}

func invoiceToolItemsFromSlice(values []any, input map[string]any) ([]domain.InvoiceItemInput, error) {
	items := make([]domain.InvoiceItemInput, 0, len(values))
	for index, value := range values {
		item, err := invoiceToolItem(value, input, index)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return validateInvoiceToolItems(items)
}

func invoiceToolFallbackItem(input map[string]any) ([]domain.InvoiceItemInput, error) {
	description := strings.TrimSpace(invoiceToolString(firstInvoiceToolValue(input, "description", "item", "service")))
	if description == "" {
		return nil, nil
	}
	return invoiceToolFallbackItemWithDescription(input, description)
}

func invoiceToolFallbackItemWithDescription(input map[string]any, description string) ([]domain.InvoiceItemInput, error) {
	amountValue := firstInvoiceToolValue(input, "unit_price", "price", "rate", "amount")
	if amountValue == nil {
		amountValue = input["total"]
	}
	unitPrice, err := invoiceToolFloat64(amountValue)
	if err != nil {
		return nil, fmt.Errorf("invalid amount: %w", err)
	}
	item := domain.InvoiceItemInput{
		Description: description,
		Quantity:    1,
		Unit:        strings.TrimSpace(invoiceToolString(input["unit"])),
		UnitPrice:   unitPrice,
	}
	return validateInvoiceToolItems([]domain.InvoiceItemInput{item})
}

func invoiceToolItem(value any, input map[string]any, index int) (domain.InvoiceItemInput, error) {
	itemMap, ok := value.(map[string]any)
	if !ok {
		return domain.InvoiceItemInput{}, fmt.Errorf("item %d must be an object", index+1)
	}

	description := strings.TrimSpace(invoiceToolString(firstInvoiceToolValue(itemMap, "description", "name", "item", "service")))
	if description == "" {
		return domain.InvoiceItemInput{}, fmt.Errorf("item %d description is required", index+1)
	}

	quantityValue := firstInvoiceToolValue(itemMap, "quantity", "qty")
	quantity := float64(1)
	if quantityValue != nil {
		parsed, err := invoiceToolFloat64(quantityValue)
		if err != nil {
			return domain.InvoiceItemInput{}, fmt.Errorf("item %d quantity: %w", index+1, err)
		}
		quantity = parsed
	}

	unitPriceValue := firstInvoiceToolValue(itemMap, "unit_price", "price", "rate", "amount")
	unitPrice, err := invoiceToolFloat64(unitPriceValue)
	if err != nil {
		return domain.InvoiceItemInput{}, fmt.Errorf("item %d unit_price: %w", index+1, err)
	}

	taxRateValue := firstInvoiceToolValue(itemMap, "tax_rate", "tax")
	taxRate, err := invoiceToolFloat64(taxRateValue)
	if err != nil {
		return domain.InvoiceItemInput{}, fmt.Errorf("item %d tax_rate: %w", index+1, err)
	}

	discountPctValue := firstInvoiceToolValue(itemMap, "discount_pct", "discount_percent")
	discountPct, err := invoiceToolFloat64(discountPctValue)
	if err != nil {
		return domain.InvoiceItemInput{}, fmt.Errorf("item %d discount_pct: %w", index+1, err)
	}

	sortOrder := index
	sortOrderValue := itemMap["sort_order"]
	if sortOrderValue != nil {
		parsed, err := invoiceToolInt64(sortOrderValue)
		if err != nil {
			return domain.InvoiceItemInput{}, fmt.Errorf("item %d sort_order: %w", index+1, err)
		}
		sortOrder = int(parsed)
	}

	return domain.InvoiceItemInput{
		Description: description,
		Quantity:    quantity,
		Unit:        strings.TrimSpace(invoiceToolString(itemMap["unit"])),
		UnitPrice:   unitPrice,
		TaxRate:     taxRate,
		DiscountPct: discountPct,
		SortOrder:   sortOrder,
	}, nil
}

func validateInvoiceToolItems(items []domain.InvoiceItemInput) ([]domain.InvoiceItemInput, error) {
	for index, item := range items {
		if strings.TrimSpace(item.Description) == "" {
			return nil, fmt.Errorf("item %d description is required", index+1)
		}
		if item.Quantity <= 0 {
			return nil, fmt.Errorf("item %d quantity must be positive", index+1)
		}
		if item.UnitPrice < 0 {
			return nil, fmt.Errorf("item %d unit_price must be non-negative", index+1)
		}
		if item.TaxRate < 0 || item.TaxRate > 100 {
			return nil, fmt.Errorf("item %d tax_rate must be between 0 and 100", index+1)
		}
		if item.DiscountPct < 0 || item.DiscountPct > 100 {
			return nil, fmt.Errorf("item %d discount_pct must be between 0 and 100", index+1)
		}
	}
	return items, nil
}

func invoiceToolDate(value any) (string, error) {
	switch v := value.(type) {
	case nil:
		return "", nil
	case time.Time:
		if v.IsZero() {
			return "", nil
		}
		return v.Format("2006-01-02"), nil
	case string:
		raw := strings.TrimSpace(v)
		if raw == "" {
			return "", nil
		}
		for _, layout := range []string{"2006-01-02", time.RFC3339, time.RFC3339Nano} {
			parsed, err := time.Parse(layout, raw)
			if err == nil {
				return parsed.Format("2006-01-02"), nil
			}
		}
		return "", errors.New("must be YYYY-MM-DD or RFC3339")
	default:
		return "", fmt.Errorf("unsupported value %T", value)
	}
}

func normalizeListInvoicesToolStatus(value any) string {
	status := strings.ToLower(strings.TrimSpace(invoiceToolString(value)))
	if status == "" {
		return ""
	}
	status = strings.ReplaceAll(status, "-", "_")
	status = strings.ReplaceAll(status, " ", "_")
	if domain.IsValidStatus(status) {
		return status
	}
	return ""
}

func normalizeInvoiceToolDiscountType(value any) string {
	discountType := strings.ToLower(strings.TrimSpace(invoiceToolString(value)))
	if discountType == "" {
		return "fixed"
	}
	switch strings.ReplaceAll(discountType, " ", "_") {
	case "fixed", "amount", "flat":
		return "fixed"
	case "percentage", "percent", "pct":
		return "percentage"
	default:
		return ""
	}
}

func normalizePaymentToolMethod(value any) string {
	method := strings.ToLower(strings.TrimSpace(invoiceToolString(value)))
	if method == "" {
		return domain.PaymentMethodCash
	}
	method = strings.ReplaceAll(method, "-", "_")
	method = strings.ReplaceAll(method, " ", "_")
	switch method {
	case domain.PaymentMethodCash:
		return domain.PaymentMethodCash
	case domain.PaymentMethodCard, "credit_card", "debit_card":
		return domain.PaymentMethodCard
	case domain.PaymentMethodBankTransfer, "transfer", "bank", "wire", "wire_transfer":
		return domain.PaymentMethodBankTransfer
	case domain.PaymentMethodOther:
		return domain.PaymentMethodOther
	default:
		return ""
	}
}

func validInvoiceToolCurrency(currency string) bool {
	if len(currency) != 3 {
		return false
	}
	for _, char := range currency {
		if char < 'A' || char > 'Z' {
			return false
		}
	}
	return true
}

func invoiceToolPayload(invoice domain.Invoice) map[string]any {
	return map[string]any{
		"id":             invoice.ID,
		"invoice_number": invoice.InvoiceNumber,
		"client_name":    invoice.ClientName,
		"status":         invoice.Status,
		"issue_date":     invoice.IssueDate.Format("2006-01-02"),
		"due_date":       invoice.DueDate.Format("2006-01-02"),
		"paid_date":      formatInvoiceToolOptionalDate(invoice.PaidDate),
		"currency":       invoice.Currency,
		"total":          invoice.Total,
		"amount_paid":    invoice.AmountPaid,
		"amount_due":     invoice.AmountDue,
		"created_at":     invoice.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func formatInvoiceToolOptionalDate(value *time.Time) string {
	if value == nil || value.IsZero() {
		return ""
	}
	return value.Format("2006-01-02")
}

func firstInvoiceToolValue(input map[string]any, keys ...string) any {
	for _, key := range keys {
		value, ok := input[key]
		if !ok {
			continue
		}
		if raw, ok := value.(string); ok && strings.TrimSpace(raw) == "" {
			continue
		}
		if number, ok := value.(float64); ok && math.IsNaN(number) {
			continue
		}
		return value
	}
	return nil
}

func invoiceToolFloat64(value any) (float64, error) {
	switch v := value.(type) {
	case nil:
		return 0, nil
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case json.Number:
		return v.Float64()
	case string:
		raw := strings.TrimSpace(v)
		if raw == "" {
			return 0, nil
		}
		return strconv.ParseFloat(raw, 64)
	default:
		return 0, fmt.Errorf("unsupported value %T", value)
	}
}

func invoiceToolInt64(value any) (int64, error) {
	switch v := value.(type) {
	case nil:
		return 0, nil
	case int:
		return int64(v), nil
	case int64:
		return v, nil
	case int32:
		return int64(v), nil
	case float64:
		if math.Trunc(v) != v {
			return 0, errors.New("must be an integer")
		}
		return int64(v), nil
	case json.Number:
		return v.Int64()
	case string:
		raw := strings.TrimSpace(v)
		if raw == "" {
			return 0, nil
		}
		return strconv.ParseInt(raw, 10, 64)
	default:
		return 0, fmt.Errorf("unsupported value %T", value)
	}
}

func invoiceToolString(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case json.Number:
		return v.String()
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprint(v)
	}
}
