package pos_menu

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/pos_menu/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/pagination"
	"github.com/voxire/onion/internal/platform/tenant"
)

const (
	listPOSMenuItemsToolName         = "pos.list_menu_items"
	defaultListPOSMenuItemsToolLimit = int64(20)
	maxListPOSMenuItemsToolLimit     = int64(50)
	posMenuToolTimestampDate         = "2006-01-02T15:04:05Z"
)

type listPOSMenuItemsToolInput struct {
	Query      string
	CategoryID int64
	Category   string
	Active     *bool
	Page       int64
	Limit      int64
}

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	switch name {
	case listPOSMenuItemsToolName:
		output, err := m.executeListPOSMenuItemsTool(ctx, input)
		return output, true, err
	default:
		return nil, false, nil
	}
}

func (m *Module) executeListPOSMenuItemsTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	parsed, err := parseListPOSMenuItemsToolInput(input)
	if err != nil {
		return nil, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, errors.New("failed to resolve business scope")
	}

	items, err := m.service.List(ctx, businessID)
	if err != nil {
		return nil, err
	}
	items = filterPOSMenuItemsForTool(items, parsed)

	start := pagination.SliceStart(parsed.Page, parsed.Limit, len(items))
	if start > len(items) {
		start = len(items)
	}
	end := start + int(parsed.Limit)
	if end > len(items) {
		end = len(items)
	}

	payloads := make([]map[string]any, 0, end-start)
	for _, item := range items[start:end] {
		payloads = append(payloads, posMenuItemToolPayload(item))
	}

	return map[string]any{
		"items":     payloads,
		"total":     len(items),
		"returned":  len(payloads),
		"page":      parsed.Page,
		"page_size": parsed.Limit,
		"has_more":  end < len(items),
		"filters": map[string]any{
			"query":       parsed.Query,
			"category_id": parsed.CategoryID,
			"category":    parsed.Category,
			"active":      posMenuToolOptionalBool(parsed.Active),
		},
	}, nil
}

func parseListPOSMenuItemsToolInput(input map[string]any) (listPOSMenuItemsToolInput, error) {
	query := strings.TrimSpace(posMenuToolString(firstPOSMenuToolValue(input, "query", "q", "search", "name")))

	categoryID, err := posMenuToolNonNegativeInt64(firstPOSMenuToolValue(input, "category_id", "categoryId"))
	if err != nil {
		return listPOSMenuItemsToolInput{}, fmt.Errorf("invalid category_id: %w", err)
	}
	category := strings.TrimSpace(posMenuToolString(firstPOSMenuToolValue(input, "category", "category_name")))

	active, err := posMenuToolOptionalBoolInput(firstPOSMenuToolValue(input, "active", "is_active", "available", "status"))
	if err != nil {
		return listPOSMenuItemsToolInput{}, fmt.Errorf("invalid active: %w", err)
	}

	page, err := posMenuToolNonNegativeInt64(firstPOSMenuToolValue(input, "page"))
	if err != nil {
		return listPOSMenuItemsToolInput{}, fmt.Errorf("invalid page: %w", err)
	}
	if page == 0 {
		page = 1
	}

	limit, err := posMenuToolNonNegativeInt64(firstPOSMenuToolValue(input, "limit", "page_size"))
	if err != nil {
		return listPOSMenuItemsToolInput{}, fmt.Errorf("invalid limit: %w", err)
	}
	if limit == 0 {
		limit = defaultListPOSMenuItemsToolLimit
	}
	if limit > maxListPOSMenuItemsToolLimit {
		limit = maxListPOSMenuItemsToolLimit
	}

	return listPOSMenuItemsToolInput{
		Query:      query,
		CategoryID: categoryID,
		Category:   category,
		Active:     active,
		Page:       page,
		Limit:      limit,
	}, nil
}

func filterPOSMenuItemsForTool(items []domain.MenuItem, input listPOSMenuItemsToolInput) []domain.MenuItem {
	if input.Query == "" && input.CategoryID == 0 && input.Category == "" && input.Active == nil {
		return items
	}

	query := strings.ToLower(input.Query)
	category := strings.ToLower(input.Category)
	out := make([]domain.MenuItem, 0, len(items))
	for _, item := range items {
		if input.CategoryID > 0 && item.CategoryID != input.CategoryID {
			continue
		}
		if category != "" && !strings.Contains(strings.ToLower(item.Category), category) {
			continue
		}
		if input.Active != nil && item.IsActive != *input.Active {
			continue
		}
		if query != "" && !posMenuItemToolMatchesQuery(item, query) {
			continue
		}
		out = append(out, item)
	}
	return out
}

func posMenuItemToolMatchesQuery(item domain.MenuItem, query string) bool {
	return strings.Contains(strings.ToLower(item.Name), query) ||
		strings.Contains(strings.ToLower(item.SKU), query) ||
		strings.Contains(strings.ToLower(item.Barcode), query) ||
		strings.Contains(strings.ToLower(item.Description), query)
}

func posMenuItemToolPayload(item domain.MenuItem) map[string]any {
	return map[string]any{
		"id":          item.ID,
		"owner_id":    item.OwnerID,
		"category_id": item.CategoryID,
		"category":    item.Category,
		"name":        item.Name,
		"sku":         item.SKU,
		"barcode":     item.Barcode,
		"sale_price":  item.SalePrice,
		"tax_rate":    item.TaxRate,
		"description": item.Description,
		"image_url":   item.ImageURL,
		"stock":       item.Stock,
		"is_active":   item.IsActive,
		"available":   item.IsAvailable(),
		"created_at":  formatPOSMenuToolTime(item.CreatedAt),
		"updated_at":  formatPOSMenuToolTime(item.UpdatedAt),
	}
}

func firstPOSMenuToolValue(input map[string]any, keys ...string) any {
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

func posMenuToolOptionalBoolInput(value any) (*bool, error) {
	switch v := value.(type) {
	case nil:
		return nil, nil
	case bool:
		return &v, nil
	case int:
		return posMenuToolOptionalBoolInt(int64(v))
	case int64:
		return posMenuToolOptionalBoolInt(v)
	case int32:
		return posMenuToolOptionalBoolInt(int64(v))
	case float64:
		if math.Trunc(v) != v {
			return nil, errors.New("must be 0 or 1")
		}
		return posMenuToolOptionalBoolInt(int64(v))
	case string:
		raw := strings.ToLower(strings.TrimSpace(v))
		switch raw {
		case "", "all", "any":
			return nil, nil
		case "true", "yes", "1", "active", "available", "enabled":
			out := true
			return &out, nil
		case "false", "no", "0", "inactive", "unavailable", "disabled":
			out := false
			return &out, nil
		default:
			return nil, errors.New("must be true, false, active, inactive, available, unavailable, or all")
		}
	case json.Number:
		i, err := v.Int64()
		if err != nil {
			return nil, err
		}
		return posMenuToolOptionalBoolInt(i)
	default:
		return nil, fmt.Errorf("unsupported value %T", value)
	}
}

func posMenuToolOptionalBoolInt(value int64) (*bool, error) {
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

func posMenuToolOptionalBool(value *bool) any {
	if value == nil {
		return nil
	}
	return *value
}

func posMenuToolNonNegativeInt64(value any) (int64, error) {
	out, err := posMenuToolInt64(value)
	if err != nil {
		return 0, err
	}
	if out < 0 {
		return 0, errors.New("must be non-negative")
	}
	return out, nil
}

func posMenuToolInt64(value any) (int64, error) {
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

func posMenuToolString(value any) string {
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

func formatPOSMenuToolTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value.UTC().Format(posMenuToolTimestampDate)
}
