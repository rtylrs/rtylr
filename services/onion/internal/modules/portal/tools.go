package portal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/voxire/onion/internal/modules/portal/internal/catalog"
	"github.com/voxire/onion/internal/modules/portal/internal/layout"
)

// ExecuteCaretTool runs the portal.* tools: Caret finds pages for the person
// and arranges their Portal with the same rules as the settings screen.
func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	switch name {
	case "portal.find_page":
		return findPage(input), true, nil
	case "portal.layout", "portal.customize", "portal.restore_default":
	default:
		return nil, false, nil
	}
	userID, err := signedIn(ctx)
	if err != nil {
		return nil, true, errors.New("sign in to arrange the Portal")
	}
	switch name {
	case "portal.layout":
		saved, err := m.store.Layout(ctx, userID)
		if err != nil {
			return nil, true, err
		}
		out := []map[string]any{}
		for _, area := range areas {
			out = append(out, m.areaOutput(area, m.arrangements(saved)[area], saved.Saved(area) != nil))
		}
		return map[string]any{"areas": out}, true, nil
	case "portal.customize":
		area := areaName(input)
		change, err := changeFromInput(input)
		if err != nil {
			return nil, true, err
		}
		if area == "all" {
			area = m.inferArea(change)
		}
		saved, next, err := m.change(ctx, userID, area, func(a layout.Area, current layout.Arrangement) (layout.Arrangement, error) {
			return a.Apply(current, change)
		})
		if err != nil {
			return nil, true, err
		}
		out := m.areaOutput(area, next, saved.Saved(area) != nil)
		out["message"] = "Done. The Portal shows the change now."
		return out, true, nil
	default:
		area := areaName(input)
		if area == "all" {
			area = ""
		}
		if _, ok := m.nav.Area(area); area != "" && !ok {
			return nil, true, fmt.Errorf("unknown area %q; use rail, workspace or all", area)
		}
		saved, err := m.reset(ctx, userID, area)
		if err != nil {
			return nil, true, err
		}
		out, names := []map[string]any{}, []string{}
		for _, name := range areas {
			if area == "" || area == name {
				out = append(out, m.areaOutput(name, m.arrangements(saved)[name], saved.Saved(name) != nil))
				named, _ := m.nav.Area(name)
				names = append(names, strings.ToLower(named.Name))
			}
		}
		return map[string]any{"areas": out, "message": "Restored the rtylr default of the " + strings.Join(names, " and the ") + "."}, true, nil
	}
}

func findPage(input map[string]any) map[string]any {
	query, _ := input["query"].(string)
	found := catalog.Search(query, 4)
	pages := make([]map[string]any, 0, len(found))
	for _, page := range found {
		card := map[string]string{"path": page.Path, "title": page.Title}
		if len(page.Actions) > 0 {
			card["action"] = page.Actions[0]
		}
		block, _ := json.Marshal(card)
		pages = append(pages, map[string]any{
			"path": page.Path, "app": page.App, "title": page.Title, "summary": page.Summary,
			"tasks": page.Tasks, "buttons": page.Actions, "page_block": string(block),
		})
	}
	out := map[string]any{"pages": pages, "count": len(pages)}
	if len(pages) == 0 {
		out["message"] = "No Portal page matches those words. Try other words, or say the page wasn't found."
	}
	return out
}

func (m *Module) areaOutput(name string, arrangement layout.Arrangement, customized bool) map[string]any {
	area, _ := m.nav.Area(name)
	items := []map[string]any{}
	add := func(ids []string, group string) {
		for i, id := range ids {
			item := map[string]any{"id": id, "label": area.Label(id), "position": i + 1, "hidden": slices.Contains(arrangement.Hidden, id)}
			if area.HasBottom() {
				item["group"] = group
			}
			items = append(items, item)
		}
	}
	add(arrangement.Top, "top")
	add(arrangement.Bottom, "bottom")
	return map[string]any{"area": name, "name": area.Name, "description": area.Description, "items": items, "customized": customized}
}

// areaName reads the area. People call the whole left side "the sidebar",
// so a bare "sidebar" names no single area: a reset covers both, and a move
// follows the items it names.
func areaName(input map[string]any) string {
	value, _ := input["area"].(string)
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "rail", "app bar", "app_bar", "appbar", "apps", "icons", "app icons", "left bar":
		return layout.Rail
	case "workspace", "workspace sidebar", "home sidebar", "pages", "page list":
		return layout.Workspace
	case "", "all", "everything", "both", "sidebar", "side bar", "navigation", "nav", "menu", "left side":
		return "all"
	}
	return strings.ToLower(strings.TrimSpace(value))
}

// inferArea picks the area whose items the change names, preferring the app
// bar.
func (m *Module) inferArea(change layout.Change) string {
	names := slices.Concat(change.Hide, change.Show)
	for _, move := range change.Moves {
		names = append(names, move.ID)
	}
	for _, name := range names {
		if _, ok := m.nav.Rail.Lookup(name); ok {
			return layout.Rail
		}
	}
	return layout.Workspace
}

func changeFromInput(input map[string]any) (layout.Change, error) {
	change := layout.Change{Hide: stringList(input["hide"]), Show: stringList(input["show"])}
	moveFrom := func(raw map[string]any) layout.Move {
		position := firstString(raw, "position", "to")
		if slot, ok := raw["position"].(float64); ok {
			position = strconv.Itoa(int(slot))
		}
		return layout.Move{ID: firstString(raw, "id", "item", "move"), Group: firstString(raw, "group"), Position: position}
	}
	if moves, ok := input["moves"].([]any); ok {
		for _, raw := range moves {
			if object, ok := raw.(map[string]any); ok {
				change.Moves = append(change.Moves, moveFrom(object))
			}
		}
	}
	if move := moveFrom(input); move.ID != "" {
		change.Moves = append(change.Moves, move)
	}
	for _, move := range change.Moves {
		if move.ID == "" {
			return change, errors.New(`each move needs the item id, e.g. {"id":"caret","group":"top","position":"first"}`)
		}
	}
	return change, nil
}

func firstString(input map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := input[key].(string); ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func stringList(value any) []string {
	switch value := value.(type) {
	case string:
		if strings.TrimSpace(value) != "" {
			return []string{value}
		}
	case []any:
		out := []string{}
		for _, item := range value {
			if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}
