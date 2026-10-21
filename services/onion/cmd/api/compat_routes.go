package main

import (
	"fmt"
	stdhttp "net/http"

	"github.com/voxire/onion/internal/compat/crmcompat"
	"github.com/voxire/onion/internal/compat/financecompat"
	"github.com/voxire/onion/internal/compat/hrcompat"
	"github.com/voxire/onion/internal/compat/hrshiftcompat"
	"github.com/voxire/onion/internal/compat/organizationcompat"
	"github.com/voxire/onion/internal/compat/poscompat"
	"github.com/voxire/onion/internal/compat/recruitcareerscompat"
	"github.com/voxire/onion/internal/compat/recruitcompat"
	"github.com/voxire/onion/internal/compat/storecompat"
	"github.com/voxire/onion/internal/platform/server/http"
)

// compatPackages are the generated compat contract adapters, each serving
// its module group's legacy routes from the route table.
var compatPackages = []struct {
	name       string
	routes     func() []string
	httpRoutes func() []string
	register   func(*stdhttp.ServeMux, *http.LegacyRoutes) error
}{
	{"crm", crmcompat.Routes, crmcompat.HTTPRoutes, crmcompat.Register},
	{"finance", financecompat.Routes, financecompat.HTTPRoutes, financecompat.Register},
	{"hr", hrcompat.Routes, hrcompat.HTTPRoutes, hrcompat.Register},
	{"hr shift", hrshiftcompat.Routes, hrshiftcompat.HTTPRoutes, hrshiftcompat.Register},
	{"organization", organizationcompat.Routes, organizationcompat.HTTPRoutes, organizationcompat.Register},
	{"pos", poscompat.Routes, poscompat.HTTPRoutes, poscompat.Register},
	{"recruit", recruitcompat.Routes, recruitcompat.HTTPRoutes, recruitcompat.Register},
	{"recruit careers", recruitcareerscompat.Routes, recruitcareerscompat.HTTPRoutes, recruitcareerscompat.Register},
	{"store", storecompat.Routes, storecompat.HTTPRoutes, storecompat.Register},
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

// compatHTTPRoutes lists the compat routes served by @http methods, which run
// the original handler with its route middlewares.
func compatHTTPRoutes() http.RouteSet {
	groups := make([][]string, 0, len(compatPackages))
	for _, pkg := range compatPackages {
		groups = append(groups, pkg.httpRoutes())
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
