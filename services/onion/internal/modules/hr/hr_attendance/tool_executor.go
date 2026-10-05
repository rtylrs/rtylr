package hr_attendance

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/hr/hr_attendance/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/tenant"
)

const (
	listHRAttendanceToolName        = "hr.list_attendance"
	getHRLocationAttendanceToolName = "hr.get_location_attendance"
	defaultHRAttendanceToolLimit    = int64(20)
	maxHRAttendanceToolLimit        = int64(50)
)

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	switch name {
	case listHRAttendanceToolName:
		output, err := m.executeListAttendanceTool(ctx, input)
		return output, true, err
	case getHRLocationAttendanceToolName:
		output, err := m.executeGetLocationAttendanceTool(ctx, input)
		return output, true, err
	default:
		return nil, false, nil
	}
}

func (m *Module) executeListAttendanceTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	ownerID, employeeID, err := m.hrAttendanceToolScope(ctx)
	// employee_id from the input is intentionally ignored: reads are scoped
	// to the caller's own employee record (#817).
	if err != nil {
		return nil, err
	}
	limit, offset, err := hrAttendanceToolPagination(input)
	if err != nil {
		return nil, err
	}
	dateFrom, err := hrAttendanceToolDate(firstHRAttendanceToolValue(input, "date_from", "from"), "date_from")
	if err != nil {
		return nil, err
	}
	dateTo, err := hrAttendanceToolDate(firstHRAttendanceToolValue(input, "date_to", "to"), "date_to")
	if err != nil {
		return nil, err
	}

	records, total, err := m.service.GetMyAttendance(ctx, ownerID, employeeID, domain.ListAttendanceParams{
		DateFrom: dateFrom,
		DateTo:   dateTo,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		return nil, err
	}

	payloads := make([]map[string]any, 0, len(records))
	for _, record := range records {
		payloads = append(payloads, hrAttendanceRecordToolPayload(record))
	}

	return map[string]any{
		"records":  payloads,
		"total":    total,
		"returned": len(payloads),
		"limit":    limit,
		"offset":   offset,
		"filters": map[string]any{
			"employee_id": employeeID,
			"date_from":   dateFrom,
			"date_to":     dateTo,
		},
	}, nil
}

func (m *Module) executeGetLocationAttendanceTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	ownerID, _, err := m.hrAttendanceToolScope(ctx)
	if err != nil {
		return nil, err
	}
	locationID, err := hrAttendanceToolPositiveInt64(firstHRAttendanceToolValue(input, "location_id", "location"))
	if err != nil {
		return nil, fmt.Errorf("invalid location_id: %w", err)
	}
	date, err := hrAttendanceToolDate(firstHRAttendanceToolValue(input, "date", "day"), "date")
	if err != nil {
		return nil, err
	}
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	records, err := m.service.GetLocationAttendance(ctx, ownerID, locationID, date)
	if err != nil {
		return nil, err
	}

	payloads := make([]map[string]any, 0, len(records))
	for _, record := range records {
		payloads = append(payloads, hrAttendanceRecordToolPayload(record))
	}

	return map[string]any{
		"records":  payloads,
		"returned": len(payloads),
		"filters": map[string]any{
			"location_id": locationID,
			"date":        date,
		},
	}, nil
}

// hrAttendanceToolScope derives the caller's tenant scope and employee record
// from the JWT. Client-supplied owner/employee ids are ignored: trusting them
// let any authenticated user read any tenant's attendance via a Caret tool
// (#817).
func (m *Module) hrAttendanceToolScope(ctx context.Context) (businessID, employeeID int64, err error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return 0, 0, errors.New("authentication is required")
	}
	businessID, err = tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if err != nil {
		return 0, 0, err
	}
	employeeID, err = m.employeeResolver.GetEmployeeIDByUserID(ctx, businessID, authUser.UserID)
	if err != nil || employeeID <= 0 {
		return 0, 0, errors.New("no employee record for caller")
	}
	return businessID, employeeID, nil
}

func hrAttendanceToolPagination(input map[string]any) (int32, int32, error) {
	limit, err := hrAttendanceToolNonNegativeInt64(firstHRAttendanceToolValue(input, "limit", "page_size"))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid limit: %w", err)
	}
	if limit == 0 {
		limit = defaultHRAttendanceToolLimit
	}
	if limit > maxHRAttendanceToolLimit {
		limit = maxHRAttendanceToolLimit
	}

	offset, err := hrAttendanceToolNonNegativeInt64(firstHRAttendanceToolValue(input, "offset"))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid offset: %w", err)
	}
	if offset > math.MaxInt32 {
		return 0, 0, errors.New("offset is too large")
	}
	return int32(limit), int32(offset), nil // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
}

func firstHRAttendanceToolValue(input map[string]any, keys ...string) any {
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

func hrAttendanceToolDate(value any, field string) (string, error) {
	date := strings.TrimSpace(hrAttendanceToolString(value))
	if date == "" {
		return "", nil
	}
	if strings.EqualFold(date, "today") {
		return time.Now().Format("2006-01-02"), nil
	}
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return "", fmt.Errorf("%s must be in YYYY-MM-DD format", field)
	}
	return date, nil
}

func hrAttendanceToolPositiveInt64(value any) (int64, error) {
	parsed, err := hrAttendanceToolNonNegativeInt64(value)
	if err != nil {
		return 0, err
	}
	if parsed <= 0 {
		return 0, errors.New("value is required")
	}
	return parsed, nil
}

func hrAttendanceToolNonNegativeInt64(value any) (int64, error) {
	parsed, err := hrAttendanceToolInt64(value)
	if err != nil {
		return 0, err
	}
	if parsed < 0 {
		return 0, errors.New("value must be non-negative")
	}
	return parsed, nil
}

func hrAttendanceToolInt64(value any) (int64, error) {
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

func hrAttendanceToolString(value any) string {
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

func hrAttendanceRecordToolPayload(record domain.AttendanceRecordRecord) map[string]any {
	return map[string]any{
		"id":               record.ID,
		"employee_id":      record.EmployeeID,
		"location_id":      record.LocationID,
		"date":             record.Date,
		"clock_in_at":      record.ClockInAt,
		"clock_out_at":     record.ClockOutAt,
		"duration_minutes": record.DurationMinutes,
		"clock_in_method":  record.ClockInMethod,
		"is_late":          record.IsLate,
		"late_minutes":     record.LateMinutes,
		"status":           record.Status,
	}
}
