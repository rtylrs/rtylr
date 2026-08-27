package hr_time

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/hr_time/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/tenant"
)

const (
	listHRTimeEntriesToolName = "hr.list_time_entries"
	listHRTimesheetsToolName  = "hr.list_timesheets"
	getHRTimesheetToolName    = "hr.get_timesheet"
	defaultHRTimeToolLimit    = int64(20)
	maxHRTimeToolLimit        = int64(50)
)

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	switch name {
	case listHRTimeEntriesToolName:
		output, err := m.executeListTimeEntriesTool(ctx, input)
		return output, true, err
	case listHRTimesheetsToolName:
		output, err := m.executeListTimesheetsTool(ctx, input)
		return output, true, err
	case getHRTimesheetToolName:
		output, err := m.executeGetTimesheetTool(ctx, input)
		return output, true, err
	default:
		return nil, false, nil
	}
}

func (m *Module) executeListTimeEntriesTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	businessID, err := hrTimeToolBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	limit, offset, err := hrTimeToolPagination(input)
	if err != nil {
		return nil, err
	}
	employeeID, err := hrTimeToolNonNegativeInt64(firstHRTimeToolValue(input, "employee_id", "employee"))
	if err != nil {
		return nil, fmt.Errorf("invalid employee_id: %w", err)
	}
	periodStart, periodEnd, err := hrTimeToolPeriod(input)
	if err != nil {
		return nil, err
	}

	// The filters are optional now (tryrtylr/rtylr#855), but this tool has always
	// required an employee and a period, so it keeps passing all three.
	entries, total, err := m.service.ListTimeEntries(ctx, businessID, domain.ListTimeEntriesParams{
		EmployeeID: &employeeID,
		FromDate:   &periodStart,
		ToDate:     &periodEnd,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return nil, err
	}

	payloads := make([]map[string]any, 0, len(entries))
	for _, entry := range entries {
		payloads = append(payloads, hrTimeEntryToolPayload(entry))
	}

	return map[string]any{
		"entries":  payloads,
		"total":    total,
		"returned": len(payloads),
		"limit":    limit,
		"offset":   offset,
		"filters": map[string]any{
			"employee_id":  employeeID,
			"period_start": periodStart,
			"period_end":   periodEnd,
		},
	}, nil
}

func (m *Module) executeListTimesheetsTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	businessID, err := hrTimeToolBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	limit, offset, err := hrTimeToolPagination(input)
	if err != nil {
		return nil, err
	}
	employeeID, err := hrTimeToolNonNegativeInt64(firstHRTimeToolValue(input, "employee_id", "employee"))
	if err != nil {
		return nil, fmt.Errorf("invalid employee_id: %w", err)
	}

	timesheets, total, err := m.service.ListTimesheets(ctx, businessID, domain.ListTimesheetsParams{
		EmployeeID: employeeID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return nil, err
	}

	payloads := make([]map[string]any, 0, len(timesheets))
	for _, timesheet := range timesheets {
		payloads = append(payloads, hrTimesheetToolPayload(timesheet))
	}

	return map[string]any{
		"timesheets": payloads,
		"total":      total,
		"returned":   len(payloads),
		"limit":      limit,
		"offset":     offset,
		"filters": map[string]any{
			"employee_id": employeeID,
		},
	}, nil
}

func (m *Module) executeGetTimesheetTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	businessID, err := hrTimeToolBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	timesheetID, err := hrTimeToolNonNegativeInt64(firstHRTimeToolValue(input, "timesheet_id", "id"))
	if err != nil {
		return nil, fmt.Errorf("invalid timesheet_id: %w", err)
	}

	var timesheet *domain.TimesheetRecord
	if timesheetID > 0 {
		timesheet, err = m.service.GetTimesheet(ctx, businessID, timesheetID)
	} else {
		employeeID, employeeErr := hrTimeToolPositiveInt64(firstHRTimeToolValue(input, "employee_id", "employee"))
		if employeeErr != nil {
			return nil, errors.New("timesheet_id or employee_id with period_start is required")
		}
		periodStart, periodErr := hrTimeToolDate(firstHRTimeToolValue(input, "period_start", "date_from", "from"), "period_start")
		if periodErr != nil {
			return nil, periodErr
		}
		if periodStart == "" {
			return nil, errors.New("period_start is required when timesheet_id is omitted")
		}
		timesheet, err = m.service.GetTimesheetForPeriod(ctx, businessID, employeeID, periodStart)
	}
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"timesheet": hrTimesheetToolPayloadPtr(timesheet),
	}, nil
}

func hrTimeToolBusinessID(ctx context.Context) (int64, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return 0, errors.New("authentication is required")
	}
	businessID, err := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if err != nil {
		return 0, err
	}
	return businessID, nil
}

func hrTimeToolPagination(input map[string]any) (int32, int32, error) {
	limit, err := hrTimeToolNonNegativeInt64(firstHRTimeToolValue(input, "limit", "page_size"))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid limit: %w", err)
	}
	if limit == 0 {
		limit = defaultHRTimeToolLimit
	}
	if limit > maxHRTimeToolLimit {
		limit = maxHRTimeToolLimit
	}

	offset, err := hrTimeToolNonNegativeInt64(firstHRTimeToolValue(input, "offset"))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid offset: %w", err)
	}
	if offset > math.MaxInt32 {
		return 0, 0, errors.New("offset is too large")
	}
	return int32(limit), int32(offset), nil // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
}

func hrTimeToolPeriod(input map[string]any) (string, string, error) {
	if dateValue := firstHRTimeToolValue(input, "date", "day"); dateValue != nil {
		date, err := hrTimeToolDate(dateValue, "date")
		if err != nil {
			return "", "", err
		}
		if date != "" {
			start, err := time.Parse("2006-01-02", date)
			if err != nil {
				return "", "", errors.New("date must be in YYYY-MM-DD format")
			}
			return date, start.AddDate(0, 0, 6).Format("2006-01-02"), nil
		}
	}

	periodStart, err := hrTimeToolDate(firstHRTimeToolValue(input, "period_start", "date_from", "from"), "period_start")
	if err != nil {
		return "", "", err
	}
	periodEnd, err := hrTimeToolDate(firstHRTimeToolValue(input, "period_end", "date_to", "to"), "period_end")
	if err != nil {
		return "", "", err
	}
	return periodStart, periodEnd, nil
}

func firstHRTimeToolValue(input map[string]any, keys ...string) any {
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

func hrTimeToolDate(value any, field string) (string, error) {
	date := strings.TrimSpace(hrTimeToolString(value))
	if date == "" {
		return "", nil
	}
	if strings.EqualFold(date, "today") {
		return time.Now().UTC().Format("2006-01-02"), nil
	}
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return "", fmt.Errorf("%s must be in YYYY-MM-DD format", field)
	}
	return date, nil
}

func hrTimeToolPositiveInt64(value any) (int64, error) {
	parsed, err := hrTimeToolNonNegativeInt64(value)
	if err != nil {
		return 0, err
	}
	if parsed <= 0 {
		return 0, errors.New("value is required")
	}
	return parsed, nil
}

func hrTimeToolNonNegativeInt64(value any) (int64, error) {
	parsed, err := hrTimeToolInt64(value)
	if err != nil {
		return 0, err
	}
	if parsed < 0 {
		return 0, errors.New("value must be non-negative")
	}
	return parsed, nil
}

func hrTimeToolInt64(value any) (int64, error) {
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

func hrTimeToolString(value any) string {
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

func hrTimeEntryToolPayload(entry domain.TimeEntryRecord) map[string]any {
	return map[string]any{
		"id":                  entry.ID,
		"employee_id":         entry.EmployeeID,
		"timesheet_id":        entry.TimesheetID,
		"project_id":          entry.ProjectID,
		"task_id":             entry.TaskID,
		"category_id":         entry.CategoryID,
		"date":                entry.Date,
		"duration_minutes":    entry.DurationMinutes,
		"start_time":          entry.StartTime,
		"end_time":            entry.EndTime,
		"description":         entry.Description,
		"is_billable":         entry.IsBillable,
		"status":              entry.Status,
		"rejection_reason":    entry.RejectionReason,
		"approved_by":         entry.ApprovedBy,
		"source":              entry.Source,
		"billable_invoice_id": entry.BillableInvoiceID,
	}
}

func hrTimesheetToolPayloadPtr(timesheet *domain.TimesheetRecord) any {
	if timesheet == nil {
		return nil
	}
	return hrTimesheetToolPayload(*timesheet)
}

func hrTimesheetToolPayload(timesheet domain.TimesheetRecord) map[string]any {
	return map[string]any{
		"id":                     timesheet.ID,
		"employee_id":            timesheet.EmployeeID,
		"period_start":           timesheet.PeriodStart,
		"period_end":             timesheet.PeriodEnd,
		"status":                 timesheet.Status,
		"current_approval_level": timesheet.CurrentApprovalLevel,
		"submitted_at":           timesheet.SubmittedAt,
		"total_minutes":          timesheet.TotalMinutes,
		"rejection_reason":       timesheet.RejectionReason,
	}
}
