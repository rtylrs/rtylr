package config

import (
	"slices"
	"testing"
)

func TestAnonymousRoutesExcludeAuthenticatedConfigRead(t *testing.T) {
	routes := (&Module{}).GetAnonymousRoutes()

	if slices.Contains(routes, "/hub/config/v1/config") {
		t.Fatal("config read route must require auth so bearer tokens are decoded")
	}
	if !slices.Contains(routes, "/hub/config/v1/industries") {
		t.Fatal("industries catalog should remain anonymous")
	}
}
