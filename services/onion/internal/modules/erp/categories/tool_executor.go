package categories

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/erp/categories/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/tenant"
)

const (
	createCategoryToolName = "erp.create_category"
	listCategoriesToolName = "erp.list_categories"
)

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	switch name {
	case listCategoriesToolName:
		output, err := m.executeListCategoriesTool(ctx)
		return output, true, err
	case createCategoryToolName:
		output, err := m.executeCreateCategoryTool(ctx, input)
		return output, true, err
	default:
		return nil, false, nil
	}
}

func (m *Module) executeListCategoriesTool(ctx context.Context) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, bizErr
	}

	categories, err := m.service.ListCategories(ctx, businessID)
	if err != nil {
		return nil, err
	}

	payloads := make([]map[string]any, 0, len(categories))
	for _, category := range categories {
		payloads = append(payloads, categoryToolPayload(category))
	}

	return map[string]any{
		"categories": payloads,
		"returned":   len(payloads),
	}, nil
}

func (m *Module) executeCreateCategoryTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	parsed, err := parseCreateCategoryToolInput(input)
	if err != nil {
		return nil, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, bizErr
	}

	id, err := m.service.CreateCategory(ctx, businessID, parsed)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"category": map[string]any{
			"id":          id,
			"name":        parsed.Name,
			"description": parsed.Description,
		},
	}, nil
}

func parseCreateCategoryToolInput(input map[string]any) (domain.CategoryInput, error) {
	parsed := domain.CategoryInput{
		Name:        strings.TrimSpace(categoryToolString(firstCategoryToolValue(input, "name", "category_name"))),
		Description: strings.TrimSpace(categoryToolString(firstCategoryToolValue(input, "description"))),
	}
	if parsed.Name == "" {
		return domain.CategoryInput{}, errors.New("name is required")
	}
	return parsed, nil
}

func firstCategoryToolValue(input map[string]any, keys ...string) any {
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

func categoryToolPayload(category domain.Category) map[string]any {
	return map[string]any{
		"id":          category.ID,
		"name":        category.Name,
		"description": category.Description,
		"created_at":  categoryToolTime(category.CreatedAt),
		"updated_at":  categoryToolTime(category.UpdatedAt),
	}
}

func categoryToolString(value any) string {
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

func categoryToolTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}
