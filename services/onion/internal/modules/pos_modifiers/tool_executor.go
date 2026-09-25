package pos_modifiers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/pos_modifiers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/pagination"
	"github.com/voxire/onion/internal/platform/tenant"
)

const (
	listPOSModifiersToolName         = "pos.list_modifiers"
	defaultListPOSModifiersToolLimit = int64(20)
	maxListPOSModifiersToolLimit     = int64(50)
	posModifierToolTimestampDate     = "2006-01-02T15:04:05Z"
)

type listPOSModifiersToolInput struct {
	Query    string
	Group    string
	Required *bool
	Page     int64
	Limit    int64
}

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	switch name {
	case listPOSModifiersToolName:
		output, err := m.executeListPOSModifiersTool(ctx, input)
		return output, true, err
	default:
		return nil, false, nil
	}
}

func (m *Module) executeListPOSModifiersTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	parsed, err := parseListPOSModifiersToolInput(input)
	if err != nil {
		return nil, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, errors.New("permission denied")
	}
	modifiers, err := m.service.List(ctx, businessID)
	if err != nil {
		return nil, err
	}
	modifiers = filterPOSModifiersForTool(modifiers, parsed)

	start := pagination.SliceStart(parsed.Page, parsed.Limit, len(modifiers))
	if start > len(modifiers) {
		start = len(modifiers)
	}
	end := start + int(parsed.Limit)
	if end > len(modifiers) {
		end = len(modifiers)
	}

	payloads := make([]map[string]any, 0, end-start)
	for _, modifier := range modifiers[start:end] {
		payloads = append(payloads, posModifierToolPayload(modifier))
	}

	return map[string]any{
		"modifiers": payloads,
		"total":     len(modifiers),
		"returned":  len(payloads),
		"page":      parsed.Page,
		"page_size": parsed.Limit,
		"has_more":  end < len(modifiers),
		"filters": map[string]any{
			"query":    parsed.Query,
			"group":    parsed.Group,
			"required": posModifierToolOptionalBool(parsed.Required),
		},
	}, nil
}

func parseListPOSModifiersToolInput(input map[string]any) (listPOSModifiersToolInput, error) {
	query := strings.TrimSpace(posModifierToolString(firstPOSModifierToolValue(input, "query", "q", "search", "name")))
	group := strings.TrimSpace(posModifierToolString(firstPOSModifierToolValue(input, "group", "group_name")))

	required, err := posModifierToolOptionalBoolInput(firstPOSModifierToolValue(input, "required", "is_required"))
	if err != nil {
		return listPOSModifiersToolInput{}, fmt.Errorf("invalid required: %w", err)
	}

	page, err := posModifierToolNonNegativeInt64(firstPOSModifierToolValue(input, "page"))
	if err != nil {
		return listPOSModifiersToolInput{}, fmt.Errorf("invalid page: %w", err)
	}
	if page == 0 {
		page = 1
	}

	limit, err := posModifierToolNonNegativeInt64(firstPOSModifierToolValue(input, "limit", "page_size"))
	if err != nil {
		return listPOSModifiersToolInput{}, fmt.Errorf("invalid limit: %w", err)
	}
	if limit == 0 {
		limit = defaultListPOSModifiersToolLimit
	}
	if limit > maxListPOSModifiersToolLimit {
		limit = maxListPOSModifiersToolLimit
	}

	return listPOSModifiersToolInput{
		Query:    query,
		Group:    group,
		Required: required,
		Page:     page,
		Limit:    limit,
	}, nil
}

func filterPOSModifiersForTool(modifiers []domain.Modifier, input listPOSModifiersToolInput) []domain.Modifier {
	if input.Query == "" && input.Group == "" && input.Required == nil {
		return modifiers
	}

	query := strings.ToLower(input.Query)
	group := strings.ToLower(input.Group)
	out := make([]domain.Modifier, 0, len(modifiers))
	for _, modifier := range modifiers {
		if input.Required != nil && modifier.IsRequired != *input.Required {
			continue
		}
		if group != "" && !strings.Contains(strings.ToLower(modifier.GroupName), group) {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(modifier.Name), query) {
			continue
		}
		out = append(out, modifier)
	}
	return out
}

func posModifierToolPayload(modifier domain.Modifier) map[string]any {
	return map[string]any{
		"id":          modifier.ID,
		"owner_id":    modifier.OwnerID,
		"name":        modifier.Name,
		"price_delta": modifier.PriceDelta,
		"group_name":  modifier.GroupName,
		"is_required": modifier.IsRequired,
		"is_free":     modifier.IsFree(),
		"is_upcharge": modifier.IsUpcharge(),
		"created_at":  formatPOSModifierToolTime(modifier.CreatedAt),
		"updated_at":  formatPOSModifierToolTime(modifier.UpdatedAt),
	}
}

func firstPOSModifierToolValue(input map[string]any, keys ...string) any {
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

func posModifierToolOptionalBoolInput(value any) (*bool, error) {
	switch v := value.(type) {
	case nil:
		return nil, nil
	case bool:
		return &v, nil
	case int:
		return posModifierToolOptionalBoolInt(int64(v))
	case int64:
		return posModifierToolOptionalBoolInt(v)
	case int32:
		return posModifierToolOptionalBoolInt(int64(v))
	case float64:
		if math.Trunc(v) != v {
			return nil, errors.New("must be 0 or 1")
		}
		return posModifierToolOptionalBoolInt(int64(v))
	case string:
		raw := strings.ToLower(strings.TrimSpace(v))
		switch raw {
		case "", "all", "any":
			return nil, nil
		case "true", "yes", "1", "required":
			out := true
			return &out, nil
		case "false", "no", "0", "optional":
			out := false
			return &out, nil
		default:
			return nil, errors.New("must be true, false, required, optional, or all")
		}
	case json.Number:
		i, err := v.Int64()
		if err != nil {
			return nil, err
		}
		return posModifierToolOptionalBoolInt(i)
	default:
		return nil, fmt.Errorf("unsupported value %T", value)
	}
}

func posModifierToolOptionalBoolInt(value int64) (*bool, error) {
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

func posModifierToolOptionalBool(value *bool) any {
	if value == nil {
		return nil
	}
	return *value
}

func posModifierToolNonNegativeInt64(value any) (int64, error) {
	out, err := posModifierToolInt64(value)
	if err != nil {
		return 0, err
	}
	if out < 0 {
		return 0, errors.New("must be non-negative")
	}
	return out, nil
}

func posModifierToolInt64(value any) (int64, error) {
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

func posModifierToolString(value any) string {
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

func formatPOSModifierToolTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value.UTC().Format(posModifierToolTimestampDate)
}
