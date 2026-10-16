package api

import (
	"context"
	"errors"
	"net/http"
	"testing"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	onekithttp "github.com/voxire/onion/internal/platform/onekiterrors"
	erpbrandv1 "github.com/voxire/voxie/gen/go/erp/brand/v1"
)

func TestBrandAPIRequiresExactPermissionBeforeServiceAccess(t *testing.T) {
	user := &authPkg.AuthUser{
		UserID: 7, BusinessID: 9,
		EffectiveAccess: &authorization.EffectiveAccess{Apps: []authorization.EffectiveAppAccess{{
			AppID: authorization.AppERP, Licensed: true, HasAccess: true,
			Permissions: []string{"erp:products:view"},
		}}},
	}
	ctx := authPkg.WithAuthUser(context.Background(), user)
	api := NewBrandAPI(nil)

	cases := []struct {
		name string
		call func() error
	}{
		{name: "list without permission", call: func() error {
			user.EffectiveAccess.Apps[0].Permissions = nil
			_, err := api.ListBrands(ctx, &erpbrandv1.ListBrandsRequest{})
			return err
		}},
		{name: "create with view only", call: func() error {
			user.EffectiveAccess.Apps[0].Permissions = []string{"erp:products:view"}
			_, err := api.CreateBrand(ctx, &erpbrandv1.CreateBrandRequest{Name: "probe"})
			return err
		}},
		{name: "update with view only", call: func() error {
			user.EffectiveAccess.Apps[0].Permissions = []string{"erp:products:view"}
			_, err := api.UpdateBrand(ctx, &erpbrandv1.UpdateBrandRequest{Id: 1, Body: &erpbrandv1.UpdateBrandBody{Name: "probe"}})
			return err
		}},
		{name: "delete with view only", call: func() error {
			user.EffectiveAccess.Apps[0].Permissions = []string{"erp:products:view"}
			_, err := api.DeleteBrand(ctx, &erpbrandv1.DeleteBrandRequest{Id: 1})
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
