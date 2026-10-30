// Package portal owns the shape of the Portal for each person: the map of its
// pages and the arrangement of its navigation. The Portal, its settings and
// Caret all read and change it through this module, under the same rules.
package portal

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/voxire/onion/internal/modules/portal/internal/catalog"
	"github.com/voxire/onion/internal/modules/portal/internal/infrastructure"
	"github.com/voxire/onion/internal/modules/portal/internal/layout"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	portalv1 "github.com/voxire/voxie/gen/go/hub/portal/v1"
)

type layoutStore interface {
	Layout(ctx context.Context, userID int64) (layout.Layout, error)
	// UpdateLayout applies edit to the saved layout as one locked change.
	UpdateLayout(ctx context.Context, userID int64, edit func(layout.Layout) (layout.Layout, error)) (layout.Layout, error)
}

type Module struct {
	store layoutStore
	nav   layout.Navigation
}

func NewModule(db *sql.DB) *Module {
	return &Module{store: infrastructure.NewRepository(db), nav: catalog.Navigation()}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := portalv1.RegisterPortalLayoutServiceServer(&layoutAPI{m}, portalv1.WithMux(mux)); err != nil {
		panic(fmt.Errorf("register portal layout service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string { return nil }

var areas = []string{layout.Rail, layout.Workspace}

// signedIn is the person whose Portal this is.
func signedIn(ctx context.Context) (int64, error) {
	user, ok := authPkg.GetAuthUser(ctx)
	if !ok || user == nil || user.UserID <= 0 {
		return 0, onekiterrors.PermissionDenied()
	}
	return user.UserID, nil
}

// arrangements resolves every area of a saved layout.
func (m *Module) arrangements(saved layout.Layout) map[string]layout.Arrangement {
	out := map[string]layout.Arrangement{}
	for _, name := range areas {
		area, _ := m.nav.Area(name)
		out[name] = area.Resolve(saved.Saved(name))
	}
	return out
}

// change rewrites one area of the person's layout and returns the result.
func (m *Module) change(ctx context.Context, userID int64, name string, edit func(layout.Area, layout.Arrangement) (layout.Arrangement, error)) (layout.Layout, layout.Arrangement, error) {
	area, ok := m.nav.Area(name)
	if !ok {
		return layout.Layout{}, layout.Arrangement{}, fmt.Errorf("unknown area %q; use rail or workspace", name)
	}
	var next layout.Arrangement
	saved, err := m.store.UpdateLayout(ctx, userID, func(saved layout.Layout) (layout.Layout, error) {
		arrangement, err := edit(area, area.Resolve(saved.Saved(name)))
		if err != nil {
			return saved, err
		}
		next = arrangement
		return saved.With(name, area.Stored(arrangement)), nil
	})
	return saved, next, err
}

// changeAreas applies a change to each named area as one locked update.
func (m *Module) changeAreas(ctx context.Context, userID int64, edits map[string]layout.Change) (layout.Layout, error) {
	for name := range edits {
		if _, ok := m.nav.Area(name); !ok {
			return layout.Layout{}, fmt.Errorf("unknown area %q; use rail or workspace", name)
		}
	}
	return m.store.UpdateLayout(ctx, userID, func(saved layout.Layout) (layout.Layout, error) {
		changed, before := false, []string{}
		for _, name := range areas {
			change, ok := edits[name]
			if !ok {
				continue
			}
			area, _ := m.nav.Area(name)
			current := area.Resolve(saved.Saved(name))
			next, err := area.Apply(current, change)
			if err != nil {
				return saved, err
			}
			changed = changed || !sameArrangement(current, next)
			before = append(before, describeOrder(area, current))
			saved = saved.With(name, area.Stored(next))
		}
		if !changed {
			// Answering "Done" here would report a move nobody can see.
			return saved, fmt.Errorf("nothing changed, so nothing moved on screen. It still reads: %s Without a group, first, last and a slot stay in the item's own group; to move an item between the top and the bottom group give group top or bottom, or place it before:<id> or after:<id> an item in that group", strings.Join(before, " "))
		}
		return saved, nil
	})
}

// reset restores one area, or every area when name is empty.
func (m *Module) reset(ctx context.Context, userID int64, name string) (layout.Layout, error) {
	if name == "" {
		return m.store.UpdateLayout(ctx, userID, func(layout.Layout) (layout.Layout, error) {
			return layout.Layout{}, nil
		})
	}
	saved, _, err := m.change(ctx, userID, name, func(area layout.Area, _ layout.Arrangement) (layout.Arrangement, error) {
		return area.Default(), nil
	})
	return saved, err
}

func sameArrangement(a, b layout.Arrangement) bool {
	return slices.Equal(a.Top, b.Top) && slices.Equal(a.Bottom, b.Bottom) && slices.Equal(a.Hidden, b.Hidden)
}
