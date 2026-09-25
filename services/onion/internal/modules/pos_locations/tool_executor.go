package pos_locations

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/pos_locations/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/pagination"
	"github.com/voxire/onion/internal/platform/tenant"
)

const (
	listPOSLocationsToolName         = "pos.list_locations"
	defaultListPOSLocationsToolLimit = int64(20)
	maxListPOSLocationsToolLimit     = int64(50)
	posLocationToolTimestampDate     = "2006-01-02T15:04:05Z"
)

type listPOSLocationsToolInput struct {
	Query  string
	Active *bool
	Page   int64
	Limit  int64
}

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	switch name {
	case listPOSLocationsToolName:
		output, err := m.executeListPOSLocationsTool(ctx, input)
		return output, true, err
	default:
		return nil, false, nil
	}
}

func (m *Module) executeListPOSLocationsTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	parsed, err := parseListPOSLocationsToolInput(input)
	if err != nil {
		return nil, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	locations, err := m.service.List(ctx, businessID)
	if err != nil {
		return nil, err
	}
	locations = filterPOSLocationsForTool(locations, parsed)

	start := pagination.SliceStart(parsed.Page, parsed.Limit, len(locations))
	if start > len(locations) {
		start = len(locations)
	}
	end := start + int(parsed.Limit)
	if end > len(locations) {
		end = len(locations)
	}

	payloads := make([]map[string]any, 0, end-start)
	for _, location := range locations[start:end] {
		payloads = append(payloads, posLocationToolPayload(location))
	}

	return map[string]any{
		"locations": payloads,
		"total":     len(locations),
		"returned":  len(payloads),
		"page":      parsed.Page,
		"page_size": parsed.Limit,
		"has_more":  end < len(locations),
		"filters": map[string]any{
			"query":  parsed.Query,
			"active": posLocationToolOptionalBool(parsed.Active),
		},
	}, nil
}

func parseListPOSLocationsToolInput(input map[string]any) (listPOSLocationsToolInput, error) {
	query := strings.TrimSpace(posLocationToolString(firstPOSLocationToolValue(input, "query", "q", "search", "name")))

	active, err := posLocationToolOptionalBoolInput(firstPOSLocationToolValue(input, "active", "is_active", "enabled", "status"))
	if err != nil {
		return listPOSLocationsToolInput{}, fmt.Errorf("invalid active: %w", err)
	}

	page, err := posLocationToolNonNegativeInt64(firstPOSLocationToolValue(input, "page"))
	if err != nil {
		return listPOSLocationsToolInput{}, fmt.Errorf("invalid page: %w", err)
	}
	if page == 0 {
		page = 1
	}

	limit, err := posLocationToolNonNegativeInt64(firstPOSLocationToolValue(input, "limit", "page_size"))
	if err != nil {
		return listPOSLocationsToolInput{}, fmt.Errorf("invalid limit: %w", err)
	}
	if limit == 0 {
		limit = defaultListPOSLocationsToolLimit
	}
	if limit > maxListPOSLocationsToolLimit {
		limit = maxListPOSLocationsToolLimit
	}

	return listPOSLocationsToolInput{
		Query:  query,
		Active: active,
		Page:   page,
		Limit:  limit,
	}, nil
}

func filterPOSLocationsForTool(locations []domain.Location, input listPOSLocationsToolInput) []domain.Location {
	if input.Query == "" && input.Active == nil {
		return locations
	}

	query := strings.ToLower(input.Query)
	out := make([]domain.Location, 0, len(locations))
	for _, location := range locations {
		if input.Active != nil && location.IsActive != *input.Active {
			continue
		}
		if query != "" && !posLocationToolMatchesQuery(location, query) {
			continue
		}
		out = append(out, location)
	}
	return out
}

func posLocationToolMatchesQuery(location domain.Location, query string) bool {
	return strings.Contains(strings.ToLower(location.Name), query) ||
		strings.Contains(strings.ToLower(location.Address), query) ||
		strings.Contains(strings.ToLower(location.Phone), query)
}

func posLocationToolPayload(location domain.Location) map[string]any {
	return map[string]any{
		"id":         location.ID,
		"owner_id":   location.OwnerID,
		"name":       location.Name,
		"address":    location.Address,
		"phone":      location.Phone,
		"is_active":  location.IsActive,
		"created_at": formatPOSLocationToolTime(location.CreatedAt),
		"updated_at": formatPOSLocationToolTime(location.UpdatedAt),
	}
}

func firstPOSLocationToolValue(input map[string]any, keys ...string) any {
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

func posLocationToolOptionalBoolInput(value any) (*bool, error) {
	switch v := value.(type) {
	case nil:
		return nil, nil
	case bool:
		return &v, nil
	case int:
		return posLocationToolOptionalBoolInt(int64(v))
	case int64:
		return posLocationToolOptionalBoolInt(v)
	case int32:
		return posLocationToolOptionalBoolInt(int64(v))
	case float64:
		if math.Trunc(v) != v {
			return nil, errors.New("must be 0 or 1")
		}
		return posLocationToolOptionalBoolInt(int64(v))
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

func posLocationToolOptionalBoolInt(value int64) (*bool, error) {
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

func posLocationToolOptionalBool(value *bool) any {
	if value == nil {
		return nil
	}
	return *value
}

func posLocationToolNonNegativeInt64(value any) (int64, error) {
	out, err := posLocationToolInt64(value)
	if err != nil {
		return 0, err
	}
	if out < 0 {
		return 0, errors.New("must be non-negative")
	}
	return out, nil
}

func posLocationToolInt64(value any) (int64, error) {
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

func posLocationToolString(value any) string {
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

func formatPOSLocationToolTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value.UTC().Format(posLocationToolTimestampDate)
}
