package quotes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/quotes/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	required, dispatched := quoteToolPermissions[name]
	if !dispatched {
		return nil, false, nil
	}
	if err := requireQuoteToolPermission(ctx, required); err != nil {
		return nil, true, err
	}
	output, err := m.executeCreateQuoteTool(ctx, input)
	return output, true, err
}

func (m *Module) executeCreateQuoteTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	parsed, err := parseCreateQuoteToolInput(input)
	if err != nil {
		return nil, err
	}

	id, err := m.service.Create(ctx, businessID, parsed)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"quote": map[string]any{
			"id":             id,
			"customer_name":  parsed.CustomerName,
			"customer_email": parsed.CustomerEmail,
			"status":         parsed.Status,
			"items":          parsed.Items,
			"subtotal":       parsed.Subtotal,
			"tax":            parsed.Tax,
			"discount":       parsed.Discount,
			"total":          parsed.Total,
			"valid_until":    parsed.ValidUntil,
			"notes":          parsed.Notes,
		},
	}, nil
}

func parseCreateQuoteToolInput(input map[string]any) (domain.QuoteInput, error) {
	customerName := strings.TrimSpace(quoteToolString(input["customer_name"]))
	if customerName == "" {
		customerName = strings.TrimSpace(quoteToolString(input["client_name"]))
	}
	customerEmail := strings.TrimSpace(quoteToolString(input["customer_email"]))
	if customerEmail == "" {
		customerEmail = strings.TrimSpace(quoteToolString(input["client_email"]))
	}

	status := normalizeQuoteStatus(input["status"])
	if status == "" {
		return domain.QuoteInput{}, errors.New("status must be one of draft, sent, accepted, declined, or expired")
	}

	itemsValue := input["items"]
	if itemsValue == nil {
		itemsValue = input["line_items"]
	}
	items, err := quoteToolItems(itemsValue)
	if err != nil {
		return domain.QuoteInput{}, fmt.Errorf("invalid items: %w", err)
	}

	subtotal, err := quoteToolFloat64(input["subtotal"])
	if err != nil {
		return domain.QuoteInput{}, fmt.Errorf("invalid subtotal: %w", err)
	}
	tax, err := quoteToolFloat64(input["tax"])
	if err != nil {
		return domain.QuoteInput{}, fmt.Errorf("invalid tax: %w", err)
	}
	discount, err := quoteToolFloat64(input["discount"])
	if err != nil {
		return domain.QuoteInput{}, fmt.Errorf("invalid discount: %w", err)
	}

	totalValue, totalProvided := nonEmptyQuoteToolValue(input["total"])
	total := subtotal - discount + tax
	if totalProvided {
		total, err = quoteToolFloat64(totalValue)
		if err != nil {
			return domain.QuoteInput{}, fmt.Errorf("invalid total: %w", err)
		}
	}

	validUntil, err := quoteToolDate(input["valid_until"])
	if err != nil {
		return domain.QuoteInput{}, fmt.Errorf("invalid valid_until: %w", err)
	}

	parsed := domain.QuoteInput{
		CustomerName:  customerName,
		CustomerEmail: customerEmail,
		Status:        status,
		Items:         items,
		Subtotal:      subtotal,
		Tax:           tax,
		Discount:      discount,
		Total:         total,
		ValidUntil:    validUntil,
		Notes:         strings.TrimSpace(quoteToolString(input["notes"])),
	}
	if err := domain.ValidateQuoteInput(parsed); err != nil {
		return domain.QuoteInput{}, err
	}
	return parsed, nil
}

func normalizeQuoteStatus(value any) string {
	status := strings.ToLower(strings.TrimSpace(quoteToolString(value)))
	if status == "" {
		return "draft"
	}
	status = strings.ReplaceAll(status, " ", "_")
	if _, ok := domain.ValidQuoteStatuses[status]; ok {
		return status
	}
	return ""
}

func quoteToolItems(value any) (string, error) {
	switch v := value.(type) {
	case nil:
		return "", nil
	case string:
		raw := strings.TrimSpace(v)
		if raw == "" {
			return "", nil
		}
		if json.Valid([]byte(raw)) {
			return raw, nil
		}
		return marshalQuoteToolItems([]map[string]string{{"description": raw}})
	case json.Number:
		return marshalQuoteToolItems([]map[string]string{{"description": v.String()}})
	case fmt.Stringer:
		return quoteToolItems(v.String())
	default:
		return marshalQuoteToolItems(v)
	}
}

func marshalQuoteToolItems(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func quoteToolDate(value any) (string, error) {
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

func quoteToolFloat64(value any) (float64, error) {
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

func quoteToolString(value any) string {
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

func nonEmptyQuoteToolValue(value any) (any, bool) {
	if value == nil {
		return nil, false
	}
	if raw, ok := value.(string); ok && strings.TrimSpace(raw) == "" {
		return nil, false
	}
	if number, ok := value.(float64); ok && math.IsNaN(number) {
		return nil, false
	}
	return value, true
}
