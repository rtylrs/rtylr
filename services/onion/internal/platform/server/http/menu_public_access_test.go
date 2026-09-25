package http

import (
	"net/http"
	"testing"
)

func TestCustomerFacingMenuRoutesAreAnonymous(t *testing.T) {
	policy := NewRouteAccessPolicy()
	policy.RegisterAnonymousRoutes(
		"/pos/menu-item/v1/menu-items/public",
		"/hub/business/v1/businesses/by-slug",
		"/hub/business/v1/businesses/public",
		"/hub/business/v1/businesses/name-exists",
	)

	for _, route := range []string{
		"/pos/menu-item/v1/menu-items/public",
		"/hub/business/v1/businesses/by-slug",
		"/hub/business/v1/businesses/public",
		"/hub/business/v1/businesses/name-exists",
	} {
		if !policy.IsAllowedAnonymously(route) {
			t.Errorf("route %q is not anonymous", route)
		}
	}
	if policy.IsAllowedAnonymously("/hub/menu/v1/menus") {
		t.Fatal("internal menu management route must not be anonymous")
	}
}

func TestCustomerFacingMenuRoutesDoNotUseEmployeeFeaturePermissions(t *testing.T) {
	for _, route := range []string{
		"/pos/menu-item/v1/menu-items/public",
		"/hub/business/v1/businesses/public",
	} {
		if appID, permission, ok := featurePermissionForRequest(http.MethodGet, route); ok {
			t.Fatalf("public route %q mapped to employee permission %s:%s", route, appID, permission)
		}
	}

	appID, permission, ok := featurePermissionForRequest(http.MethodGet, "/hub/menu/v1/menus")
	if !ok || appID != "menu" || permission != "menu:menus:view" {
		t.Fatalf("internal menu route mapping = %s:%s,%t", appID, permission, ok)
	}
}
