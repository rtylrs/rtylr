package portal

import (
	"context"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/portal/internal/layout"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	portalv1 "github.com/voxire/voxie/gen/go/hub/portal/v1"
)

type layoutAPI struct{ m *Module }

func (a *layoutAPI) GetPortalLayout(ctx context.Context, _ *portalv1.GetPortalLayoutRequest) (*portalv1.GetPortalLayoutResponse, error) {
	userID, err := signedIn(ctx)
	if err != nil {
		return nil, err
	}
	saved, err := a.m.store.Layout(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &portalv1.GetPortalLayoutResponse{Layout: a.m.layoutProto(saved)}, nil
}

func (a *layoutAPI) SavePortalLayout(ctx context.Context, req *portalv1.SavePortalLayoutRequest) (*portalv1.SavePortalLayoutResponse, error) {
	userID, err := signedIn(ctx)
	if err != nil {
		return nil, err
	}
	saved, err := a.m.store.UpdateLayout(ctx, userID, func(saved layout.Layout) (layout.Layout, error) {
		for name, sent := range map[string]*portalv1.PortalArrangement{layout.Rail: req.Rail, layout.Workspace: req.Workspace} {
			if sent == nil {
				continue
			}
			area, _ := a.m.nav.Area(name)
			saved = saved.With(name, area.Stored(area.Resolve(&layout.Arrangement{Top: sent.Top, Bottom: sent.Bottom, Hidden: sent.Hidden})))
		}
		return saved, nil
	})
	if err != nil {
		return nil, err
	}
	return &portalv1.SavePortalLayoutResponse{Layout: a.m.layoutProto(saved)}, nil
}

func (a *layoutAPI) ResetPortalLayout(ctx context.Context, req *portalv1.ResetPortalLayoutRequest) (*portalv1.ResetPortalLayoutResponse, error) {
	userID, err := signedIn(ctx)
	if err != nil {
		return nil, err
	}
	name := strings.ToLower(strings.TrimSpace(req.Area))
	if _, ok := a.m.nav.Area(name); name != "" && !ok {
		return nil, onekiterrors.New(http.StatusBadRequest, "area must be rail, workspace or empty")
	}
	saved, err := a.m.reset(ctx, userID, name)
	if err != nil {
		return nil, err
	}
	return &portalv1.ResetPortalLayoutResponse{Layout: a.m.layoutProto(saved)}, nil
}

func (m *Module) layoutProto(saved layout.Layout) *portalv1.PortalLayout {
	resolved := m.arrangements(saved)
	proto := func(name string) *portalv1.PortalArrangement {
		arrangement := resolved[name]
		return &portalv1.PortalArrangement{
			Top:        arrangement.Top,
			Bottom:     arrangement.Bottom,
			Hidden:     arrangement.Hidden,
			Customized: saved.Saved(name) != nil,
		}
	}
	return &portalv1.PortalLayout{Rail: proto(layout.Rail), Workspace: proto(layout.Workspace)}
}
