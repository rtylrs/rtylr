package api

import (
	"context"
	"errors"
	"net/http"
	"testing"

	onekithttp "github.com/voxire/onion/internal/platform/onekiterrors"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
	erpvariantv1 "github.com/voxire/voxie/gen/go/erp/variant/v1"
)

func TestVariantAPIRequiresExactPermissionBeforeServiceAccess(t *testing.T) {
	user := &authPkg.AuthUser{
		UserID: 7, BusinessID: 9,
		EffectiveAccess: &authorization.EffectiveAccess{Apps: []authorization.EffectiveAppAccess{{
			AppID: authorization.AppERP, Licensed: true, HasAccess: true,
			Permissions: []string{"erp:products:view"},
		}}},
	}
	ctx := authPkg.WithAuthUser(context.Background(), user)
	api := NewVariantAPI(nil)

	cases := []struct {
		name string
		call func() error
	}{
		{name: "list without permission", call: func() error {
			user.EffectiveAccess.Apps[0].Permissions = nil
			_, err := api.ListVariants(ctx, &erpvariantv1.ListVariantsRequest{})
			return err
		}},
		{name: "create with view only", call: func() error {
			user.EffectiveAccess.Apps[0].Permissions = []string{"erp:products:view"}
			_, err := api.CreateVariant(ctx, &erpvariantv1.CreateVariantRequest{ProductId: testutil.Int64Ptr(5201), Name: "probe"})
			return err
		}},
		{name: "update with view only", call: func() error {
			user.EffectiveAccess.Apps[0].Permissions = []string{"erp:products:view"}
			_, err := api.UpdateVariant(ctx, &erpvariantv1.UpdateVariantRequest{Id: 1, Body: &erpvariantv1.UpdateVariantBody{Name: "probe"}})
			return err
		}},
		{name: "delete with view only", call: func() error {
			user.EffectiveAccess.Apps[0].Permissions = []string{"erp:products:view"}
			_, err := api.DeleteVariant(ctx, &erpvariantv1.DeleteVariantRequest{Id: 1})
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
