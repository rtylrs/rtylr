package api

import (
	"context"
	"errors"
	"net/http"
	"testing"

	onekithttp "github.com/voxire/onion/internal/platform/onekiterrors"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	erpsupplierv1 "github.com/voxire/voxie/gen/go/erp/supplier/v1"
)

func TestSupplierAPIRequiresExactPermissionBeforeServiceAccess(t *testing.T) {
	user := &authPkg.AuthUser{
		UserID:     7,
		BusinessID: 9,
		EffectiveAccess: &authorization.EffectiveAccess{Apps: []authorization.EffectiveAppAccess{{
			AppID:       authorization.AppERP,
			Licensed:    true,
			HasAccess:   true,
			Permissions: []string{"erp:suppliers:view"},
		}}},
	}
	ctx := authPkg.WithAuthUser(context.Background(), user)
	api := NewSupplierAPI(nil)

	cases := []struct {
		name string
		call func() error
	}{
		{
			name: "list with no permission",
			call: func() error {
				_, err := api.ListSuppliers(ctx, &erpsupplierv1.ListSuppliersRequest{})
				return err
			},
		},
		{
			name: "create with view only",
			call: func() error {
				_, err := api.CreateSupplier(ctx, &erpsupplierv1.CreateSupplierRequest{Name: "probe"})
				return err
			},
		},
		{
			name: "update with view only",
			call: func() error {
				_, err := api.UpdateSupplier(ctx, &erpsupplierv1.UpdateSupplierRequest{Id: 1, Body: &erpsupplierv1.UpdateSupplierBody{Name: "probe"}})
				return err
			},
		},
		{
			name: "delete with view only",
			call: func() error {
				_, err := api.DeleteSupplier(ctx, &erpsupplierv1.DeleteSupplierRequest{Id: 1})
				return err
			},
		},
	}

	user.EffectiveAccess.Apps[0].Permissions = nil
	for _, tc := range cases[:1] {
		t.Run(tc.name, func(t *testing.T) { assertSupplierForbidden(t, tc.call()) })
	}
	user.EffectiveAccess.Apps[0].Permissions = []string{"erp:suppliers:view"}
	for _, tc := range cases[1:] {
		t.Run(tc.name, func(t *testing.T) { assertSupplierForbidden(t, tc.call()) })
	}
}

func assertSupplierForbidden(t *testing.T, err error) {
	t.Helper()
	var httpErr *onekithttp.Error
	if !errors.As(err, &httpErr) || int(httpErr.GetStatusCode()) != http.StatusForbidden {
		t.Fatalf("error = %v, want forbidden", err)
	}
}
