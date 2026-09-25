package bills

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/bills/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/money"
	"github.com/voxire/onion/internal/platform/tenant"
)

const (
	listBillsToolName  = "finance.list_bills"
	recordBillToolName = "finance.record_bill"
)

type listBillsToolInput struct {
	Status string
}

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	switch name {
	case listBillsToolName:
		output, err := m.executeListBillsTool(ctx, input)
		return output, true, err
	case recordBillToolName:
		output, err := m.executeRecordBillTool(ctx, input)
		return output, true, err
	default:
		return nil, false, nil
	}
}

func (m *Module) executeListBillsTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, bizErr
	}

	parsed, err := parseListBillsToolInput(input)
	if err != nil {
		return nil, err
	}

	bills, err := m.service.List(ctx, businessID)
	if err != nil {
		return nil, err
	}

	payloads := make([]map[string]any, 0, len(bills))
	for _, bill := range bills {
		if parsed.Status != "" && bill.Status != parsed.Status {
			continue
		}
		payloads = append(payloads, billToolPayload(bill))
	}

	return map[string]any{
		"bills":    payloads,
		"returned": len(payloads),
		"filters": map[string]any{
			"status": parsed.Status,
		},
	}, nil
}

func (m *Module) executeRecordBillTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, bizErr
	}

	parsed, err := parseRecordBillToolInput(input)
	if err != nil {
		return nil, err
	}

	id, err := m.service.Create(ctx, businessID, parsed)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"bill": map[string]any{
			"id":          id,
			"supplier_id": parsed.SupplierID,
			"amount":      parsed.Amount,
			"currency":    parsed.Currency,
			"due_date":    parsed.DueDate,
			"status":      parsed.Status,
			"description": parsed.Description,
		},
	}, nil
}

func parseRecordBillToolInput(input map[string]any) (domain.BillInput, error) {
	supplierID, err := billToolInt64(input["supplier_id"])
	if err != nil {
		return domain.BillInput{}, fmt.Errorf("invalid supplier_id: %w", err)
	}
	if supplierID <= 0 {
		return domain.BillInput{}, errors.New("supplier_id is required")
	}

	amountValue, err := billToolFloat64(input["amount"])
	if err != nil {
		return domain.BillInput{}, fmt.Errorf("invalid amount: %w", err)
	}
	amount, err := money.FromFloat64(amountValue)
	if err != nil {
		return domain.BillInput{}, fmt.Errorf("invalid amount: %w", err)
	}

	dueDate, err := billToolDate(input["due_date"])
	if err != nil {
		return domain.BillInput{}, fmt.Errorf("invalid due_date: %w", err)
	}

	status := normalizeBillStatus(input["status"])
	if status == "" {
		return domain.BillInput{}, errors.New("status must be one of pending, paid, or overdue")
	}

	description := strings.TrimSpace(billToolString(input["description"]))
	if description == "" {
		description = strings.TrimSpace(billToolString(input["notes"]))
	}

	parsed := domain.BillInput{
		SupplierID:  supplierID,
		Amount:      amount,
		Currency:    strings.ToUpper(strings.TrimSpace(billToolString(input["currency"]))),
		DueDate:     dueDate,
		Status:      status,
		Description: description,
	}
	if err := domain.ValidateBillInput(parsed); err != nil {
		return domain.BillInput{}, err
	}
	return parsed, nil
}

func parseListBillsToolInput(input map[string]any) (listBillsToolInput, error) {
	status := normalizeBillStatus(input["status"])
	if strings.TrimSpace(billToolString(input["status"])) != "" && status == "" {
		return listBillsToolInput{}, errors.New("status must be one of pending, paid, or overdue")
	}
	return listBillsToolInput{Status: status}, nil
}

func normalizeBillStatus(value any) string {
	status := strings.ToLower(strings.TrimSpace(billToolString(value)))
	if status == "" {
		return "pending"
	}
	status = strings.ReplaceAll(status, " ", "_")
	if _, ok := domain.ValidBillStatuses[status]; ok {
		return status
	}
	return ""
}

func billToolPayload(bill domain.Bill) map[string]any {
	return map[string]any{
		"id":          bill.ID,
		"supplier_id": bill.SupplierID,
		"amount":      bill.Amount,
		"currency":    bill.Currency,
		"due_date":    billToolDateOnly(bill.DueDate),
		"status":      bill.Status,
		"description": bill.Description,
		"created_at":  billToolTime(bill.CreatedAt),
		"updated_at":  billToolTime(bill.UpdatedAt),
	}
}

func billToolDate(value any) (string, error) {
	switch v := value.(type) {
	case nil:
		return "", errors.New("is required")
	case time.Time:
		if v.IsZero() {
			return "", errors.New("is required")
		}
		return v.Format("2006-01-02"), nil
	case string:
		raw := strings.TrimSpace(v)
		if raw == "" {
			return "", errors.New("is required")
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

func billToolInt64(value any) (int64, error) {
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

func billToolFloat64(value any) (float64, error) {
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

func billToolString(value any) string {
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

func billToolDateOnly(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format("2006-01-02")
}

func billToolTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}
