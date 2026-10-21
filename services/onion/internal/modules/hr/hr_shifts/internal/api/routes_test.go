package api

import (
	"context"
	"testing"

	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func TestRegisterRoutesIncludesContractAvailabilityPaths(t *testing.T) {
	router := serverHttp.NewLegacyRoutes()
	RegisterRoutes(context.Background(), router, nil)

	registered := map[string]bool{}
	for _, route := range router.Routes() {
		registered[route] = true
	}
	for _, route := range []string{
		"GET /hr/shift/v1/availability",
		"POST /hr/shift/v1/availability",
		"DELETE /hr/shift/v1/availability/{id}",
		"GET /hr/shifts/availability",
		"POST /hr/shifts/availability",
		"DELETE /hr/shifts/availability",
	} {
		if !registered[route] {
			t.Errorf("missing registered route %s", route)
		}
	}
}
