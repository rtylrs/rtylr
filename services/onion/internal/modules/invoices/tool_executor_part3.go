package invoices

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/invoices/internal/domain"
)

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

func invoiceToolItem(value any, _ map[string]any, index int) (domain.InvoiceItemInput, error) {
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
