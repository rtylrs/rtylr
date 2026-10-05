package users

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/pos/users/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/pagination"
)

const (
	listPOSUsersToolName         = "pos.list_users"
	defaultListPOSUsersToolLimit = int64(20)
	maxListPOSUsersToolLimit     = int64(50)
	posUserToolTimestampDate     = "2006-01-02T15:04:05Z"
)

type listPOSUsersToolInput struct {
	Query      string
	LocationID int64
	Role       string
	Active     *bool
	Page       int64
	Limit      int64
}

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	switch name {
	case listPOSUsersToolName:
		output, err := m.executeListPOSUsersTool(ctx, input)
		return output, true, err
	default:
		return nil, false, nil
	}
}

func (m *Module) executeListPOSUsersTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}
	if authUser.BusinessID <= 0 {
		return nil, errors.New("business context is required")
	}

	parsed, err := parseListPOSUsersToolInput(input)
	if err != nil {
		return nil, err
	}

	var users []domain.POSUser
	if parsed.LocationID > 0 {
		users, err = m.service.ListByLocation(ctx, authUser.BusinessID, parsed.LocationID)
	} else {
		users, err = m.service.ListByOwner(ctx, authUser.BusinessID)
	}
	if err != nil {
		return nil, err
	}
	users = filterPOSUsersForTool(users, parsed)

	start := pagination.SliceStart(parsed.Page, parsed.Limit, len(users))
	if start > len(users) {
		start = len(users)
	}
	end := start + int(parsed.Limit)
	if end > len(users) {
		end = len(users)
	}

	payloads := make([]map[string]any, 0, end-start)
	for _, user := range users[start:end] {
		payloads = append(payloads, posUserToolPayload(user))
	}

	return map[string]any{
		"users":     payloads,
		"total":     len(users),
		"returned":  len(payloads),
		"page":      parsed.Page,
		"page_size": parsed.Limit,
		"has_more":  end < len(users),
		"filters": map[string]any{
			"query":       parsed.Query,
			"location_id": parsed.LocationID,
			"role":        parsed.Role,
			"active":      posUserToolOptionalBool(parsed.Active),
		},
	}, nil
}

func parseListPOSUsersToolInput(input map[string]any) (listPOSUsersToolInput, error) {
	query := strings.TrimSpace(posUserToolString(firstPOSUserToolValue(input, "query", "q", "search", "name")))

	locationID, err := posUserToolNonNegativeInt64(firstPOSUserToolValue(input, "location_id", "branch_id", "location", "branch"))
	if err != nil {
		return listPOSUsersToolInput{}, fmt.Errorf("invalid location_id: %w", err)
	}

	role := normalizePOSUserToolRole(firstPOSUserToolValue(input, "role", "pos_role"))
	if strings.TrimSpace(posUserToolString(firstPOSUserToolValue(input, "role", "pos_role"))) != "" && role == "" {
		return listPOSUsersToolInput{}, errors.New("role must be manager, cashier, server, bartender, kitchen, or all")
	}

	active, err := posUserToolOptionalBoolInput(firstPOSUserToolValue(input, "active", "is_active", "enabled", "status"))
	if err != nil {
		return listPOSUsersToolInput{}, fmt.Errorf("invalid active: %w", err)
	}

	page, err := posUserToolNonNegativeInt64(firstPOSUserToolValue(input, "page"))
	if err != nil {
		return listPOSUsersToolInput{}, fmt.Errorf("invalid page: %w", err)
	}
	if page == 0 {
		page = 1
	}

	limit, err := posUserToolNonNegativeInt64(firstPOSUserToolValue(input, "limit", "page_size"))
	if err != nil {
		return listPOSUsersToolInput{}, fmt.Errorf("invalid limit: %w", err)
	}
	if limit == 0 {
		limit = defaultListPOSUsersToolLimit
	}
	if limit > maxListPOSUsersToolLimit {
		limit = maxListPOSUsersToolLimit
	}

	return listPOSUsersToolInput{
		Query:      query,
		LocationID: locationID,
		Role:       role,
		Active:     active,
		Page:       page,
		Limit:      limit,
	}, nil
}

func filterPOSUsersForTool(users []domain.POSUser, input listPOSUsersToolInput) []domain.POSUser {
	if input.Query == "" && input.Role == "" && input.Active == nil {
		return users
	}

	query := strings.ToLower(input.Query)
	out := make([]domain.POSUser, 0, len(users))
	for _, user := range users {
		if input.Role != "" && !strings.EqualFold(string(user.Role), input.Role) {
			continue
		}
		if input.Active != nil && user.IsActive != *input.Active {
			continue
		}
		if query != "" && !posUserToolMatchesQuery(user, query) {
			continue
		}
		out = append(out, user)
	}
	return out
}

func posUserToolMatchesQuery(user domain.POSUser, query string) bool {
	return strings.Contains(strings.ToLower(user.Name), query) ||
		strings.Contains(strconv.FormatInt(user.UserID, 10), query)
}

func posUserToolPayload(user domain.POSUser) map[string]any {
	return map[string]any{
		"id":               user.ID,
		"owner_id":         user.OwnerID,
		"user_id":          user.UserID,
		"location_id":      user.LocationID,
		"name":             user.Name,
		"role":             string(user.Role),
		"is_active":        user.IsActive,
		"max_discount_pct": user.MaxDiscountPct,
		"capabilities": map[string]any{
			"can_void":           user.CanVoid,
			"can_refund":         user.CanRefund,
			"can_open_drawer":    user.CanOpenDrawer,
			"can_apply_discount": user.CanApplyDiscount,
		},
		"created_by": user.CreatedBy,
		"created_at": formatPOSUserToolTime(user.CreatedAt),
		"updated_at": formatPOSUserToolTime(user.UpdatedAt),
	}
}

func normalizePOSUserToolRole(value any) string {
	role := strings.ToLower(strings.TrimSpace(posUserToolString(value)))
	role = strings.ReplaceAll(role, "-", "_")
	role = strings.ReplaceAll(role, " ", "_")
	switch role {
	case "", "all", "any":
		return ""
	case "manager", "cashier", "server", "bartender", "kitchen":
		return role
	default:
		return ""
	}
}

func firstPOSUserToolValue(input map[string]any, keys ...string) any {
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

func posUserToolOptionalBoolInput(value any) (*bool, error) {
	switch v := value.(type) {
	case nil:
		return nil, nil
	case bool:
		return &v, nil
	case int:
		return posUserToolOptionalBoolInt(int64(v))
	case int64:
		return posUserToolOptionalBoolInt(v)
	case int32:
		return posUserToolOptionalBoolInt(int64(v))
	case float64:
		if math.Trunc(v) != v {
			return nil, errors.New("must be 0 or 1")
		}
		return posUserToolOptionalBoolInt(int64(v))
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

func posUserToolOptionalBoolInt(value int64) (*bool, error) {
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

func posUserToolOptionalBool(value *bool) any {
	if value == nil {
		return nil
	}
	return *value
}

func posUserToolNonNegativeInt64(value any) (int64, error) {
	out, err := posUserToolInt64(value)
	if err != nil {
		return 0, err
	}
	if out < 0 {
		return 0, errors.New("must be non-negative")
	}
	return out, nil
}

func posUserToolInt64(value any) (int64, error) {
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

func posUserToolString(value any) string {
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

func formatPOSUserToolTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value.UTC().Format(posUserToolTimestampDate)
}
