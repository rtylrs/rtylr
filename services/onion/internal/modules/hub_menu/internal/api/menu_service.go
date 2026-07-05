package api

import (
	"context"
	"net/http"

	hubmenuv1 "github.com/voxire/voxie/gen/go/hub/menu/v1"

	"github.com/voxire/onion/internal/platform/onekiterrors"
)

type MenuAPI struct{}

func NewMenuAPI() *MenuAPI {
	return &MenuAPI{}
}

func (a *MenuAPI) ListMenu(_ context.Context, _ *hubmenuv1.ListMenuRequest) (*hubmenuv1.ListMenuResponse, error) {
	return &hubmenuv1.ListMenuResponse{Items: nil}, nil
}

func (a *MenuAPI) UpdateMenu(_ context.Context, _ *hubmenuv1.UpdateMenuRequest) (*hubmenuv1.UpdateMenuResponse, error) {
	return nil, onekiterrors.New(http.StatusNotImplemented, "Hub menu storage is not implemented")
}
