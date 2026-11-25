package hubmenu

import (
	"context"
	"testing"

	authPkg "github.com/voxire/onion/internal/platform/auth"
)

func TestExecuteCaretTool_UnknownToolReturnsNotExecuted(t *testing.T) {
	m := &Module{}
	out, executed, err := m.ExecuteCaretTool(context.Background(), "pos.some_other_tool", nil)
	if err != nil || executed || out != nil {
		t.Fatalf("expected (nil, false, nil), got (%v, %v, %v)", out, executed, err)
	}
}

func TestExecuteCaretTool_RequiresAuthAndBrandingPermission(t *testing.T) {
	m := &Module{}
	if _, executed, err := m.ExecuteCaretTool(context.Background(), toolGetMenuDesign, nil); !executed || err == nil {
		t.Fatalf("expected auth error for unauthenticated call, got executed=%v err=%v", executed, err)
	}

	ctxWithoutPerm := authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{
		UserID:     1,
		BusinessID: 42,
	})
	if _, executed, err := m.ExecuteCaretTool(ctxWithoutPerm, toolUpdateMenuDesign, nil); !executed || err == nil {
		t.Fatalf("expected permission error without menu:branding:edit, got executed=%v err=%v", executed, err)
	}
}

func TestDeepMergeMap_MergesNestedThemeTokens(t *testing.T) {
	dst := map[string]any{
		"colors": map[string]any{"primary": "#111111", "accent": "#4583C4"},
		"mode":   "light",
	}
	src := map[string]any{
		"colors": map[string]any{"accent": "#c8a96e"},
		"mode":   "dark",
	}
	deepMergeMap(dst, src)

	colors := dst["colors"].(map[string]any)
	if colors["primary"] != "#111111" || colors["accent"] != "#c8a96e" || dst["mode"] != "dark" {
		t.Fatalf("unexpected merged map: %+v", dst)
	}
}
