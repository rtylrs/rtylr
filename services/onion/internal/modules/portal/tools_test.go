package portal

import (
	"context"
	"slices"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/modules/portal/internal/catalog"
	"github.com/voxire/onion/internal/modules/portal/internal/layout"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	portalv1 "github.com/voxire/voxie/gen/go/hub/portal/v1"
)

type memoryStore map[int64]layout.Layout

func (s memoryStore) Layout(_ context.Context, userID int64) (layout.Layout, error) {
	return s[userID], nil
}

func (s memoryStore) SaveLayout(_ context.Context, userID int64, l layout.Layout) error {
	if l.Empty() {
		delete(s, userID)
	} else {
		s[userID] = l
	}
	return nil
}

func signedInAs(userID int64) context.Context {
	return authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{UserID: userID, BusinessID: 1})
}

func TestCaretMovesItselfToTheTopAndRestoresTheDefault(t *testing.T) {
	store := memoryStore{}
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
	if !got.Layout.Rail.Customized || got.Layout.Rail.Top[0] != "caret" || len(got.Layout.Rail.Bottom) != 0 {
		t.Fatalf("the Portal reads %+v", got.Layout.Rail)
	}
	if got.Layout.Workspace.Customized {
		t.Fatal("moving Caret must not touch the workspace sidebar")
	}
	if _, ok := store[8]; ok {
		t.Fatal("another person's layout changed")
	}

	if _, _, err := m.ExecuteCaretTool(ctx, "portal.restore_default", map[string]any{"area": "all"}); err != nil {
		t.Fatal(err)
	}
	if _, ok := store[7]; ok {
		t.Fatal("restoring the default should forget the saved layout")
	}
}

func TestCustomizeInfersTheAreaAndHidesPages(t *testing.T) {
	m := &Module{store: memoryStore{}, nav: catalog.Navigation()}
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
	m := &Module{store: memoryStore{}, nav: catalog.Navigation()}
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
	m := &Module{store: memoryStore{}, nav: catalog.Navigation()}
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
