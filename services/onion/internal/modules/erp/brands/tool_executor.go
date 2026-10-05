package brands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/erp/brands/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/tenant"
)

const (
	createBrandToolName = "erp.create_brand"
	listBrandsToolName  = "erp.list_brands"
)

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	switch name {
	case listBrandsToolName:
		output, err := m.executeListBrandsTool(ctx)
		return output, true, err
	case createBrandToolName:
		output, err := m.executeCreateBrandTool(ctx, input)
		return output, true, err
	default:
		return nil, false, nil
	}
}

func (m *Module) executeListBrandsTool(ctx context.Context) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, bizErr
	}

	brands, err := m.service.ListBrands(ctx, businessID)
	if err != nil {
		return nil, err
	}

	payloads := make([]map[string]any, 0, len(brands))
	for _, brand := range brands {
		payloads = append(payloads, brandToolPayload(brand))
	}

	return map[string]any{
		"brands":   payloads,
		"returned": len(payloads),
	}, nil
}

func (m *Module) executeCreateBrandTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	parsed, err := parseCreateBrandToolInput(input)
	if err != nil {
		return nil, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, bizErr
	}

	id, err := m.service.CreateBrand(ctx, businessID, parsed)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"brand": map[string]any{
			"id":          id,
			"name":        parsed.Name,
			"logo_url":    parsed.LogoURL,
			"description": parsed.Description,
		},
	}, nil
}

func parseCreateBrandToolInput(input map[string]any) (domain.BrandInput, error) {
	parsed := domain.BrandInput{
		Name:        strings.TrimSpace(brandToolString(firstBrandToolValue(input, "name", "brand_name"))),
		LogoURL:     strings.TrimSpace(brandToolString(firstBrandToolValue(input, "logo_url", "logo"))),
		Description: strings.TrimSpace(brandToolString(firstBrandToolValue(input, "description"))),
	}
	if parsed.Name == "" {
		return domain.BrandInput{}, errors.New("name is required")
	}
	return parsed, nil
}

func firstBrandToolValue(input map[string]any, keys ...string) any {
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

func brandToolPayload(brand domain.Brand) map[string]any {
	return map[string]any{
		"id":            brand.ID,
		"name":          brand.Name,
		"logo_url":      brand.LogoURL,
		"description":   brand.Description,
		"product_count": brand.ProductCount,
		"created_at":    brandToolTime(brand.CreatedAt),
		"updated_at":    brandToolTime(brand.UpdatedAt),
	}
}

func brandToolString(value any) string {
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

func brandToolTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}
