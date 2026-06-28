package config

import (
	"slices"
	"testing"

	"github.com/voxire/onion/internal/modules/config/internal/routes/handlers"
)

func TestAnonymousRoutesExcludeAuthenticatedConfigRead(t *testing.T) {
	routes := (&Module{}).GetAnonymousRoutes()

	for _, route := range handlers.ConfigByOwnerRoutes() {
		if slices.Contains(routes, route) {
			t.Fatalf("config read route %q must require auth so bearer tokens are decoded", route)
		}
	}
}
