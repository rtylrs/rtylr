package hr_overtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/voxire/onion/internal/modules/hr_overtime/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
)

const (
	listHROvertimeRequestsToolName = "hr.list_overtime_requests"
	getHROvertimeConfigToolName    = "hr.get_overtime_config"
	defaultHROvertimeToolLimit     = int64(20)
	maxHROvertimeToolLimit         = int64(50)
)

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	switch name {
	case listHROvertimeRequestsToolName:
		output, err := m.executeListOvertimeRequestsTool(ctx, input)
		return output, true, err
	case getHROvertimeConfigToolName:
		output, err := m.executeGetOvertimeConfigTool(ctx)
		return output, true, err
	default:
		return nil, false, nil
	}
}

func (m *Module) executeListOvertimeRequestsTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	ownerID, err := hrOvertimeToolOwnerID(ctx)
	if err != nil {
		return nil, err
	}
	limit, offset, err := hrOvertimeToolPagination(input)
	if err != nil {
		return nil, err
	}

	requests, total, err := m.service.ListOvertimeRequests(ctx, ownerID, domain.ListOvertimeRequestsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	payloads := make([]map[string]any, 0, len(requests))
	for _, request := range requests {
		payloads = append(payloads, hrOvertimeRequestToolPayload(request))
	}

	return map[string]any{
		"requests": payloads,
		"total":    total,
		"returned": len(payloads),
		"limit":    limit,
		"offset":   offset,
	}, nil
}

func (m *Module) executeGetOvertimeConfigTool(ctx context.Context) (map[string]any, error) {
	ownerID, err := hrOvertimeToolOwnerID(ctx)
	if err != nil {
		return nil, err
	}

	config, err := m.service.GetOvertimeConfig(ctx, ownerID)
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
		"config":     hrOvertimeConfigToolPayload(*config),
	}, nil
}

func hrOvertimeToolOwnerID(ctx context.Context) (int64, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return 0, errors.New("authentication is required")
	}
	return authUser.UserID, nil
}

func hrOvertimeToolPagination(input map[string]any) (int32, int32, error) {
	limit, err := hrOvertimeToolNonNegativeInt64(firstHROvertimeToolValue(input, "limit", "page_size"))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid limit: %w", err)
	}
	if limit == 0 {
		limit = defaultHROvertimeToolLimit
	}
	if limit > maxHROvertimeToolLimit {
		limit = maxHROvertimeToolLimit
	}

	offset, err := hrOvertimeToolNonNegativeInt64(firstHROvertimeToolValue(input, "offset"))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid offset: %w", err)
	}
	if offset > math.MaxInt32 {
		return 0, 0, errors.New("offset is too large")
	}
	return int32(limit), int32(offset), nil // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
}

func firstHROvertimeToolValue(input map[string]any, keys ...string) any {
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

func hrOvertimeToolNonNegativeInt64(value any) (int64, error) {
	parsed, err := hrOvertimeToolInt64(value)
	if err != nil {
		return 0, err
	}
	if parsed < 0 {
		return 0, errors.New("value must be non-negative")
	}
	return parsed, nil
}

func hrOvertimeToolInt64(value any) (int64, error) {
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

func hrOvertimeRequestToolPayload(request domain.OvertimeRequestRecord) map[string]any {
	return map[string]any{
		"id":                     request.ID,
		"employee_id":            request.EmployeeID,
		"date":                   request.Date,
		"planned_extra_minutes":  request.PlannedExtraMinutes,
		"reason":                 request.Reason,
		"status":                 request.Status,
		"current_approval_level": request.CurrentApprovalLevel,
		"approved_by":            request.ApprovedBy,
		"multiplier_applied":     request.MultiplierApplied,
	}
}

func hrOvertimeConfigToolPayload(config domain.OvertimeConfigRecord) map[string]any {
	return map[string]any{
		"id":                      config.ID,
		"daily_threshold_minutes": config.DailyThresholdMinutes,
		"default_multiplier":      config.DefaultMultiplier,
		"auto_detect_enabled":     config.AutoDetectEnabled,
	}
}
