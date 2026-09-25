package api

import (
	"context"
	"net/http"
	"testing"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	posmenuitemv1 "github.com/voxire/voxie/gen/go/pos/menu_item/v1"
)

func posMenuUser(app authorization.AppID, licensed bool, permissions ...string) *authPkg.AuthUser {
	return &authPkg.AuthUser{
		UserID:     1,
		BusinessID: 1,
		EffectiveAccess: &authorization.EffectiveAccess{
			Apps: []authorization.EffectiveAppAccess{{
				AppID:       app,
				Licensed:    licensed,
				HasAccess:   licensed,
				Permissions: permissions,
			}},
		},
	}
}

func posMenuPermissionStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}
	if oneKitErr, ok := err.(interface{ HTTPStatusCode() int }); ok {
		return oneKitErr.HTTPStatusCode()
	}
	return http.StatusInternalServerError
}

func TestMenuItemAPIRequiresCanonicalMenuPermissionBeforeServiceAccess(t *testing.T) {
	tests := []struct {
		name string
		call func(context.Context) error
	}{
		{name: "list", call: func(ctx context.Context) error {
			_, err := (&MenuItemAPI{}).ListMenuItems(ctx, &posmenuitemv1.ListMenuItemsRequest{})
			return err
		}},
		{name: "create", call: func(ctx context.Context) error {
			_, err := (&MenuItemAPI{}).CreateMenuItem(ctx, &posmenuitemv1.CreateMenuItemRequest{})
			return err
		}},
		{name: "update", call: func(ctx context.Context) error {
			_, err := (&MenuItemAPI{}).UpdateMenuItem(ctx, &posmenuitemv1.UpdateMenuItemRequest{})
			return err
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := authPkg.WithAuthUser(context.Background(), posMenuUser(authorization.AppMenu, true))
			if got := posMenuPermissionStatus(tt.call(ctx)); got != http.StatusForbidden {
				t.Fatalf("status = %d, want %d", got, http.StatusForbidden)
			}
		})
	}
}

func TestMenuItemPermissionRequiresLicensedMenuApp(t *testing.T) {
	for name, user := range map[string]*authPkg.AuthUser{
		"wrong app":  posMenuUser(authorization.AppPOS, true, "menu:menus:edit"),
		"unlicensed": posMenuUser(authorization.AppMenu, false, "menu:menus:edit"),
	} {
		t.Run(name, func(t *testing.T) {
			ctx := authPkg.WithAuthUser(context.Background(), user)
			if _, _, err := requireMenuPermission(ctx, "menu:menus:edit"); posMenuPermissionStatus(err) != http.StatusForbidden {
				t.Fatalf("permission status = %d, want %d", posMenuPermissionStatus(err), http.StatusForbidden)
			}
		})
	}
}
