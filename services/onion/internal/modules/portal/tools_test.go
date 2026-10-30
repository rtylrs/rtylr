package portal

import (
	"context"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/voxire/onion/internal/modules/portal/internal/catalog"
	"github.com/voxire/onion/internal/modules/portal/internal/layout"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	portalv1 "github.com/voxire/voxie/gen/go/hub/portal/v1"
)

type memoryStore struct {
	mu      sync.Mutex
	layouts map[int64]layout.Layout
}

func newMemoryStore() *memoryStore { return &memoryStore{layouts: map[int64]layout.Layout{}} }

func (s *memoryStore) Layout(_ context.Context, userID int64) (layout.Layout, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.layouts[userID], nil
}

func (s *memoryStore) UpdateLayout(_ context.Context, userID int64, edit func(layout.Layout) (layout.Layout, error)) (layout.Layout, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	next, err := edit(s.layouts[userID])
	if err != nil {
		return layout.Layout{}, err
	}
	if next.Empty() {
		delete(s.layouts, userID)
	} else {
		s.layouts[userID] = next
	}
	return next, nil
}

func signedInAs(userID int64) context.Context {
	return authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{UserID: userID, BusinessID: 1})
}

func TestCaretMovesItselfToTheTopAndRestoresTheDefault(t *testing.T) {
	store := newMemoryStore()
	m := &Module{store: store, nav: catalog.Navigation()}
	ctx := signedInAs(7)

	out, executed, err := m.ExecuteCaretTool(ctx, "portal.customize", map[string]any{"area": "app bar", "id": "caret", "group": "top", "position": "first"})
	if err != nil || !executed {
		t.Fatalf("customize: executed=%v err=%v", executed, err)
	}
	if first := out["items"].([]map[string]any)[0]; first["id"] != "caret" || first["group"] != "top" {
		t.Fatalf("first item = %v, want Caret at the top", first)
	}
	got, _ := (&layoutAPI{m}).GetPortalLayout(ctx, &portalv1.GetPortalLayoutRequest{})
	if !got.Layout.Rail.Customized || got.Layout.Rail.Top[0] != "caret" || strings.Join(got.Layout.Rail.Bottom, ",") != "settings,profile" {
		t.Fatalf("the Portal reads %+v", got.Layout.Rail)
	}
	if got.Layout.Workspace.Customized {
		t.Fatal("moving Caret must not touch the workspace sidebar")
	}
	if _, ok := store.layouts[8]; ok {
		t.Fatal("another person's layout changed")
	}

	if _, _, err := m.ExecuteCaretTool(ctx, "portal.restore_default", map[string]any{"area": "all"}); err != nil {
		t.Fatal(err)
	}
	if _, ok := store.layouts[7]; ok {
		t.Fatal("restoring the default should forget the saved layout")
	}
}

func TestCustomizeInfersTheAreaAndHidesPages(t *testing.T) {
	m := &Module{store: newMemoryStore(), nav: catalog.Navigation()}
	out, _, err := m.ExecuteCaretTool(signedInAs(7), "portal.customize", map[string]any{"hide": []any{"Audit Log"}})
	if err != nil {
		t.Fatal(err)
	}
	if out["area"] != layout.Workspace {
		t.Fatalf("area = %v, want the workspace sidebar", out["area"])
	}
	_, _, err = m.ExecuteCaretTool(signedInAs(7), "portal.customize", map[string]any{"area": "rail", "hide": []any{"home"}})
	if err == nil || !strings.Contains(err.Error(), "can't be hidden") {
		t.Fatalf("hiding Home: got %v", err)
	}
}

func TestSaveRepairsWhatTheClientSends(t *testing.T) {
	m := &Module{store: newMemoryStore(), nav: catalog.Navigation()}
	api := &layoutAPI{m}
	ctx := signedInAs(7)
	def, _ := m.nav.Area(layout.Rail)
	sent := def.Default()
	sent.Top = append([]string{"crm", "made-up"}, slices.DeleteFunc(slices.Clone(sent.Top), func(id string) bool { return id == "crm" })...)
	res, err := api.SavePortalLayout(ctx, &portalv1.SavePortalLayoutRequest{Rail: &portalv1.PortalArrangement{Top: sent.Top, Bottom: sent.Bottom, Hidden: []string{"home"}}})
	if err != nil {
		t.Fatal(err)
	}
	if rail := res.Layout.Rail; rail.Top[0] != "crm" || slices.Contains(rail.Top, "made-up") || len(rail.Hidden) != 0 {
		t.Fatalf("saved %+v", rail)
	}
	res, err = api.SavePortalLayout(ctx, &portalv1.SavePortalLayoutRequest{Rail: &portalv1.PortalArrangement{Top: def.Default().Top, Bottom: def.Default().Bottom}})
	if err != nil || res.Layout.Rail.Customized {
		t.Fatalf("saving the default arrangement should read as not customized: %+v %v", res.Layout.Rail, err)
	}
}

func TestToolsNeedASignedInPerson(t *testing.T) {
	m := &Module{store: newMemoryStore(), nav: catalog.Navigation()}
	if _, executed, err := m.ExecuteCaretTool(context.Background(), "portal.layout", nil); !executed || err == nil {
		t.Fatalf("executed=%v err=%v", executed, err)
	}
	if _, executed, _ := m.ExecuteCaretTool(context.Background(), "crm.list_customers", nil); executed {
		t.Fatal("portal must leave other modules' tools alone")
	}
}

func TestFindPageHandsCaretACardToShow(t *testing.T) {
	out := findPage(map[string]any{"query": "add team members"})
	pages := out["pages"].([]map[string]any)
	if len(pages) == 0 {
		t.Fatal("no page for adding team members")
	}
	block, _ := pages[0]["page_block"].(string)
	if !strings.Contains(block, `"path":"/team`) || !strings.Contains(block, `"title":`) {
		t.Fatalf("page_block = %s", block)
	}
	if empty := findPage(map[string]any{"query": "zzz"}); empty["count"] != 0 || empty["message"] == nil {
		t.Fatalf("no match should say so: %v", empty)
	}
}

func TestParallelChangesBothLand(t *testing.T) {
	m := &Module{store: newMemoryStore(), nav: catalog.Navigation()}
	ctx := signedInAs(7)
	var wg sync.WaitGroup
	for _, input := range []map[string]any{
		{"area": "rail", "id": "caret", "group": "top", "position": "first"},
		{"area": "workspace", "hide": []any{"auditLog"}},
	} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, _, err := m.ExecuteCaretTool(ctx, "portal.customize", input); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	got, _ := (&layoutAPI{m}).GetPortalLayout(ctx, &portalv1.GetPortalLayoutRequest{})
	if got.Layout.Rail.Top[0] != "caret" || !slices.Equal(got.Layout.Workspace.Hidden, []string{"auditLog"}) {
		t.Fatalf("a change was lost: %+v", got.Layout)
	}
}

// People call the whole left side "the sidebar": resetting it puts Caret back
// at the bottom too, and moving Caret "in the sidebar" moves it in the app bar.
func TestTheSidebarMeansTheWholeLeftSide(t *testing.T) {
	store := newMemoryStore()
	m := &Module{store: store, nav: catalog.Navigation()}
	ctx := signedInAs(7)
	if _, _, err := m.ExecuteCaretTool(ctx, "portal.customize", map[string]any{"area": "sidebar", "id": "caret", "position": "first", "group": "top"}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := m.ExecuteCaretTool(ctx, "portal.customize", map[string]any{"area": "sidebar", "hide": []any{"auditLog"}}); err != nil {
		t.Fatal(err)
	}
	out, _, err := m.ExecuteCaretTool(ctx, "portal.restore_default", map[string]any{"area": "sidebar"})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := store.layouts[7]; ok {
		t.Fatalf("resetting the sidebar left %+v", store.layouts[7])
	}
	if msg, _ := out["message"].(string); !strings.Contains(msg, "app bar") || !strings.Contains(msg, "workspace sidebar") {
		t.Fatalf("message = %q", msg)
	}
	if _, _, err := m.ExecuteCaretTool(ctx, "portal.customize", map[string]any{"area": "home sidebar", "hide": []any{"auditLog"}}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := m.ExecuteCaretTool(ctx, "portal.restore_default", map[string]any{"area": "home sidebar"}); err != nil {
		t.Fatal(err)
	}
	if _, ok := store.layouts[7]; ok {
		t.Fatal("naming the Home sidebar should reset it")
	}
}

func TestChangesReportTheOrderOnScreen(t *testing.T) {
	m := &Module{store: newMemoryStore(), nav: catalog.Navigation()}
	out, _, err := m.ExecuteCaretTool(signedInAs(7), "portal.customize", map[string]any{
		"area":  "rail",
		"moves": []any{map[string]any{"id": "flow", "position": "top"}},
		"hide":  []any{"pos"},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := "App bar from the top: Work, Home, Insights, CRM, Operations, Finance, HR, Automations; pinned at the bottom: Caret, Settings, Profile; hidden: POS."
	if out["order"] != want || !strings.HasSuffix(out["message"].(string), want) {
		t.Fatalf("order = %q\nmessage = %q", out["order"], out["message"])
	}
}

// "Move the user profile to the top": the avatar menu and Settings arrange
// like any app, but stay reachable.
func TestProfileAndSettingsMoveButNeverHide(t *testing.T) {
	m := &Module{store: newMemoryStore(), nav: catalog.Navigation()}
	ctx := signedInAs(7)
	out, _, err := m.ExecuteCaretTool(ctx, "portal.customize", map[string]any{"id": "Profile", "group": "top", "position": "first"})
	if err != nil {
		t.Fatal(err)
	}
	if want := "App bar from the top: Profile, Home,"; !strings.HasPrefix(out["order"].(string), want) {
		t.Fatalf("order = %q", out["order"])
	}
	if _, _, err := m.ExecuteCaretTool(ctx, "portal.customize", map[string]any{"area": "rail", "hide": []any{"settings"}}); err == nil {
		t.Fatal("hid Settings")
	}
}

func TestOneRequestCanChangeBothAreas(t *testing.T) {
	m := &Module{store: newMemoryStore(), nav: catalog.Navigation()}
	ctx := signedInAs(7)
	out, _, err := m.ExecuteCaretTool(ctx, "portal.customize", map[string]any{"hide": []any{"pos", "auditLog"}})
	if err != nil {
		t.Fatal(err)
	}
	if order, _ := out["order"].(string); len(out["areas"].([]map[string]any)) != 2 || !strings.Contains(order, "hidden: POS") || !strings.Contains(order, "hidden: Audit Log") {
		t.Fatalf("got %v", out)
	}
	// "Home" names an item in both areas; its neighbour says which one.
	out, _, err = m.ExecuteCaretTool(ctx, "portal.customize", map[string]any{"area": "sidebar", "id": "Home", "position": "after:team"})
	if err != nil {
		t.Fatal(err)
	}
	if out["area"] != layout.Workspace || !strings.HasPrefix(out["order"].(string), "Workspace sidebar from the top: Team, Home,") {
		t.Fatalf("got %v", out["order"])
	}
	if _, _, err := m.ExecuteCaretTool(ctx, "portal.customize", map[string]any{"id": "payroll"}); err == nil {
		t.Fatal("an unknown item should be refused")
	}
}

func TestFractionalSlotsAreRefused(t *testing.T) {
	m := &Module{store: newMemoryStore(), nav: catalog.Navigation()}
	if _, _, err := m.ExecuteCaretTool(signedInAs(7), "portal.customize", map[string]any{"area": "rail", "id": "flow", "position": 2.5}); err == nil {
		t.Fatal("slot 2.5 moved an item")
	}
	out, _, err := m.ExecuteCaretTool(signedInAs(7), "portal.customize", map[string]any{"area": "rail", "id": "flow", "position": float64(2)})
	if err != nil || !strings.HasPrefix(out["order"].(string), "App bar from the top: Home, Work,") {
		t.Fatalf("slot 2: %v %v", out["order"], err)
	}
}

func TestAnEmptyChangeIsRefused(t *testing.T) {
	store := newMemoryStore()
	m := &Module{store: store, nav: catalog.Navigation()}
	for _, input := range []map[string]any{{}, {"area": "sidebar"}, {"area": "rail"}} {
		if _, _, err := m.ExecuteCaretTool(signedInAs(7), "portal.customize", input); err == nil || !strings.Contains(err.Error(), "nothing to change") {
			t.Errorf("%v: got %v", input, err)
		}
	}
}

// The 2026-10-30 report: Profile went to the top, then "settings above
// profile" kept Settings in the bottom group and still answered "Done".
func TestAMoveThatChangesNothingFails(t *testing.T) {
	m := &Module{store: newMemoryStore(), nav: catalog.Navigation()}
	ctx := signedInAs(7)
	if _, _, err := m.ExecuteCaretTool(ctx, "portal.customize", map[string]any{"id": "profile", "group": "top", "position": "first"}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := m.ExecuteCaretTool(ctx, "portal.customize", map[string]any{"id": "caret", "position": "first"}); err == nil || !strings.Contains(err.Error(), "nothing changed") {
		t.Fatalf("no-op move answered err = %v", err)
	}
	out, _, err := m.ExecuteCaretTool(ctx, "portal.customize", map[string]any{"id": "settings", "position": "before:profile"})
	if err != nil {
		t.Fatal(err)
	}
	if want := "App bar from the top: Settings, Profile, Home,"; !strings.HasPrefix(out["order"].(string), want) {
		t.Fatalf("order = %q", out["order"])
	}
}

func TestToTheTopMeansTheTopOfTheBar(t *testing.T) {
	m := &Module{store: newMemoryStore(), nav: catalog.Navigation()}
	out, _, err := m.ExecuteCaretTool(signedInAs(7), "portal.customize", map[string]any{"area": "rail", "id": "settings", "position": "top"})
	if err != nil {
		t.Fatal(err)
	}
	if want := "App bar from the top: Settings, Home,"; !strings.HasPrefix(out["order"].(string), want) {
		t.Fatalf("order = %q", out["order"])
	}
}
