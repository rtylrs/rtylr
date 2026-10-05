package tables

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/pos/tables/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/pagination"
	"github.com/voxire/onion/internal/platform/tenant"
)

const (
	listPOSTablesToolName         = "pos.list_tables"
	defaultListPOSTablesToolLimit = int64(20)
	maxListPOSTablesToolLimit     = int64(50)
	posTableToolTimestampDate     = "2006-01-02T15:04:05Z"
)

type listPOSTablesToolInput struct {
	Query      string
	LocationID int64
	Status     string
	MinSeats   int64
	Page       int64
	Limit      int64
}

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	switch name {
	case listPOSTablesToolName:
		output, err := m.executeListPOSTablesTool(ctx, input)
		return output, true, err
	default:
		return nil, false, nil
	}
}

func (m *Module) executeListPOSTablesTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	parsed, err := parseListPOSTablesToolInput(input)
	if err != nil {
		return nil, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, errors.New("failed to resolve business ID")
	}
	tables, err := m.service.List(ctx, businessID)
	if err != nil {
		return nil, err
	}
	tables = filterPOSTablesForTool(tables, parsed)

	start := pagination.SliceStart(parsed.Page, parsed.Limit, len(tables))
	if start > len(tables) {
		start = len(tables)
	}
	end := start + int(parsed.Limit)
	if end > len(tables) {
		end = len(tables)
	}

	payloads := make([]map[string]any, 0, end-start)
	for _, table := range tables[start:end] {
		payloads = append(payloads, posTableToolPayload(table))
	}

	return map[string]any{
		"tables":    payloads,
		"total":     len(tables),
		"returned":  len(payloads),
		"page":      parsed.Page,
		"page_size": parsed.Limit,
		"has_more":  end < len(tables),
		"filters": map[string]any{
			"query":       parsed.Query,
			"location_id": parsed.LocationID,
			"status":      parsed.Status,
			"min_seats":   parsed.MinSeats,
		},
	}, nil
}

func parseListPOSTablesToolInput(input map[string]any) (listPOSTablesToolInput, error) {
	query := strings.TrimSpace(posTableToolString(firstPOSTableToolValue(input, "query", "q", "search", "label")))

	locationID, err := posTableToolNonNegativeInt64(firstPOSTableToolValue(input, "location_id", "branch_id", "location", "branch"))
	if err != nil {
		return listPOSTablesToolInput{}, fmt.Errorf("invalid location_id: %w", err)
	}

	status := normalizePOSTableToolStatus(firstPOSTableToolValue(input, "status"))
	if strings.TrimSpace(posTableToolString(input["status"])) != "" && status == "" {
		return listPOSTablesToolInput{}, errors.New("status must be available, occupied, reserved, cleaning, or all")
	}

	minSeats, err := posTableToolNonNegativeInt64(firstPOSTableToolValue(input, "min_seats", "seats"))
	if err != nil {
		return listPOSTablesToolInput{}, fmt.Errorf("invalid min_seats: %w", err)
	}

	page, err := posTableToolNonNegativeInt64(firstPOSTableToolValue(input, "page"))
	if err != nil {
		return listPOSTablesToolInput{}, fmt.Errorf("invalid page: %w", err)
	}
	if page == 0 {
		page = 1
	}

	limit, err := posTableToolNonNegativeInt64(firstPOSTableToolValue(input, "limit", "page_size"))
	if err != nil {
		return listPOSTablesToolInput{}, fmt.Errorf("invalid limit: %w", err)
	}
	if limit == 0 {
		limit = defaultListPOSTablesToolLimit
	}
	if limit > maxListPOSTablesToolLimit {
		limit = maxListPOSTablesToolLimit
	}

	return listPOSTablesToolInput{
		Query:      query,
		LocationID: locationID,
		Status:     status,
		MinSeats:   minSeats,
		Page:       page,
		Limit:      limit,
	}, nil
}

func filterPOSTablesForTool(tables []domain.Table, input listPOSTablesToolInput) []domain.Table {
	if input.Query == "" && input.LocationID == 0 && input.Status == "" && input.MinSeats == 0 {
		return tables
	}

	query := strings.ToLower(input.Query)
	out := make([]domain.Table, 0, len(tables))
	for _, table := range tables {
		if input.LocationID > 0 && (table.LocationID == nil || *table.LocationID != input.LocationID) {
			continue
		}
		if input.Status != "" && !strings.EqualFold(table.Status, input.Status) {
			continue
		}
		if input.MinSeats > 0 && int64(table.Seats) < input.MinSeats {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(table.Label), query) {
			continue
		}
		out = append(out, table)
	}
	return out
}

func posTableToolPayload(table domain.Table) map[string]any {
	return map[string]any{
		"id":           table.ID,
		"owner_id":     table.OwnerID,
		"location_id":  posTableToolOptionalID(table.LocationID),
		"label":        table.Label,
		"seats":        table.Seats,
		"status":       table.Status,
		"is_available": table.IsAvailable(),
		"is_occupied":  table.IsOccupied(),
		"created_at":   formatPOSTableToolTime(table.CreatedAt),
		"updated_at":   formatPOSTableToolTime(table.UpdatedAt),
	}
}

func normalizePOSTableToolStatus(value any) string {
	status := strings.ToLower(strings.TrimSpace(posTableToolString(value)))
	status = strings.ReplaceAll(status, "-", "_")
	status = strings.ReplaceAll(status, " ", "_")
	switch status {
	case "", "all", "any":
		return ""
	case "available", "occupied", "reserved", "cleaning":
		return status
	default:
		return ""
	}
}

func firstPOSTableToolValue(input map[string]any, keys ...string) any {
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

func posTableToolNonNegativeInt64(value any) (int64, error) {
	out, err := posTableToolInt64(value)
	if err != nil {
		return 0, err
	}
	if out < 0 {
		return 0, errors.New("must be non-negative")
	}
	return out, nil
}

func posTableToolInt64(value any) (int64, error) {
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

func posTableToolString(value any) string {
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

func posTableToolOptionalID(value *int64) any {
	if value == nil || *value <= 0 {
		return nil
	}
	return *value
}

func formatPOSTableToolTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value.UTC().Format(posTableToolTimestampDate)
}
