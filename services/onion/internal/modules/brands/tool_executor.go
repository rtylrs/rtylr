package brands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/voxire/onion/internal/modules/brands/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/tenant"
)

const createBrandToolName = "erp.create_brand"

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	if name != createBrandToolName {
		return nil, false, nil
	}
	output, err := m.executeCreateBrandTool(ctx, input)
	return output, true, err
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

	ownerID := tenant.ResolveOwnerID(ctx, authUser.BusinessID, authUser.UserID)
	id, err := m.service.CreateBrand(ctx, ownerID, parsed)
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
