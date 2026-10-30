package layout

import (
	"slices"
	"strings"
	"testing"
)

var rail = Area{Name: "App bar", Items: []Item{
	{ID: "home", Label: "Home", Required: true},
	{ID: "insights", Label: "Insights"},
	{ID: "crm", Label: "CRM"},
	{ID: "automations", Label: "Automations"},
	{ID: "caret", Label: "Caret", Bottom: true},
}}

var sidebar = Area{Name: "Workspace sidebar", Items: []Item{
	{ID: "dashboard", Label: "Home", Required: true},
	{ID: "team", Label: "Team"},
	{ID: "auditLog", Label: "Audit Log"},
}}

func TestMoveCaretToTheTop(t *testing.T) {
	got, err := rail.Apply(rail.Default(), Change{Moves: []Move{{ID: "caret", Group: "top", Position: "first"}}})
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"caret", "home", "insights", "crm", "automations"}; !slices.Equal(got.Top, want) || len(got.Bottom) != 0 {
		t.Fatalf("got %+v, want top %v and an empty bottom", got, want)
	}
	if rail.IsDefault(got) {
		t.Fatal("a moved Caret is not the default")
	}
}

func TestMoveRelativeToAnotherItemTakesItsGroup(t *testing.T) {
	got, err := rail.Apply(rail.Default(), Change{Moves: []Move{{ID: "automations", Position: "after:Caret"}}})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got.Bottom, []string{"caret", "automations"}) || slices.Contains(got.Top, "automations") {
		t.Fatalf("got %+v", got)
	}
	got, err = rail.Apply(got, Change{Moves: []Move{{ID: "crm", Position: "before:home"}}})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got.Top, []string{"crm", "home", "insights"}) {
		t.Fatalf("got %+v", got)
	}
}

func TestHideAndShowKeepThePlace(t *testing.T) {
	hidden, err := sidebar.Apply(sidebar.Default(), Change{Hide: []string{"Audit Log"}})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(hidden.Hidden, []string{"auditLog"}) || !slices.Equal(hidden.Top, sidebar.Default().Top) {
		t.Fatalf("got %+v", hidden)
	}
	shown, err := sidebar.Apply(hidden, Change{Show: []string{"auditlog"}})
	if err != nil {
		t.Fatal(err)
	}
	if !sidebar.IsDefault(shown) || sidebar.Stored(shown) != nil {
		t.Fatalf("showing it again should restore the default, got %+v", shown)
	}
}

func TestRefusesWhatCantBeDone(t *testing.T) {
	cases := map[string]Change{
		"can't be hidden":    {Hide: []string{"home"}},
		"has no item":        {Moves: []Move{{ID: "payroll"}}},
		"no bottom group":    {Moves: []Move{{ID: "team", Group: "bottom"}}},
		"relative to itself": {Moves: []Move{{ID: "team", Position: "after:team"}}},
		"nothing to change":  {},
	}
	area := map[string]Area{"can't be hidden": rail, "has no item": rail}
	for want, change := range cases {
		a, ok := area[want]
		if !ok {
			a = sidebar
		}
		if _, err := a.Apply(a.Default(), change); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%+v: got %v, want an error containing %q", change, err, want)
		}
	}
}

func TestResolveRepairsASavedArrangement(t *testing.T) {
	got := rail.Resolve(&Arrangement{
		Top:    []string{"crm", "gone", "home", "crm"},
		Hidden: []string{"home", "crm"},
	})
	// insights and automations were never saved: each goes after the item
	// that precedes it by default (home and crm); Caret keeps its bottom place.
	if want := []string{"crm", "automations", "home", "insights"}; !slices.Equal(got.Top, want) {
		t.Fatalf("top = %v, want %v", got.Top, want)
	}
	if !slices.Equal(got.Bottom, []string{"caret"}) || !slices.Equal(got.Hidden, []string{"crm"}) {
		t.Fatalf("got %+v", got)
	}
}

func TestMoveToASlot(t *testing.T) {
	got, err := rail.Apply(rail.Default(), Change{Moves: []Move{{ID: "crm", Position: "2"}}})
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"home", "crm", "insights", "automations"}; !slices.Equal(got.Top, want) {
		t.Fatalf("top = %v, want %v", got.Top, want)
	}
	if got, _ := rail.Apply(rail.Default(), Change{Moves: []Move{{ID: "crm", Position: "40"}}}); got.Top[len(got.Top)-1] != "crm" {
		t.Fatalf("a slot past the end should mean last, got %v", got.Top)
	}
	if _, err := rail.Apply(rail.Default(), Change{Moves: []Move{{ID: "crm", Position: "0"}}}); err == nil {
		t.Fatal("slot 0 should be refused")
	}
}
