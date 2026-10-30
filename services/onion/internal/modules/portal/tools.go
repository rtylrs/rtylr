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
		change, err := changeFromInput(input)
		if err != nil {
			return nil, true, err
		}
		if len(change.Moves) == 0 && len(change.Hide) == 0 && len(change.Show) == 0 {
			// Answering "Done" here would claim a change that never happened.
			return nil, true, errors.New("nothing to change: give moves, hide or show")
		}
		edits := map[string]layout.Change{}
		if area := areaName(input); area == "all" {
			// "The sidebar" spans both areas: each item goes to the area it
			// is in, and one request can change both.
			if edits, err = m.splitByArea(change); err != nil {
				return nil, true, err
			}
		} else {
			edits[area] = change
		}
		saved, err := m.changeAreas(ctx, userID, edits)
		if err != nil {
			return nil, true, err
		}
		changed, orders := []map[string]any{}, []string{}
		for _, name := range areas {
			if _, ok := edits[name]; ok {
				out := m.areaOutput(name, m.arrangements(saved)[name], saved.Saved(name) != nil)
				changed = append(changed, out)
				orders = append(orders, out["order"].(string))
			}
		}
		message := "Done; the Portal shows it now. " + strings.Join(orders, " ")
		if len(changed) == 1 {
			changed[0]["message"] = message
			return changed[0], true, nil
		}
		return map[string]any{"areas": changed, "message": message}, true, nil
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
	return map[string]any{"area": name, "name": area.Name, "description": area.Description, "order": describeOrder(area, arrangement), "items": items, "customized": customized}
}

// describeOrder spells the area out as the person sees it, so Caret reports
// what is on screen rather than working it out from positions.
func describeOrder(area layout.Area, arrangement layout.Arrangement) string {
	labels := func(ids []string) string {
		out := []string{}
		for _, id := range ids {
			if !slices.Contains(arrangement.Hidden, id) {
				out = append(out, area.Label(id))
			}
		}
		return strings.Join(out, ", ")
	}
	text := area.Name + " from the top: " + labels(arrangement.Top)
	if bottom := labels(arrangement.Bottom); bottom != "" {
		text += "; pinned at the bottom: " + bottom
	}
	if len(arrangement.Hidden) > 0 {
		hidden := []string{}
		for _, id := range arrangement.Hidden {
			hidden = append(hidden, area.Label(id))
		}
		text += "; hidden: " + strings.Join(hidden, ", ")
	}
	return text + "."
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

// splitByArea gives each part of a change to the area its item is in. An
// item named in both areas ("Home") goes where its neighbour is, otherwise
// to the app bar.
func (m *Module) splitByArea(change layout.Change) (map[string]layout.Change, error) {
	out := map[string]layout.Change{}
	for _, move := range change.Moves {
		neighbour := ""
		if kind, name, ok := strings.Cut(move.Position, ":"); ok && (kind == "before" || kind == "after") {
			neighbour = name
		}
		area, err := m.areaOf(move.ID, neighbour)
		if err != nil {
			return nil, err
		}
		part := out[area]
		part.Moves = append(part.Moves, move)
		out[area] = part
	}
	for _, list := range []struct {
		names []string
		add   func(*layout.Change, string)
	}{
		{change.Hide, func(c *layout.Change, id string) { c.Hide = append(c.Hide, id) }},
		{change.Show, func(c *layout.Change, id string) { c.Show = append(c.Show, id) }},
	} {
		for _, name := range list.names {
			area, err := m.areaOf(name, "")
			if err != nil {
				return nil, err
			}
			part := out[area]
			list.add(&part, name)
			out[area] = part
		}
	}
	return out, nil
}

func (m *Module) areaOf(item, neighbour string) (string, error) {
	in := func(name string) []string {
		found := []string{}
		for _, area := range areas {
			if a, _ := m.nav.Area(area); name != "" {
				if _, ok := a.Lookup(name); ok {
					found = append(found, area)
				}
			}
		}
		return found
	}
	found := in(item)
	if near := in(neighbour); len(near) > 0 {
		found = slices.DeleteFunc(found, func(area string) bool { return !slices.Contains(near, area) })
	}
	if len(found) == 0 {
		return "", fmt.Errorf("neither the app bar nor the workspace sidebar has %q where you asked; check portal.layout", item)
	}
	return found[0], nil
}

func changeFromInput(input map[string]any) (layout.Change, error) {
	change := layout.Change{Hide: stringList(input["hide"]), Show: stringList(input["show"])}
	moveFrom := func(raw map[string]any) layout.Move {
		position := firstString(raw, "position", "to")
		if slot, ok := raw["position"].(float64); ok {
			// 2.5 stays 2.5 so it is refused, as the string would be.
			position = strconv.FormatFloat(slot, 'f', -1, 64)
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
