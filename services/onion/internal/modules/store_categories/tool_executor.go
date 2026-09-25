package store_categories

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/store_categories/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/pagination"
	"github.com/voxire/onion/internal/platform/tenant"
)

const (
	listStoreCategoriesToolName         = "store.list_categories"
	defaultListStoreCategoriesToolLimit = int64(20)
	maxListStoreCategoriesToolLimit     = int64(50)
	storeCategoryToolTimestampDate      = "2006-01-02T15:04:05Z"
)

type listStoreCategoriesToolInput struct {
	Query    string
	ParentID int64
	Page     int64
	Limit    int64
}

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	switch name {
	case listStoreCategoriesToolName:
		output, err := m.executeListStoreCategoriesTool(ctx, input)
		return output, true, err
	default:
		return nil, false, nil
	}
}

func (m *Module) executeListStoreCategoriesTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, errors.New("no business scope for request")
	}

	parsed, err := parseListStoreCategoriesToolInput(input)
	if err != nil {
		return nil, err
	}

	categories, err := m.service.List(ctx, businessID)
	if err != nil {
		return nil, err
	}
	categories = filterStoreCategoriesForTool(categories, parsed)

	start := pagination.SliceStart(parsed.Page, parsed.Limit, len(categories))
	if start > len(categories) {
		start = len(categories)
	}
	end := start + int(parsed.Limit)
	if end > len(categories) {
		end = len(categories)
	}

	payloads := make([]map[string]any, 0, end-start)
	for _, category := range categories[start:end] {
		payloads = append(payloads, storeCategoryToolPayload(category))
	}

	return map[string]any{
		"categories": payloads,
		"total":      len(categories),
		"returned":   len(payloads),
		"page":       parsed.Page,
		"page_size":  parsed.Limit,
		"has_more":   end < len(categories),
		"filters": map[string]any{
			"query":     parsed.Query,
			"parent_id": parsed.ParentID,
		},
	}, nil
}

func parseListStoreCategoriesToolInput(input map[string]any) (listStoreCategoriesToolInput, error) {
	query := strings.TrimSpace(storeCategoryToolString(firstStoreCategoryToolValue(input, "query", "q", "search", "name")))

	parentID, err := storeCategoryToolNonNegativeInt64(firstStoreCategoryToolValue(input, "parent_id", "parent"))
	if err != nil {
		return listStoreCategoriesToolInput{}, fmt.Errorf("invalid parent_id: %w", err)
	}

	page, err := storeCategoryToolNonNegativeInt64(firstStoreCategoryToolValue(input, "page"))
	if err != nil {
		return listStoreCategoriesToolInput{}, fmt.Errorf("invalid page: %w", err)
	}
	if page == 0 {
		page = 1
	}

	limit, err := storeCategoryToolNonNegativeInt64(firstStoreCategoryToolValue(input, "limit", "page_size"))
	if err != nil {
		return listStoreCategoriesToolInput{}, fmt.Errorf("invalid limit: %w", err)
	}
	if limit == 0 {
		limit = defaultListStoreCategoriesToolLimit
	}
	if limit > maxListStoreCategoriesToolLimit {
		limit = maxListStoreCategoriesToolLimit
	}

	return listStoreCategoriesToolInput{
		Query:    query,
		ParentID: parentID,
		Page:     page,
		Limit:    limit,
	}, nil
}

func filterStoreCategoriesForTool(categories []domain.StoreCategory, input listStoreCategoriesToolInput) []domain.StoreCategory {
	if input.Query == "" && input.ParentID == 0 {
		return categories
	}

	query := strings.ToLower(input.Query)
	out := make([]domain.StoreCategory, 0, len(categories))
	for _, category := range categories {
		if input.ParentID > 0 && category.ParentID != input.ParentID {
			continue
		}
		if query != "" && !storeCategoryToolMatchesQuery(category, query) {
			continue
		}
		out = append(out, category)
	}
	return out
}

func storeCategoryToolMatchesQuery(category domain.StoreCategory, query string) bool {
	return strings.Contains(strings.ToLower(category.Name), query) ||
		strings.Contains(strings.ToLower(category.Slug), query) ||
		strings.Contains(strings.ToLower(category.Description), query)
}

func storeCategoryToolPayload(category domain.StoreCategory) map[string]any {
	return map[string]any{
		"id":          category.ID,
		"owner_id":    category.OwnerID,
		"name":        category.Name,
		"slug":        category.Slug,
		"description": category.Description,
		"image_url":   category.ImageURL,
		"parent_id":   category.ParentID,
		"sort_order":  category.SortOrder,
		"created_at":  formatStoreCategoryToolTime(category.CreatedAt),
	}
}

func firstStoreCategoryToolValue(input map[string]any, keys ...string) any {
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

func storeCategoryToolNonNegativeInt64(value any) (int64, error) {
	out, err := storeCategoryToolInt64(value)
	if err != nil {
		return 0, err
	}
	if out < 0 {
		return 0, errors.New("must be non-negative")
	}
	return out, nil
}

func storeCategoryToolInt64(value any) (int64, error) {
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

func storeCategoryToolString(value any) string {
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

func formatStoreCategoryToolTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value.UTC().Format(storeCategoryToolTimestampDate)
}
