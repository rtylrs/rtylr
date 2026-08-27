package hr_leave

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/hr_leave/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
)

const (
	listHRLeaveTypesToolName    = "hr.list_leave_types"
	listHRLeaveRequestsToolName = "hr.list_leave_requests"
	getHRLeaveBalancesToolName  = "hr.get_leave_balances"
	getHRLeaveCalendarToolName  = "hr.get_leave_calendar"
	defaultHRLeaveToolLimit     = int64(20)
	maxHRLeaveToolLimit         = int64(50)
)

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	switch name {
	case listHRLeaveTypesToolName:
		output, err := m.executeListLeaveTypesTool(ctx, input)
		return output, true, err
	case listHRLeaveRequestsToolName:
		output, err := m.executeListLeaveRequestsTool(ctx, input)
		return output, true, err
	case getHRLeaveBalancesToolName:
		output, err := m.executeGetLeaveBalancesTool(ctx, input)
		return output, true, err
	case getHRLeaveCalendarToolName:
		output, err := m.executeGetLeaveCalendarTool(ctx, input)
		return output, true, err
	default:
		return nil, false, nil
	}
}

func (m *Module) executeListLeaveTypesTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	ownerID, err := hrLeaveToolOwnerID(ctx)
	if err != nil {
		return nil, err
	}
	limit, offset, err := hrLeaveToolPagination(input)
	if err != nil {
		return nil, err
	}

	leaveTypes, total, err := m.service.ListLeaveTypes(ctx, ownerID, domain.ListLeaveTypesParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	payloads := make([]map[string]any, 0, len(leaveTypes))
	for _, leaveType := range leaveTypes {
		payloads = append(payloads, hrLeaveTypeToolPayload(leaveType))
	}

	return map[string]any{
		"leave_types": payloads,
		"total":       total,
		"returned":    len(payloads),
		"limit":       limit,
		"offset":      offset,
	}, nil
}

func (m *Module) executeListLeaveRequestsTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	ownerID, err := hrLeaveToolOwnerID(ctx)
	if err != nil {
		return nil, err
	}
	limit, offset, err := hrLeaveToolPagination(input)
	if err != nil {
		return nil, err
	}

	requests, total, err := m.service.ListLeaveRequests(ctx, ownerID, ownerID, domain.ListLeaveRequestsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	payloads := make([]map[string]any, 0, len(requests))
	for _, request := range requests {
		payloads = append(payloads, hrLeaveRequestToolPayload(request))
	}

	return map[string]any{
		"requests": payloads,
		"total":    total,
		"returned": len(payloads),
		"limit":    limit,
		"offset":   offset,
	}, nil
}

func (m *Module) executeGetLeaveBalancesTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	ownerID, err := hrLeaveToolOwnerID(ctx)
	if err != nil {
		return nil, err
	}
	employeeID, err := hrLeaveToolPositiveInt64(firstHRLeaveToolValue(input, "employee_id", "employee"))
	if err != nil {
		return nil, fmt.Errorf("invalid employee_id: %w", err)
	}
	year, err := hrLeaveToolInt32(firstHRLeaveToolValue(input, "year"))
	if err != nil {
		return nil, fmt.Errorf("invalid year: %w", err)
	}
	if year == 0 {
		year = int32(time.Now().UTC().Year()) // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	}

	balances, err := m.service.GetLeaveBalances(ctx, ownerID, ownerID, employeeID, year)
	if err != nil {
		return nil, err
	}

	payloads := make([]map[string]any, 0, len(balances))
	for _, balance := range balances {
		payloads = append(payloads, hrLeaveBalanceToolPayload(balance))
	}

	return map[string]any{
		"employee_id": employeeID,
		"year":        year,
		"balances":    payloads,
		"returned":    len(payloads),
	}, nil
}

func (m *Module) executeGetLeaveCalendarTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	ownerID, err := hrLeaveToolOwnerID(ctx)
	if err != nil {
		return nil, err
	}
	locationID, err := hrLeaveToolNonNegativeInt64(firstHRLeaveToolValue(input, "location_id", "location"))
	if err != nil {
		return nil, fmt.Errorf("invalid location_id: %w", err)
	}
	dateFrom, err := hrLeaveToolDate(firstHRLeaveToolValue(input, "date_from", "from"), "date_from")
	if err != nil {
		return nil, err
	}
	dateTo, err := hrLeaveToolDate(firstHRLeaveToolValue(input, "date_to", "to"), "date_to")
	if err != nil {
		return nil, err
	}
	if dateFrom == "" {
		dateFrom = time.Now().UTC().Format("2006-01-02")
	}
	if dateTo == "" {
		start, parseErr := time.Parse("2006-01-02", dateFrom)
		if parseErr != nil {
			return nil, errors.New("date_from must be in YYYY-MM-DD format")
		}
		dateTo = start.AddDate(0, 0, 30).Format("2006-01-02")
	}

	requests, err := m.service.GetLeaveCalendar(ctx, ownerID, domain.LeaveCalendarParams{
		LocationID: locationID,
		DateFrom:   dateFrom,
		DateTo:     dateTo,
	})
	if err != nil {
		return nil, err
	}

	payloads := make([]map[string]any, 0, len(requests))
	for _, request := range requests {
		payloads = append(payloads, hrLeaveRequestToolPayload(request))
	}

	return map[string]any{
		"requests": payloads,
		"returned": len(payloads),
		"filters": map[string]any{
			"location_id": locationID,
			"date_from":   dateFrom,
			"date_to":     dateTo,
		},
	}, nil
}

func hrLeaveToolOwnerID(ctx context.Context) (int64, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return 0, errors.New("authentication is required")
	}
	return authUser.UserID, nil
}

func hrLeaveToolPagination(input map[string]any) (int32, int32, error) {
	limit, err := hrLeaveToolNonNegativeInt64(firstHRLeaveToolValue(input, "limit", "page_size"))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid limit: %w", err)
	}
	if limit == 0 {
		limit = defaultHRLeaveToolLimit
	}
	if limit > maxHRLeaveToolLimit {
		limit = maxHRLeaveToolLimit
	}

	offset, err := hrLeaveToolNonNegativeInt64(firstHRLeaveToolValue(input, "offset"))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid offset: %w", err)
	}
	if offset > math.MaxInt32 {
		return 0, 0, errors.New("offset is too large")
	}

	return int32(limit), int32(offset), nil // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
}

func firstHRLeaveToolValue(input map[string]any, keys ...string) any {
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

func hrLeaveToolDate(value any, field string) (string, error) {
	date := strings.TrimSpace(hrLeaveToolString(value))
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

func hrLeaveToolPositiveInt64(value any) (int64, error) {
	parsed, err := hrLeaveToolNonNegativeInt64(value)
	if err != nil {
		return 0, err
	}
	if parsed <= 0 {
		return 0, errors.New("value is required")
	}
	return parsed, nil
}

func hrLeaveToolNonNegativeInt64(value any) (int64, error) {
	parsed, err := hrLeaveToolInt64(value)
	if err != nil {
		return 0, err
	}
	if parsed < 0 {
		return 0, errors.New("value must be non-negative")
	}
	return parsed, nil
}

func hrLeaveToolInt32(value any) (int32, error) {
	parsed, err := hrLeaveToolNonNegativeInt64(value)
	if err != nil {
		return 0, err
	}
	if parsed > math.MaxInt32 {
		return 0, errors.New("value is too large")
	}
	return int32(parsed), nil // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
}

func hrLeaveToolInt64(value any) (int64, error) {
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

func hrLeaveToolString(value any) string {
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

func hrLeaveTypeToolPayload(leaveType domain.LeaveTypeRecord) map[string]any {
	return map[string]any{
		"id":                                leaveType.ID,
		"name":                              leaveType.Name,
		"is_paid":                           leaveType.IsPaid,
		"requires_approval":                 leaveType.RequiresApproval,
		"accrual_rule":                      leaveType.AccrualRule,
		"accrual_amount":                    leaveType.AccrualAmount,
		"max_carryover_days":                leaveType.MaxCarryoverDays,
		"min_notice_days":                   leaveType.MinNoticeDays,
		"documentation_required_after_days": leaveType.RequiresDocumentationAfterDays,
		"is_active":                         leaveType.IsActive,
	}
}

func hrLeaveRequestToolPayload(request domain.LeaveRequestRecord) map[string]any {
	return map[string]any{
		"id":                     request.ID,
		"employee_id":            request.EmployeeID,
		"leave_type_id":          request.LeaveTypeID,
		"start_date":             request.StartDate,
		"end_date":               request.EndDate,
		"total_days":             request.TotalDays,
		"status":                 request.Status,
		"current_approval_level": request.CurrentApprovalLevel,
		"rejection_reason":       request.RejectionReason,
		"notes":                  request.Notes,
	}
}

func hrLeaveBalanceToolPayload(balance domain.LeaveBalanceRecord) map[string]any {
	return map[string]any{
		"id":                 balance.ID,
		"employee_id":        balance.EmployeeID,
		"leave_type_id":      balance.LeaveTypeID,
		"year":               balance.Year,
		"entitled_days":      balance.EntitledDays,
		"used_days":          balance.UsedDays,
		"pending_days":       balance.PendingDays,
		"remaining_days":     domain.RemainingBalance(balance),
		"last_accrual_month": balance.LastAccrualMonth,
	}
}
