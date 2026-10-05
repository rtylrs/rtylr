package terminals

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/pos/terminals/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/pagination"
)

const (
	listPOSTerminalsToolName         = "pos.list_terminals"
	defaultListPOSTerminalsToolLimit = int64(20)
	maxListPOSTerminalsToolLimit     = int64(50)
	posTerminalToolTimestampDate     = "2006-01-02T15:04:05Z"
)

type listPOSTerminalsToolInput struct {
	Query    string
	BranchID int64
	Active   *bool
	Online   *bool
	Page     int64
	Limit    int64
}

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	switch name {
	case listPOSTerminalsToolName:
		output, err := m.executeListPOSTerminalsTool(ctx, input)
		return output, true, err
	default:
		return nil, false, nil
	}
}

func (m *Module) executeListPOSTerminalsTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}
	if authUser.BusinessID <= 0 {
		return nil, errors.New("business context is required")
	}

	parsed, err := parseListPOSTerminalsToolInput(input)
	if err != nil {
		return nil, err
	}

	terminals, err := m.service.List(ctx, authUser.BusinessID)
	if err != nil {
		return nil, err
	}
	terminals = filterPOSTerminalsForTool(terminals, parsed)

	start := pagination.SliceStart(parsed.Page, parsed.Limit, len(terminals))
	if start > len(terminals) {
		start = len(terminals)
	}
	end := start + int(parsed.Limit)
	if end > len(terminals) {
		end = len(terminals)
	}

	payloads := make([]map[string]any, 0, end-start)
	for _, terminal := range terminals[start:end] {
		payloads = append(payloads, posTerminalToolPayload(terminal))
	}

	return map[string]any{
		"terminals": payloads,
		"total":     len(terminals),
		"returned":  len(payloads),
		"page":      parsed.Page,
		"page_size": parsed.Limit,
		"has_more":  end < len(terminals),
		"filters": map[string]any{
			"query":     parsed.Query,
			"branch_id": parsed.BranchID,
			"active":    posTerminalToolOptionalBool(parsed.Active),
			"online":    posTerminalToolOptionalBool(parsed.Online),
		},
	}, nil
}

func parseListPOSTerminalsToolInput(input map[string]any) (listPOSTerminalsToolInput, error) {
	query := strings.TrimSpace(posTerminalToolString(firstPOSTerminalToolValue(input, "query", "q", "search", "name")))

	branchID, err := posTerminalToolNonNegativeInt64(firstPOSTerminalToolValue(input, "branch_id", "location_id", "branch", "location"))
	if err != nil {
		return listPOSTerminalsToolInput{}, fmt.Errorf("invalid branch_id: %w", err)
	}

	active, err := posTerminalToolOptionalBoolInput(firstPOSTerminalToolValue(input, "active", "is_active", "enabled", "status"))
	if err != nil {
		return listPOSTerminalsToolInput{}, fmt.Errorf("invalid active: %w", err)
	}
	online, err := posTerminalToolOptionalBoolInput(firstPOSTerminalToolValue(input, "online", "is_online"))
	if err != nil {
		return listPOSTerminalsToolInput{}, fmt.Errorf("invalid online: %w", err)
	}

	page, err := posTerminalToolNonNegativeInt64(firstPOSTerminalToolValue(input, "page"))
	if err != nil {
		return listPOSTerminalsToolInput{}, fmt.Errorf("invalid page: %w", err)
	}
	if page == 0 {
		page = 1
	}

	limit, err := posTerminalToolNonNegativeInt64(firstPOSTerminalToolValue(input, "limit", "page_size"))
	if err != nil {
		return listPOSTerminalsToolInput{}, fmt.Errorf("invalid limit: %w", err)
	}
	if limit == 0 {
		limit = defaultListPOSTerminalsToolLimit
	}
	if limit > maxListPOSTerminalsToolLimit {
		limit = maxListPOSTerminalsToolLimit
	}

	return listPOSTerminalsToolInput{
		Query:    query,
		BranchID: branchID,
		Active:   active,
		Online:   online,
		Page:     page,
		Limit:    limit,
	}, nil
}

func filterPOSTerminalsForTool(terminals []domain.Terminal, input listPOSTerminalsToolInput) []domain.Terminal {
	if input.Query == "" && input.BranchID == 0 && input.Active == nil && input.Online == nil {
		return terminals
	}

	query := strings.ToLower(input.Query)
	out := make([]domain.Terminal, 0, len(terminals))
	for _, terminal := range terminals {
		if input.BranchID > 0 && terminal.BranchID != input.BranchID {
			continue
		}
		if input.Active != nil && terminal.IsActive != *input.Active {
			continue
		}
		if input.Online != nil && terminal.IsOnline() != *input.Online {
			continue
		}
		if query != "" && !posTerminalToolMatchesQuery(terminal, query) {
			continue
		}
		out = append(out, terminal)
	}
	return out
}

func posTerminalToolMatchesQuery(terminal domain.Terminal, query string) bool {
	if strings.Contains(strings.ToLower(terminal.Name), query) {
		return true
	}
	if terminal.DeviceID != nil && strings.Contains(strings.ToLower(*terminal.DeviceID), query) {
		return true
	}
	return false
}

func posTerminalToolPayload(terminal domain.Terminal) map[string]any {
	return map[string]any{
		"id":           terminal.ID,
		"business_id":  terminal.BusinessID,
		"branch_id":    terminal.BranchID,
		"name":         terminal.Name,
		"device_id":    posTerminalToolOptionalString(terminal.DeviceID),
		"is_active":    terminal.IsActive,
		"is_online":    terminal.IsOnline(),
		"last_seen_at": formatPOSTerminalToolOptionalTime(terminal.LastSeenAt),
		"created_at":   formatPOSTerminalToolTime(terminal.CreatedAt),
	}
}

func firstPOSTerminalToolValue(input map[string]any, keys ...string) any {
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

func posTerminalToolOptionalBoolInput(value any) (*bool, error) {
	switch v := value.(type) {
	case nil:
		return nil, nil
	case bool:
		return &v, nil
	case int:
		return posTerminalToolOptionalBoolInt(int64(v))
	case int64:
		return posTerminalToolOptionalBoolInt(v)
	case int32:
		return posTerminalToolOptionalBoolInt(int64(v))
	case float64:
		if math.Trunc(v) != v {
			return nil, errors.New("must be 0 or 1")
		}
		return posTerminalToolOptionalBoolInt(int64(v))
	case string:
		raw := strings.ToLower(strings.TrimSpace(v))
		switch raw {
		case "", "all", "any":
			return nil, nil
		case "true", "yes", "1", "active", "enabled", "online":
			out := true
			return &out, nil
		case "false", "no", "0", "inactive", "disabled", "offline":
			out := false
			return &out, nil
		default:
			return nil, errors.New("must be true, false, active, inactive, online, offline, or all")
		}
	case json.Number:
		i, err := v.Int64()
		if err != nil {
			return nil, err
		}
		if i == 0 {
			out := false
			return &out, nil
		}
		if i == 1 {
			out := true
			return &out, nil
		}
		return nil, errors.New("must be 0 or 1")
	default:
		return nil, fmt.Errorf("unsupported value %T", value)
	}
}

func posTerminalToolOptionalBoolInt(value int64) (*bool, error) {
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

func posTerminalToolOptionalBool(value *bool) any {
	if value == nil {
		return nil
	}
	return *value
}

func posTerminalToolNonNegativeInt64(value any) (int64, error) {
	out, err := posTerminalToolInt64(value)
	if err != nil {
		return 0, err
	}
	if out < 0 {
		return 0, errors.New("must be non-negative")
	}
	return out, nil
}

func posTerminalToolInt64(value any) (int64, error) {
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

func posTerminalToolString(value any) string {
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

func posTerminalToolOptionalString(value *string) any {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil
	}
	return *value
}

func formatPOSTerminalToolOptionalTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return formatPOSTerminalToolTime(*value)
}

func formatPOSTerminalToolTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value.UTC().Format(posTerminalToolTimestampDate)
}
