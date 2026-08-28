package api

import (
	"context"
	"errors"
	"net/http"
	"testing"

	onkithttp "github.com/corezio/onekit/http"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	erpproductv1 "github.com/voxire/voxie/gen/go/erp/product/v1"
)

func productAuthorizationUser(permissions ...string) *authPkg.AuthUser {
	return &authPkg.AuthUser{
		BusinessID: 9,
		UserID:     7,
		EffectiveAccess: &authorization.EffectiveAccess{
			Apps: []authorization.EffectiveAppAccess{{
				AppID:       authorization.AppERP,
				Licensed:    true,
				HasAccess:   true,
				Permissions: permissions,
			}},
		},
	}
}

func productAuthorizationStatus(err error) int {
	var oneKitErr *onkithttp.Error
	if errors.As(err, &oneKitErr) {
		return int(oneKitErr.GetStatusCode())
	}
	return 0
}

func TestProductAPIRequiresExactPermissionBeforeServiceAccess(t *testing.T) {
	ctx := authPkg.WithAuthUser(context.Background(), productAuthorizationUser())
	api := NewProductAPI(nil)

	cases := []struct {
		name string
		call func() error
	}{
		{
			name: "list",
			call: func() error {
				_, err := api.ListProducts(ctx, &erpproductv1.ListProductsRequest{})
				return err
			},
		},
		{
			name: "create",
			call: func() error {
				_, err := api.CreateProduct(ctx, &erpproductv1.CreateProductRequest{})
				return err
			},
		},
		{
			name: "update",
			call: func() error {
				_, err := api.UpdateProduct(ctx, &erpproductv1.UpdateProductRequest{})
				return err
			},
		},
		{
			name: "delete",
			call: func() error {
				_, err := api.DeleteProduct(ctx, &erpproductv1.DeleteProductRequest{})
				return err
			},
		},
		{
			name: "get",
			call: func() error {
				_, err := api.GetProduct(ctx, &erpproductv1.GetProductRequest{})
				return err
			},
		},
		{
			name: "barcode",
			call: func() error {
				_, err := api.GetProductByBarcode(ctx, &erpproductv1.GetProductByBarcodeRequest{})
				return err
			},
		},
		{
			name: "full",
			call: func() error {
				_, err := api.GetProductFull(ctx, &erpproductv1.GetProductFullRequest{})
				return err
			},
		},
		{
			name: "attributes",
			call: func() error {
				_, err := api.ListProductAttributes(ctx, &erpproductv1.ListProductAttributesRequest{})
				return err
			},
		},
		{
			name: "dashboard",
			call: func() error {
				_, err := api.GetDashboard(ctx, &erpproductv1.GetDashboardRequest{})
				return err
			},
		},
		{
			name: "reorder suggestions",
			call: func() error {
				_, err := api.ListReorderSuggestions(ctx, &erpproductv1.ListReorderSuggestionsRequest{})
				return err
			},
		},
		{
			name: "expiring items",
			call: func() error {
				_, err := api.ListExpiringItems(ctx, &erpproductv1.ListExpiringItemsRequest{})
				return err
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := productAuthorizationStatus(tc.call()); got != http.StatusForbidden {
				t.Fatalf("status = %d, want %d", got, http.StatusForbidden)
			}
		})
	}
}

func TestProductPermissionRejectsWrongAppAndUnlicensedUsers(t *testing.T) {
	wrongApp := productAuthorizationUser("erp:products:view")
	wrongApp.EffectiveAccess.Apps[0].AppID = authorization.AppFinance
	api := NewProductAPI(nil)
	if _, err := api.ListProducts(authPkg.WithAuthUser(context.Background(), wrongApp), &erpproductv1.ListProductsRequest{}); productAuthorizationStatus(err) != http.StatusForbidden {
		t.Fatalf("wrong app status = %d, want %d", productAuthorizationStatus(err), http.StatusForbidden)
	}

	unlicensed := productAuthorizationUser("erp:products:view")
	unlicensed.EffectiveAccess.Apps[0].Licensed = false
	if _, err := api.ListProducts(authPkg.WithAuthUser(context.Background(), unlicensed), &erpproductv1.ListProductsRequest{}); productAuthorizationStatus(err) != http.StatusForbidden {
		t.Fatalf("unlicensed status = %d, want %d", productAuthorizationStatus(err), http.StatusForbidden)
	}
}
