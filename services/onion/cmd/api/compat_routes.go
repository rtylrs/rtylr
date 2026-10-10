package main

import (
	"fmt"
	stdhttp "net/http"

	"github.com/voxire/onion/internal/compat/consolecompat"
	"github.com/voxire/onion/internal/compat/crmcompat"
	"github.com/voxire/onion/internal/compat/financecompat"
	"github.com/voxire/onion/internal/compat/hrcompat"
	"github.com/voxire/onion/internal/compat/poscompat"
	"github.com/voxire/onion/internal/compat/recruitcareerscompat"
	"github.com/voxire/onion/internal/compat/recruitcompat"
	"github.com/voxire/onion/internal/compat/storecompat"
	"github.com/voxire/onion/internal/platform/server/http"
)

// compatPackages are the generated compat contract adapters, each serving
// its module group's legacy routes from the route table.
var compatPackages = []struct {
	name     string
	routes   func() []string
	register func(*stdhttp.ServeMux, *http.LegacyRoutes) error
}{
	{"console", consolecompat.Routes, consolecompat.Register},
	{"crm", crmcompat.Routes, crmcompat.Register},
	{"finance", financecompat.Routes, financecompat.Register},
	{"hr", hrcompat.Routes, hrcompat.Register},
	{"pos", poscompat.Routes, poscompat.Register},
	{"recruit", recruitcompat.Routes, recruitcompat.Register},
	{"recruit careers", recruitcareerscompat.Routes, recruitcareerscompat.Register},
	{"store", storecompat.Routes, storecompat.Register},
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
