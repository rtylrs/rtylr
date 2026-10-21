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

	"github.com/voxire/onion/internal/modules/finance/expenses/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/tenant"
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

const (
	listExpensesToolName          = "finance.list_expenses"
	getExpenseToolName            = "finance.get_expense"
	listExpenseCategoriesToolName = "finance.list_expense_categories"
	recordExpenseToolName         = "finance.record_expense"
)

type getExpenseToolInput struct {
	ExpenseID int64
}

type listExpensesToolInput struct {
	Status string
}

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	required, dispatched := expenseToolPermissions[name]
	if !dispatched {
		return nil, false, nil
	}
	if err := requireExpenseToolPermission(ctx, required); err != nil {
		return nil, true, err
	}
	switch name {
	case listExpensesToolName:
		output, err := m.executeListExpensesTool(ctx, input)
		return output, true, err
	case getExpenseToolName:
		output, err := m.executeGetExpenseTool(ctx, input)
		return output, true, err
	case listExpenseCategoriesToolName:
		output, err := m.executeListExpenseCategoriesTool(ctx)
		return output, true, err
	case recordExpenseToolName:
		output, err := m.executeRecordExpenseTool(ctx, input)
		return output, true, err
	default:
		return nil, true, fmt.Errorf("expense tool %q has no executor", name)
	}
}

func (m *Module) executeGetExpenseTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	parsed, err := parseGetExpenseToolInput(input)
	if err != nil {
		return nil, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, bizErr
	}

	expense, err := m.service.GetByID(ctx, businessID, parsed.ExpenseID)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"expense": expenseToolPayload(expense),
	}, nil
}

func (m *Module) executeListExpensesTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	parsed, err := parseListExpensesToolInput(input)
	if err != nil {
		return nil, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, bizErr
	}

	expenses, err := m.service.List(ctx, businessID)
	if err != nil {
		return nil, err
	}

	payloads := make([]map[string]any, 0, len(expenses))
	for _, expense := range expenses {
		if parsed.Status != "" && expense.Status != parsed.Status {
			continue
		}
		payloads = append(payloads, expenseToolPayload(expense))
	}

	return map[string]any{
		"expenses": payloads,
		"returned": len(payloads),
		"filters": map[string]any{
			"status": parsed.Status,
		},
	}, nil
}

func (m *Module) executeListExpenseCategoriesTool(ctx context.Context) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, bizErr
	}

	categories, err := m.service.ListCategories(ctx, businessID)
	if err != nil {
		return nil, err
	}

	payloads := make([]map[string]any, 0, len(categories))
	for _, category := range categories {
		payloads = append(payloads, expenseCategoryToolPayload(category))
	}

	return map[string]any{
		"categories": payloads,
		"returned":   len(payloads),
	}, nil
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

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, bizErr
	}

	id, err := m.service.Create(ctx, domain.ExpenseInput{
		OwnerID:     authUser.UserID,
		BusinessID:  businessID,
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

func parseGetExpenseToolInput(input map[string]any) (getExpenseToolInput, error) {
	expenseID, err := expenseIDToolValue(firstExpenseToolValue(input, "expense_id", "expense", "id"))
	if err != nil {
		return getExpenseToolInput{}, fmt.Errorf("invalid expense_id: %w", err)
	}
	if expenseID <= 0 {
		return getExpenseToolInput{}, errors.New("expense_id is required")
	}
	return getExpenseToolInput{ExpenseID: expenseID}, nil
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

func parseListExpensesToolInput(input map[string]any) (listExpensesToolInput, error) {
	status := normalizeExpenseStatus(input["status"])
	if strings.TrimSpace(expenseToolString(input["status"])) != "" && status == "" {
		return listExpensesToolInput{}, errors.New("status must be one of pending, approved, or rejected")
	}
	return listExpensesToolInput{Status: status}, nil
}

func firstExpenseToolValue(input map[string]any, keys ...string) any {
	for _, key := range keys {
		value, ok := input[key]
		if !ok || value == nil {
			continue
		}
		if s, ok := value.(string); ok && strings.TrimSpace(s) == "" {
			continue
		}
		return value
	}
	return nil
}

func normalizeExpenseStatus(value any) string {
	status := strings.ToLower(strings.TrimSpace(expenseToolString(value)))
	if status == "" {
		return ""
	}
	status = strings.ReplaceAll(status, " ", "_")
	if _, ok := domain.ValidExpenseStatuses[status]; ok {
		return status
	}
	return ""
}

func expenseToolPayload(expense domain.Expense) map[string]any {
	return map[string]any{
		"id":           expense.ID,
		"category_id":  expense.CategoryID,
		"amount":       expense.Amount,
		"currency":     expense.Currency,
		"description":  expense.Description,
		"vendor":       expense.Vendor,
		"receipt_url":  expense.ReceiptURL,
		"expense_date": expenseToolDateOnly(expense.ExpenseDate),
		"status":       expense.Status,
		"created_at":   expenseToolTime(expense.CreatedAt),
		"updated_at":   expenseToolTime(expense.UpdatedAt),
	}
}

func expenseCategoryToolPayload(category domain.ExpenseCategory) map[string]any {
	return map[string]any{
		"id":         category.ID,
		"name":       category.Name,
		"color":      category.Color,
		"created_at": expenseToolTime(category.CreatedAt),
	}
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

func expenseIDToolValue(value any) (int64, error) {
	switch v := value.(type) {
	case string:
		raw := strings.TrimSpace(v)
		raw = strings.TrimPrefix(raw, "#")
		raw = strings.ToLower(raw)
		for _, prefix := range []string{"expense-", "expense_", "exp-", "exp_"} {
			raw = strings.TrimPrefix(raw, prefix)
		}
		return expenseToolInt64(raw)
	default:
		return expenseToolInt64(value)
	}
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

func expenseToolDateOnly(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format("2006-01-02")
}

func expenseToolTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}
