package hubmenu

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/voxire/onion/internal/modules/hub_menu/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/tenant"
)

const (
	toolGetMenuDesign     = "pos.get_menu_design"
	toolUpdateMenuDesign  = "pos.update_menu_design"
	toolPublishMenuDesign = "pos.publish_menu_design"
	toolCloneMenuDesign   = "pos.clone_menu_design"
)

// ExecuteCaretTool lets Caret AI conversationally read, customize, clone, and
// publish a tenant's branded menu design on behalf of an authorized user.
func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	switch name {
	case toolGetMenuDesign:
		out, err := m.executeGetMenuDesign(ctx, input)
		return out, true, err
	case toolUpdateMenuDesign:
		out, err := m.executeUpdateMenuDesign(ctx, input)
		return out, true, err
	case toolPublishMenuDesign:
		out, err := m.executePublishMenuDesign(ctx, input)
		return out, true, err
	case toolCloneMenuDesign:
		out, err := m.executeCloneMenuDesign(ctx, input)
		return out, true, err
	default:
		return nil, false, nil
	}
}

func resolveAuthorizedMenuBusiness(ctx context.Context) (int64, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return 0, errors.New("authentication is required")
	}
	if !authPkg.HasAppPermission(authUser, authorization.AppMenu, "menu:branding:edit") {
		return 0, errors.New("menu:branding:edit permission is required")
	}
	return tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
}

func (m *Module) executeGetMenuDesign(ctx context.Context, input map[string]any) (map[string]any, error) {
	businessID, err := resolveAuthorizedMenuBusiness(ctx)
	if err != nil {
		return nil, err
	}
	locationID := parseInt64Input(input["location_id"])
	profile, err := m.service.GetMenuProfile(ctx, businessID, locationID)
	if err != nil {
		return nil, err
	}
	return profileToMap(profile), nil
}

func (m *Module) executeUpdateMenuDesign(ctx context.Context, input map[string]any) (map[string]any, error) {
	businessID, err := resolveAuthorizedMenuBusiness(ctx)
	if err != nil {
		return nil, err
	}
	locationID := parseInt64Input(input["location_id"])
	current, err := m.service.GetMenuProfile(ctx, businessID, locationID)
	if err != nil {
		return nil, err
	}

	// Merge incoming partial patch into the existing draft JSON so Caret can
	// tweak colors, fonts, layout, or feature toggles without dropping other settings.
	merged := make(map[string]any)
	_ = json.Unmarshal([]byte(current.DraftConfigJSON), &merged)
	if patchObj, ok := input["config_patch"].(map[string]any); ok {
		deepMergeMap(merged, patchObj)
	}
	if rawJSON, ok := input["draft_config_json"].(string); ok && strings.TrimSpace(rawJSON) != "" {
		var parsed map[string]any
		if err := json.Unmarshal([]byte(rawJSON), &parsed); err != nil {
			return nil, fmt.Errorf("invalid draft_config_json: %w", err)
		}
		deepMergeMap(merged, parsed)
	}
	encoded, err := json.Marshal(merged)
	if err != nil {
		return nil, err
	}

	req := domain.SaveMenuProfileDraftInput{
		LocationID:      locationID,
		DraftConfigJSON: string(encoded),
	}
	if css, ok := input["custom_css"].(string); ok {
		req.CustomCSS = &css
	}
	if loc, ok := input["default_locale"].(string); ok && strings.TrimSpace(loc) != "" {
		req.DefaultLocale = &loc
	}
	if cur, ok := input["default_currency"].(string); ok && strings.TrimSpace(cur) != "" {
		req.DefaultCurrency = &cur
	}
	if locales, ok := input["enabled_locales"].([]any); ok {
		for _, v := range locales {
			if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
				req.EnabledLocales = append(req.EnabledLocales, s)
			}
		}
	}

	saved, err := m.service.SaveMenuProfileDraft(ctx, businessID, req)
	if err != nil {
		return nil, err
	}
	return profileToMap(saved), nil
}

func (m *Module) executePublishMenuDesign(ctx context.Context, input map[string]any) (map[string]any, error) {
	businessID, err := resolveAuthorizedMenuBusiness(ctx)
	if err != nil {
		return nil, err
	}
	locationID := parseInt64Input(input["location_id"])
	published, err := m.service.PublishMenuProfile(ctx, businessID, locationID)
	if err != nil {
		return nil, err
	}
	return profileToMap(published), nil
}

func (m *Module) executeCloneMenuDesign(ctx context.Context, input map[string]any) (map[string]any, error) {
	businessID, err := resolveAuthorizedMenuBusiness(ctx)
	if err != nil {
		return nil, err
	}
	sourceID := parseInt64Input(input["source_location_id"])
	targetID := parseInt64Input(input["target_location_id"])
	cloned, err := m.service.CloneMenuProfile(ctx, businessID, sourceID, targetID)
	if err != nil {
		return nil, err
	}
	return profileToMap(cloned), nil
}

func profileToMap(p domain.MenuProfile) map[string]any {
	return map[string]any{
		"location_id":           p.LocationID,
		"draft_config_json":     p.DraftConfigJSON,
		"published_config_json": p.PublishedConfigJSON,
		"custom_css":            p.CustomCSS,
		"default_locale":        p.DefaultLocale,
		"enabled_locales":       p.EnabledLocales,
		"default_currency":      p.DefaultCurrency,
		"is_published":          p.IsPublished,
	}
}

func parseInt64Input(raw any) int64 {
	switch v := raw.(type) {
	case float64:
		return int64(v)
	case int64:
		return v
	case int:
		return int64(v)
	case string:
		n, _ := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
		return n
	default:
		return 0
	}
}

func deepMergeMap(dst, src map[string]any) {
	for k, v := range src {
		if srcMap, ok := v.(map[string]any); ok {
			if dstMap, exists := dst[k].(map[string]any); exists {
				deepMergeMap(dstMap, srcMap)
				continue
			}
		}
		dst[k] = v
	}
}
