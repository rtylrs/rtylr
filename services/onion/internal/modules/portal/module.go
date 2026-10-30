// Package portal owns the shape of the Portal for each person: the map of its
// pages and the arrangement of its navigation. The Portal, its settings and
// Caret all read and change it through this module, under the same rules.
package portal

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/portal/internal/catalog"
	"github.com/voxire/onion/internal/modules/portal/internal/infrastructure"
	"github.com/voxire/onion/internal/modules/portal/internal/layout"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	portalv1 "github.com/voxire/voxie/gen/go/hub/portal/v1"
)

type layoutStore interface {
	Layout(ctx context.Context, userID int64) (layout.Layout, error)
	SaveLayout(ctx context.Context, userID int64, l layout.Layout) error
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
	saved, err := m.store.Layout(ctx, userID)
	if err != nil {
		return layout.Layout{}, layout.Arrangement{}, err
	}
	next, err := edit(area, area.Resolve(saved.Saved(name)))
	if err != nil {
		return layout.Layout{}, layout.Arrangement{}, err
	}
	saved = saved.With(name, area.Stored(next))
	if err := m.store.SaveLayout(ctx, userID, saved); err != nil {
		return layout.Layout{}, layout.Arrangement{}, err
	}
	return saved, next, nil
}

// reset restores one area, or every area when name is empty.
func (m *Module) reset(ctx context.Context, userID int64, name string) (layout.Layout, error) {
	if name == "" {
		return layout.Layout{}, m.store.SaveLayout(ctx, userID, layout.Layout{})
	}
	saved, _, err := m.change(ctx, userID, name, func(area layout.Area, _ layout.Arrangement) (layout.Arrangement, error) {
		return area.Default(), nil
	})
	return saved, err
}
