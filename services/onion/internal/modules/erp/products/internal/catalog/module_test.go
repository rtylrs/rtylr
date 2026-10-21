package catalog

import "testing"

func TestPublicMenuRouteIsAnonymous(t *testing.T) {
	routes := (&Module{}).GetAnonymousRoutes()
	if len(routes) != 1 || routes[0] != "/pos/menu-item/v1/menu-items/public" {
		t.Fatalf("anonymous menu routes = %v", routes)
	}
}
