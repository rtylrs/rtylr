package api

import (
	"context"
	"errors"
	"net/http"
	"testing"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	onekithttp "github.com/voxire/onion/internal/platform/onekiterrors"
	erpcategoryv1 "github.com/voxire/voxie/gen/go/erp/category/v1"
)

func TestCategoryAPIRequiresExactPermissionBeforeServiceAccess(t *testing.T) {
	user := &authPkg.AuthUser{
		UserID: 7, BusinessID: 9,
		EffectiveAccess: &authorization.EffectiveAccess{Apps: []authorization.EffectiveAppAccess{{
			AppID: authorization.AppERP, Licensed: true, HasAccess: true,
			Permissions: []string{"erp:products:view"},
		}}},
	}
	ctx := authPkg.WithAuthUser(context.Background(), user)
	api := NewCategoryAPI(nil)

	cases := []struct {
		name string
		call func() error
	}{
		{name: "list without permission", call: func() error {
			user.EffectiveAccess.Apps[0].Permissions = nil
			_, err := api.ListCategories(ctx, &erpcategoryv1.ListCategoriesRequest{})
			return err
		}},
		{name: "create with view only", call: func() error {
			user.EffectiveAccess.Apps[0].Permissions = []string{"erp:products:view"}
			_, err := api.CreateCategory(ctx, &erpcategoryv1.CreateCategoryRequest{Name: "probe"})
			return err
		}},
		{name: "update with view only", call: func() error {
			user.EffectiveAccess.Apps[0].Permissions = []string{"erp:products:view"}
			_, err := api.UpdateCategory(ctx, &erpcategoryv1.UpdateCategoryRequest{Id: 1, Body: &erpcategoryv1.UpdateCategoryBody{Name: "probe"}})
			return err
		}},
		{name: "delete with view only", call: func() error {
			user.EffectiveAccess.Apps[0].Permissions = []string{"erp:products:view"}
			_, err := api.DeleteCategory(ctx, &erpcategoryv1.DeleteCategoryRequest{Id: 1})
			return err
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var httpErr *onekithttp.Error
			if err := tc.call(); !errors.As(err, &httpErr) || int(httpErr.GetStatusCode()) != http.StatusForbidden {
				t.Fatalf("error = %v, want forbidden", err)
			}
		})
	}
}
