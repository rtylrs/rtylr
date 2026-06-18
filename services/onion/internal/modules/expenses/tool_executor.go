package expenses

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/expenses/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
)

type recordExpenseToolInput struct {
	CategoryID  *int64
	Amount      float64
	Currency    string
	Description string
	Vendor      string
	ReceiptURL  string
	ExpenseDate time.Time
}

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	if name != "finance.record_expense" {
		return nil, false, nil
	}
	output, err := m.executeRecordExpenseTool(ctx, input)
	return output, true, err
}

func (m *Module) executeRecordExpenseTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	parsed, err := parseRecordExpenseToolInput(input)
	if err != nil {
		return nil, err
	}

	id, err := m.service.Create(ctx, domain.ExpenseInput{
		OwnerID:     authUser.UserID,
		CategoryID:  parsed.CategoryID,
		Amount:      parsed.Amount,
		Currency:    parsed.Currency,
		Description: parsed.Description,
		Vendor:      parsed.Vendor,
		ReceiptURL:  parsed.ReceiptURL,
		ExpenseDate: parsed.ExpenseDate,
	})
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"expense": map[string]any{
			"id":           id,
			"category_id":  parsed.CategoryID,
			"amount":       parsed.Amount,
			"currency":     parsed.Currency,
			"description":  parsed.Description,
			"vendor":       parsed.Vendor,
			"receipt_url":  parsed.ReceiptURL,
			"expense_date": parsed.ExpenseDate.Format("2006-01-02"),
			"status":       "pending",
		},
	}, nil
}

func parseRecordExpenseToolInput(input map[string]any) (recordExpenseToolInput, error) {
	amount, err := expenseToolFloat64(input["amount"])
	if err != nil {
		return recordExpenseToolInput{}, fmt.Errorf("invalid amount: %w", err)
	}

	expenseDate, err := expenseToolDate(input["expense_date"])
	if err != nil {
		return recordExpenseToolInput{}, fmt.Errorf("invalid expense_date: %w", err)
	}

	categoryID, err := expenseToolOptionalInt64(input["category_id"])
	if err != nil {
		return recordExpenseToolInput{}, fmt.Errorf("invalid category_id: %w", err)
	}

	parsed := recordExpenseToolInput{
		CategoryID:  categoryID,
		Amount:      amount,
		Currency:    strings.ToUpper(strings.TrimSpace(expenseToolString(input["currency"]))),
		Description: strings.TrimSpace(expenseToolString(input["description"])),
		Vendor:      strings.TrimSpace(expenseToolString(input["vendor"])),
		ReceiptURL:  strings.TrimSpace(expenseToolString(input["receipt_url"])),
		ExpenseDate: expenseDate,
	}
	if err := domain.ValidatExpenseInput(parsed.Amount, parsed.Currency, parsed.Description, parsed.ExpenseDate); err != nil {
		return recordExpenseToolInput{}, err
	}
	return parsed, nil
}

func expenseToolOptionalInt64(value any) (*int64, error) {
	id, err := expenseToolInt64(value)
	if err != nil {
		return nil, err
	}
	if id < 0 {
		return nil, errors.New("must be non-negative")
	}
	if id == 0 {
		return nil, nil
	}
	return &id, nil
}

func expenseToolInt64(value any) (int64, error) {
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

func expenseToolFloat64(value any) (float64, error) {
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

func expenseToolDate(value any) (time.Time, error) {
	switch v := value.(type) {
	case nil:
		return time.Now(), nil
	case time.Time:
		if v.IsZero() {
			return time.Now(), nil
		}
		return v, nil
	case string:
		raw := strings.TrimSpace(v)
		if raw == "" {
			return time.Now(), nil
		}
		for _, layout := range []string{"2006-01-02", time.RFC3339, time.RFC3339Nano} {
			parsed, err := time.Parse(layout, raw)
			if err == nil {
				return parsed, nil
			}
		}
		return time.Time{}, errors.New("must be YYYY-MM-DD or RFC3339")
	default:
		return time.Time{}, fmt.Errorf("unsupported value %T", value)
	}
}

func expenseToolString(value any) string {
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
