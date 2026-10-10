package main

import (
	"fmt"
	stdhttp "net/http"

	"github.com/voxire/onion/internal/compat/crmcompat"
	"github.com/voxire/onion/internal/compat/hrcompat"
	"github.com/voxire/onion/internal/platform/server/http"
)

// compatPackages are the generated compat contract adapters, each serving
// its module group's legacy routes from the route table.
var compatPackages = []struct {
	name     string
	routes   func() []string
	register func(*stdhttp.ServeMux, *http.LegacyRoutes) error
}{
	{"crm", crmcompat.Routes, crmcompat.Register},
	{"hr", hrcompat.Routes, hrcompat.Register},
}

// compatRoutes lists every legacy route a compat contract serves. Routes not
// listed are still mounted straight from the table.
func compatRoutes() http.RouteSet {
	groups := make([][]string, 0, len(compatPackages))
	for _, pkg := range compatPackages {
		groups = append(groups, pkg.routes())
	}
	return http.NewRouteSet(groups...)
}

// registerCompatContracts serves the compat contracts from the legacy route
// table.
func registerCompatContracts(mux *stdhttp.ServeMux, routes *http.LegacyRoutes) {
	for _, pkg := range compatPackages {
		if err := pkg.register(mux, routes); err != nil {
			panic(fmt.Errorf("register %s compat contract: %w", pkg.name, err))
		}
	}
}
