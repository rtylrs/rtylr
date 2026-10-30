// Package layout arranges the Portal navigation for one person: the order of
// the app bar and the workspace sidebar, and what they hid.
package layout

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

const (
	Rail      = "rail"
	Workspace = "workspace"
)

// Item is one entry of a navigation area.
type Item struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	// Required items can move but can't be hidden.
	Required bool `json:"required,omitempty"`
	// Bottom items sit below the apps by default (the app bar only).
	Bottom bool `json:"bottom,omitempty"`
}

// Area is a navigation area with its items in default order.
type Area struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Items       []Item `json:"items"`
}

// Navigation holds every area a person can arrange.
type Navigation struct {
	Rail      Area `json:"rail"`
	Workspace Area `json:"workspace"`
}

// Arrangement is an area in display order. Hidden items keep their place, so
// showing one again puts it back where it was.
type Arrangement struct {
	Top    []string `json:"top"`
	Bottom []string `json:"bottom,omitempty"`
	Hidden []string `json:"hidden,omitempty"`
}

// Layout is what is saved for a person. A nil area follows the default.
type Layout struct {
	Rail      *Arrangement `json:"rail,omitempty"`
	Workspace *Arrangement `json:"workspace,omitempty"`
}

// Move puts one item somewhere else.
type Move struct {
	ID string
	// Group is "top" or "bottom" of the app bar; empty keeps the item's group,
	// or takes the group of the item named in Position.
	Group string
	// Position is "first", "last", "before:<id>", "after:<id>" or a slot
	// number counted from 1 within the group; empty means first.
	Position string
}

// Change is one edit to an area.
type Change struct {
	Moves []Move
	Hide  []string
	Show  []string
}

// Area returns the named area.
func (n Navigation) Area(name string) (Area, bool) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case Rail:
		return n.Rail, true
	case Workspace:
		return n.Workspace, true
	}
	return Area{}, false
}

// Saved returns the stored arrangement of the named area.
func (l Layout) Saved(area string) *Arrangement {
	if area == Rail {
		return l.Rail
	}
	return l.Workspace
}

// With returns the layout with the named area replaced; nil restores the
// default.
func (l Layout) With(area string, arrangement *Arrangement) Layout {
	if area == Rail {
		l.Rail = arrangement
	} else {
		l.Workspace = arrangement
	}
	return l
}

// Empty reports whether every area follows the default.
func (l Layout) Empty() bool { return l.Rail == nil && l.Workspace == nil }

// HasBottom reports whether the area has a bottom group.
func (a Area) HasBottom() bool {
	return slices.ContainsFunc(a.Items, func(item Item) bool { return item.Bottom })
}

// Lookup finds an item by id or label, ignoring case, so "Audit Log" and
// "auditlog" both name auditLog.
func (a Area) Lookup(name string) (Item, bool) {
	name = strings.TrimSpace(name)
	for _, item := range a.Items {
		if strings.EqualFold(item.ID, name) || strings.EqualFold(item.Label, name) {
			return item, true
		}
	}
	return Item{}, false
}

// Label returns an item's label, or its id when unknown.
func (a Area) Label(id string) string {
	if item, ok := a.Lookup(id); ok {
		return item.Label
	}
	return id
}

// Default is the rtylr arrangement of the area.
func (a Area) Default() Arrangement {
	var out Arrangement
	for _, item := range a.Items {
		if item.Bottom {
			out.Bottom = append(out.Bottom, item.ID)
		} else {
			out.Top = append(out.Top, item.ID)
		}
	}
	return out
}

// Resolve completes a saved arrangement against the area. Unknown and
// repeated ids drop out, required items can't stay hidden, and items it
// doesn't mention (an app added since it was saved) take their default place.
func (a Area) Resolve(saved *Arrangement) Arrangement {
	if saved == nil {
		return a.Default()
	}
	placed := map[string]bool{}
	keep := func(ids []string) []string {
		out := []string{}
		for _, id := range ids {
			if item, ok := a.Lookup(id); ok && !placed[item.ID] {
				placed[item.ID] = true
				out = append(out, item.ID)
			}
		}
		return out
	}
	out := Arrangement{Top: keep(saved.Top)}
	if a.HasBottom() {
		out.Bottom = keep(saved.Bottom)
	}
	def := a.Default()
	for _, item := range a.Items {
		if placed[item.ID] {
			continue
		}
		placed[item.ID] = true
		if item.Bottom {
			out.Bottom = insertAfterPredecessor(out.Bottom, def.Bottom, item.ID)
		} else {
			out.Top = insertAfterPredecessor(out.Top, def.Top, item.ID)
		}
	}
	hidden := map[string]bool{}
	for _, id := range saved.Hidden {
		if item, ok := a.Lookup(id); ok && !item.Required {
			hidden[item.ID] = true
		}
	}
	for _, id := range append(slices.Clone(out.Top), out.Bottom...) {
		if hidden[id] {
			out.Hidden = append(out.Hidden, id)
		}
	}
	return out
}

// insertAfterPredecessor places id after the nearest item that comes before
// it by default and is already in group, or first when there is none.
func insertAfterPredecessor(group, defaults []string, id string) []string {
	at := slices.Index(defaults, id)
	for i := at - 1; i >= 0; i-- {
		if pos := slices.Index(group, defaults[i]); pos >= 0 {
			return slices.Insert(group, pos+1, id)
		}
	}
	return slices.Insert(group, 0, id)
}

// IsDefault reports whether the arrangement is the rtylr default.
func (a Area) IsDefault(arrangement Arrangement) bool {
	def := a.Default()
	return slices.Equal(arrangement.Top, def.Top) && slices.Equal(arrangement.Bottom, def.Bottom) && len(arrangement.Hidden) == 0
}

// Stored is what to save for an arrangement: nil when it is the default.
func (a Area) Stored(arrangement Arrangement) *Arrangement {
	if a.IsDefault(arrangement) {
		return nil
	}
	return &arrangement
}

// Apply makes a change to the current arrangement. Errors name the valid
// choices so the caller can correct itself.
func (a Area) Apply(current Arrangement, change Change) (Arrangement, error) {
	if len(change.Moves) == 0 && len(change.Hide) == 0 && len(change.Show) == 0 {
		return current, errors.New("nothing to change: give moves, hide or show")
	}
	out := a.Resolve(&current)
	for _, move := range change.Moves {
		next, err := a.move(out, move)
		if err != nil {
			return current, err
		}
		out = next
	}
	hidden := slices.Clone(out.Hidden)
	for _, name := range change.Hide {
		item, err := a.item(name)
		if err != nil {
			return current, err
		}
		if item.Required {
			return current, fmt.Errorf("%s can't be hidden; it can only move", item.Label)
		}
		hidden = append(hidden, item.ID)
	}
	for _, name := range change.Show {
		item, err := a.item(name)
		if err != nil {
			return current, err
		}
		hidden = slices.DeleteFunc(hidden, func(id string) bool { return id == item.ID })
	}
	out.Hidden = hidden
	return a.Resolve(&out), nil
}

func (a Area) item(name string) (Item, error) {
	if item, ok := a.Lookup(name); ok {
		return item, nil
	}
	ids := make([]string, 0, len(a.Items))
	for _, item := range a.Items {
		ids = append(ids, item.ID)
	}
	return Item{}, fmt.Errorf("the %s has no item %q; its items are %s", strings.ToLower(a.Name), name, strings.Join(ids, ", "))
}

func (a Area) move(current Arrangement, move Move) (Arrangement, error) {
	item, err := a.item(move.ID)
	if err != nil {
		return current, err
	}
	position := strings.ToLower(strings.TrimSpace(move.Position))
	group := strings.ToLower(strings.TrimSpace(move.Group))
	var anchor string
	slot := 0
	if kind, name, ok := strings.Cut(position, ":"); ok && (kind == "before" || kind == "after") {
		other, err := a.item(name)
		if err != nil {
			return current, err
		}
		if other.ID == item.ID {
			return current, fmt.Errorf("can't place %s relative to itself", item.Label)
		}
		anchor, position = other.ID, kind
		anchorGroup := "top"
		if slices.Contains(current.Bottom, other.ID) {
			anchorGroup = "bottom"
		}
		if group != "" && group != anchorGroup {
			return current, fmt.Errorf("%s is in the %s group, not %s", other.Label, anchorGroup, group)
		}
		group = anchorGroup
	} else if position == "" {
		position = "first"
	} else if n, err := strconv.Atoi(position); err == nil {
		if n < 1 {
			return current, fmt.Errorf("slot %d doesn't exist; slots count from 1", n)
		}
		slot = n - 1
		position = "slot"
	} else if position != "first" && position != "last" {
		return current, fmt.Errorf("position %q isn't one of first, last, before:<id>, after:<id> or a slot number", move.Position)
	}
	switch group {
	case "":
		group = "top"
		if slices.Contains(current.Bottom, item.ID) {
			group = "bottom"
		}
	case "top":
	case "bottom":
		if !a.HasBottom() {
			return current, fmt.Errorf("the %s has no bottom group", strings.ToLower(a.Name))
		}
	default:
		return current, fmt.Errorf("group %q isn't top or bottom", move.Group)
	}

	isItem := func(id string) bool { return id == item.ID }
	out := Arrangement{
		Top:    slices.DeleteFunc(slices.Clone(current.Top), isItem),
		Bottom: slices.DeleteFunc(slices.Clone(current.Bottom), isItem),
		Hidden: current.Hidden,
	}
	target := &out.Top
	if group == "bottom" {
		target = &out.Bottom
	}
	switch position {
	case "first":
		*target = slices.Insert(*target, 0, item.ID)
	case "last":
		*target = append(*target, item.ID)
	case "before":
		*target = slices.Insert(*target, slices.Index(*target, anchor), item.ID)
	case "after":
		*target = slices.Insert(*target, slices.Index(*target, anchor)+1, item.ID)
	case "slot":
		*target = slices.Insert(*target, min(slot, len(*target)), item.ID)
	}
	return out, nil
}
