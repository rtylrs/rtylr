package hr_payroll

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/voxire/onion/internal/modules/hr_payroll/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
)

const (
	getHRPayrollConfigToolName = "hr.get_payroll_config"
	listHRPayrollRunsToolName  = "hr.list_payroll_runs"
	getHRPayrollRunToolName    = "hr.get_payroll_run"
	defaultHRPayrollToolLimit  = int64(20)
	maxHRPayrollToolLimit      = int64(50)
)

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	switch name {
	case getHRPayrollConfigToolName:
		output, err := m.executeGetPayrollConfigTool(ctx)
		return output, true, err
	case listHRPayrollRunsToolName:
		output, err := m.executeListPayrollRunsTool(ctx, input)
		return output, true, err
	case getHRPayrollRunToolName:
		output, err := m.executeGetPayrollRunTool(ctx, input)
		return output, true, err
	default:
		return nil, false, nil
	}
}

func (m *Module) executeGetPayrollConfigTool(ctx context.Context) (map[string]any, error) {
	ownerID, err := hrPayrollToolOwnerID(ctx)
	if err != nil {
		return nil, err
	}

	config, err := m.service.GetPayrollConfig(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return map[string]any{
			"configured": false,
			"config":     nil,
		}, nil
	}

	return map[string]any{
		"configured": true,
		"config":     hrPayrollConfigToolPayload(*config),
	}, nil
}

func (m *Module) executeListPayrollRunsTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	ownerID, err := hrPayrollToolOwnerID(ctx)
	if err != nil {
		return nil, err
	}
	limit, offset, err := hrPayrollToolPagination(input)
	if err != nil {
		return nil, err
	}

	runs, total, err := m.service.ListPayrollRuns(ctx, ownerID, domain.ListPayrollRunsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	payloads := make([]map[string]any, 0, len(runs))
	for _, run := range runs {
		payloads = append(payloads, hrPayrollRunToolPayload(run))
	}

	return map[string]any{
		"runs":     payloads,
		"total":    total,
		"returned": len(payloads),
		"limit":    limit,
		"offset":   offset,
	}, nil
}

func (m *Module) executeGetPayrollRunTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	ownerID, err := hrPayrollToolOwnerID(ctx)
	if err != nil {
		return nil, err
	}
	runID, err := hrPayrollToolPositiveInt64(firstHRPayrollToolValue(input, "run_id", "id"))
	if err != nil {
		return nil, fmt.Errorf("invalid run_id: %w", err)
	}

	run, entries, err := m.service.GetPayrollRun(ctx, ownerID, runID)
	if err != nil {
		return nil, err
	}

	entryPayloads := make([]map[string]any, 0, len(entries))
	for _, entry := range entries {
		entryPayloads = append(entryPayloads, hrPayrollEntryToolPayload(entry))
	}

	return map[string]any{
		"run":      hrPayrollRunToolPayloadPtr(run),
		"entries":  entryPayloads,
		"returned": len(entryPayloads),
	}, nil
}

func hrPayrollToolOwnerID(ctx context.Context) (int64, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return 0, errors.New("authentication is required")
	}
	return authUser.UserID, nil
}

func hrPayrollToolPagination(input map[string]any) (int32, int32, error) {
	limit, err := hrPayrollToolNonNegativeInt64(firstHRPayrollToolValue(input, "limit", "page_size"))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid limit: %w", err)
	}
	if limit == 0 {
		limit = defaultHRPayrollToolLimit
	}
	if limit > maxHRPayrollToolLimit {
		limit = maxHRPayrollToolLimit
	}

	offset, err := hrPayrollToolNonNegativeInt64(firstHRPayrollToolValue(input, "offset"))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid offset: %w", err)
	}
	if offset > math.MaxInt32 {
		return 0, 0, errors.New("offset is too large")
	}
	return int32(limit), int32(offset), nil // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
}

func firstHRPayrollToolValue(input map[string]any, keys ...string) any {
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

func hrPayrollToolPositiveInt64(value any) (int64, error) {
	parsed, err := hrPayrollToolNonNegativeInt64(value)
	if err != nil {
		return 0, err
	}
	if parsed <= 0 {
		return 0, errors.New("value is required")
	}
	return parsed, nil
}

func hrPayrollToolNonNegativeInt64(value any) (int64, error) {
	parsed, err := hrPayrollToolInt64(value)
	if err != nil {
		return 0, err
	}
	if parsed < 0 {
		return 0, errors.New("value must be non-negative")
	}
	return parsed, nil
}

func hrPayrollToolInt64(value any) (int64, error) {
	switch v := value.(type) {
	case nil:
		return 0, nil
	case int:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case float64:
		if math.Trunc(v) != v {
			return 0, errors.New("value must be an integer")
		}
		return int64(v), nil
	case json.Number:
		return v.Int64()
	case string:
		s := strings.TrimSpace(v)
		if s == "" {
			return 0, nil
		}
		return strconv.ParseInt(s, 10, 64)
	default:
		return 0, fmt.Errorf("unsupported number type %T", value)
	}
}

func hrPayrollConfigToolPayload(config domain.PayrollConfigRecord) map[string]any {
	return map[string]any{
		"id":               config.ID,
		"period_type":      config.PeriodType,
		"pay_day_of_month": config.PayDayOfMonth,
		"overtime_enabled": config.OvertimeEnabled,
	}
}

func hrPayrollRunToolPayloadPtr(run *domain.PayrollRunRecord) any {
	if run == nil {
		return nil
	}
	return hrPayrollRunToolPayload(*run)
}

// hrPayrollRunToolPayload builds the Caret tool-call payload. The tool
// contract carries these amounts as float64 dollars, so the internally
// stored cents are converted at this boundary only.
func hrPayrollRunToolPayload(run domain.PayrollRunRecord) map[string]any {
	return map[string]any{
		"id":           run.ID,
		"period_start": run.PeriodStart,
		"period_end":   run.PeriodEnd,
		"status":       run.Status,
		"created_by":   run.CreatedBy,
		"approved_by":  run.ApprovedBy,
		"exported_at":  run.ExportedAt,
		"total_gross":  domain.CentsToDollars(run.TotalGrossCents),
		"total_net":    domain.CentsToDollars(run.TotalNetCents),
	}
}

// hrPayrollEntryToolPayload builds the Caret tool-call payload. Same wire
// boundary conversion as hrPayrollRunToolPayload above.
func hrPayrollEntryToolPayload(entry domain.PayrollEntryRecord) map[string]any {
	return map[string]any{
		"id":             entry.ID,
		"employee_id":    entry.EmployeeID,
		"hours_worked":   entry.HoursWorked,
		"overtime_hours": entry.OvertimeHours,
		"base_pay":       domain.CentsToDollars(entry.BasePayCents),
		"overtime_pay":   domain.CentsToDollars(entry.OvertimePayCents),
		"bonuses":        domain.CentsToDollars(entry.BonusesCents),
		"deductions":     domain.CentsToDollars(entry.DeductionsCents),
		"net_pay":        domain.CentsToDollars(entry.NetPayCents),
	}
}
