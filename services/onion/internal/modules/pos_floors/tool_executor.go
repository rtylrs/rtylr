package pos_floors

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/pos_floors/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/pagination"
)

const (
	listPOSFloorsToolName         = "pos.list_floors"
	defaultListPOSFloorsToolLimit = int64(20)
	maxListPOSFloorsToolLimit     = int64(50)
	posFloorToolTimestampDate     = "2006-01-02T15:04:05Z"
)

type listPOSFloorsToolInput struct {
	BranchID int64
	Query    string
	Active   *bool
	Page     int64
	Limit    int64
}

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	switch name {
	case listPOSFloorsToolName:
		output, err := m.executeListPOSFloorsTool(ctx, input)
		return output, true, err
	default:
		return nil, false, nil
	}
}

func (m *Module) executeListPOSFloorsTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}
	if authUser.BusinessID <= 0 {
		return nil, errors.New("business context is required")
	}

	parsed, err := parseListPOSFloorsToolInput(input)
	if err != nil {
		return nil, err
	}

	floors, err := m.service.List(ctx, authUser.BusinessID, parsed.BranchID)
	if err != nil {
		return nil, err
	}
	floors = filterPOSFloorsForTool(floors, parsed)

	start := pagination.SliceStart(parsed.Page, parsed.Limit, len(floors))
	if start > len(floors) {
		start = len(floors)
	}
	end := start + int(parsed.Limit)
	if end > len(floors) {
		end = len(floors)
	}

	payloads := make([]map[string]any, 0, end-start)
	for _, floor := range floors[start:end] {
		payloads = append(payloads, posFloorToolPayload(floor))
	}

	return map[string]any{
		"floors":    payloads,
		"total":     len(floors),
		"returned":  len(payloads),
		"page":      parsed.Page,
		"page_size": parsed.Limit,
		"has_more":  end < len(floors),
		"filters": map[string]any{
			"branch_id": parsed.BranchID,
			"query":     parsed.Query,
			"active":    posFloorToolOptionalBool(parsed.Active),
		},
	}, nil
}

func parseListPOSFloorsToolInput(input map[string]any) (listPOSFloorsToolInput, error) {
	branchID, err := posFloorToolInt64(firstPOSFloorToolValue(input, "branch_id", "location_id", "branch", "location"))
	if err != nil {
		return listPOSFloorsToolInput{}, fmt.Errorf("invalid branch_id: %w", err)
	}
	if branchID <= 0 {
		return listPOSFloorsToolInput{}, errors.New("branch_id is required")
	}

	query := strings.TrimSpace(posFloorToolString(firstPOSFloorToolValue(input, "query", "q", "search", "name")))
	active, err := posFloorToolOptionalBoolInput(firstPOSFloorToolValue(input, "active", "is_active", "enabled", "status"))
	if err != nil {
		return listPOSFloorsToolInput{}, fmt.Errorf("invalid active: %w", err)
	}

	page, err := posFloorToolNonNegativeInt64(firstPOSFloorToolValue(input, "page"))
	if err != nil {
		return listPOSFloorsToolInput{}, fmt.Errorf("invalid page: %w", err)
	}
	if page == 0 {
		page = 1
	}

	limit, err := posFloorToolNonNegativeInt64(firstPOSFloorToolValue(input, "limit", "page_size"))
	if err != nil {
		return listPOSFloorsToolInput{}, fmt.Errorf("invalid limit: %w", err)
	}
	if limit == 0 {
		limit = defaultListPOSFloorsToolLimit
	}
	if limit > maxListPOSFloorsToolLimit {
		limit = maxListPOSFloorsToolLimit
	}

	return listPOSFloorsToolInput{
		BranchID: branchID,
		Query:    query,
		Active:   active,
		Page:     page,
		Limit:    limit,
	}, nil
}

func filterPOSFloorsForTool(floors []domain.Floor, input listPOSFloorsToolInput) []domain.Floor {
	if input.Query == "" && input.Active == nil {
		return floors
	}

	query := strings.ToLower(input.Query)
	out := make([]domain.Floor, 0, len(floors))
	for _, floor := range floors {
		if input.Active != nil && floor.IsActive != *input.Active {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(floor.Name), query) {
			continue
		}
		out = append(out, floor)
	}
	return out
}

func posFloorToolPayload(floor domain.Floor) map[string]any {
	return map[string]any{
		"id":          floor.ID,
		"business_id": floor.BusinessID,
		"branch_id":   floor.BranchID,
		"name":        floor.Name,
		"sort_order":  floor.SortOrder,
		"is_active":   floor.IsActive,
		"created_at":  formatPOSFloorToolTime(floor.CreatedAt),
	}
}

func firstPOSFloorToolValue(input map[string]any, keys ...string) any {
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

func posFloorToolOptionalBoolInput(value any) (*bool, error) {
	switch v := value.(type) {
	case nil:
		return nil, nil
	case bool:
		return &v, nil
	case int:
		return posFloorToolOptionalBoolInt(int64(v))
	case int64:
		return posFloorToolOptionalBoolInt(v)
	case int32:
		return posFloorToolOptionalBoolInt(int64(v))
	case float64:
		if math.Trunc(v) != v {
			return nil, errors.New("must be 0 or 1")
		}
		return posFloorToolOptionalBoolInt(int64(v))
	case string:
		raw := strings.ToLower(strings.TrimSpace(v))
		switch raw {
		case "", "all", "any":
			return nil, nil
		case "true", "yes", "1", "active", "enabled":
			out := true
			return &out, nil
		case "false", "no", "0", "inactive", "disabled":
			out := false
			return &out, nil
		default:
			return nil, errors.New("must be true, false, active, inactive, or all")
		}
	case json.Number:
		i, err := v.Int64()
		if err != nil {
			return nil, err
		}
		return posFloorToolOptionalBoolInt(i)
	default:
		return nil, fmt.Errorf("unsupported value %T", value)
	}
}

func posFloorToolOptionalBoolInt(value int64) (*bool, error) {
	if value == 0 {
		out := false
		return &out, nil
	}
	if value == 1 {
		out := true
		return &out, nil
	}
	return nil, errors.New("must be 0 or 1")
}

func posFloorToolOptionalBool(value *bool) any {
	if value == nil {
		return nil
	}
	return *value
}

func posFloorToolNonNegativeInt64(value any) (int64, error) {
	out, err := posFloorToolInt64(value)
	if err != nil {
		return 0, err
	}
	if out < 0 {
		return 0, errors.New("must be non-negative")
	}
	return out, nil
}

func posFloorToolInt64(value any) (int64, error) {
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

func posFloorToolString(value any) string {
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

func formatPOSFloorToolTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value.UTC().Format(posFloorToolTimestampDate)
}
