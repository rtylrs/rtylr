package api

import (
	"context"
	"testing"

	moduleinfra "github.com/voxire/onion/internal/modules/hub_menu/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/hub_menu/internal/application"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
	hubmenuv1 "github.com/voxire/voxie/gen/go/hub/menu/v1"
	posmenuitemv1 "github.com/voxire/voxie/gen/go/pos/menu_item/v1"
)

func TestDirectAuthorizationMatrix(t *testing.T) {
	testutil.RunDirectAuthorizationCases(t, []testutil.DirectAuthorizationCase{
		{Name: "menu menus view", App: authorization.AppMenu, Permission: "menu:menus:view", Call: func(ctx context.Context) (any, error) {
			response, err := NewMenuAPI(application.NewService(ctx, moduleinfra.NewRepository(testutil.DirectAuthorizationDB(ctx)))).ListMenu(ctx, nil)
			return response, err
		}},
		{Name: "menu menus edit", App: authorization.AppMenu, Permission: "menu:menus:edit", Call: func(ctx context.Context) (any, error) {
			productID := int64(5201)
			response, err := NewMenuAPI(application.NewService(ctx, moduleinfra.NewRepository(testutil.DirectAuthorizationDB(ctx)))).UpdateMenu(ctx, &hubmenuv1.UpdateMenuRequest{Items: []*posmenuitemv1.MenuItem{{ProductId: &productID, Name: "Direct Authorization Menu Item", IsActive: true}}})
			return response, err
		}},
		{Name: "menu branding edit", App: authorization.AppMenu, Permission: "menu:branding:edit", Call: func(ctx context.Context) (any, error) {
			response, err := NewMenuAPI(application.NewService(ctx, moduleinfra.NewRepository(testutil.DirectAuthorizationDB(ctx)))).ListMenuBanners(ctx, nil)
			return response, err
		}},
	})
}
