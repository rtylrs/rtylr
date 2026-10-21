package invoices

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/finance/invoices/internal/domain"
)

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

func invoiceDetailToolPayload(invoice domain.Invoice) map[string]any {
	items := make([]map[string]any, 0, len(invoice.Items))
	for _, item := range invoice.Items {
		items = append(items, map[string]any{
			"id":           item.ID,
			"invoice_id":   item.InvoiceID,
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

	payments := make([]map[string]any, 0, len(invoice.Payments))
	for _, payment := range invoice.Payments {
		payments = append(payments, map[string]any{
			"id":           payment.ID,
			"invoice_id":   payment.InvoiceID,
			"amount":       payment.Amount,
			"currency":     payment.Currency,
			"payment_date": payment.PaymentDate.Format("2006-01-02"),
			"method":       payment.Method,
			"reference":    payment.Reference,
			"notes":        payment.Notes,
			"created_at":   payment.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return map[string]any{
		"id":              invoice.ID,
		"invoice_number":  invoice.InvoiceNumber,
		"client_name":     invoice.ClientName,
		"client_email":    invoice.ClientEmail,
		"client_address":  invoice.ClientAddress,
		"client_tax_id":   invoice.ClientTaxID,
		"status":          invoice.Status,
		"issue_date":      invoice.IssueDate.Format("2006-01-02"),
		"due_date":        invoice.DueDate.Format("2006-01-02"),
		"paid_date":       formatInvoiceToolOptionalDate(invoice.PaidDate),
		"currency":        invoice.Currency,
		"subtotal":        invoice.Subtotal,
		"discount_type":   invoice.DiscountType,
		"discount_value":  invoice.DiscountValue,
		"discount_amount": invoice.DiscountAmount,
		"tax_rate":        invoice.TaxRate,
		"tax_amount":      invoice.TaxAmount,
		"total":           invoice.Total,
		"amount_paid":     invoice.AmountPaid,
		"amount_due":      invoice.AmountDue,
		"notes":           invoice.Notes,
		"terms":           invoice.Terms,
		"created_at":      invoice.CreatedAt.Format("2006-01-02T15:04:05Z"),
		"updated_at":      invoice.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		"items":           items,
		"payments":        payments,
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
