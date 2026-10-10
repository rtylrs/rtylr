package main

import (
	"fmt"
	stdhttp "net/http"

	"github.com/voxire/onion/internal/compat/crmcompat"
	"github.com/voxire/onion/internal/platform/server/http"
	crmcompatv1 "github.com/voxire/voxie/gen/go/crm/compat/v1"
)

// compatRoutes lists every legacy route a compat contract serves. Routes not
// listed are still mounted straight from the table.
func compatRoutes() http.RouteSet {
	return http.NewRouteSet(
		crmcompat.CompatibilityServiceRoutes(),
	)
}

// registerCompatContracts serves the compat contracts from the legacy route
// table, with the legacy error shape and request body limit.
func registerCompatContracts(mux *stdhttp.ServeMux, routes *http.LegacyRoutes) {
	if err := crmcompatv1.RegisterCompatibilityServiceServer(
		crmcompat.CompatibilityService{Routes: routes},
		crmcompatv1.WithMux(mux),
		crmcompatv1.WithMaxRequestBodyBytes(http.DefaultMaxRequestBodyBytes),
		crmcompatv1.WithErrorWriter(func(w stdhttp.ResponseWriter, r *stdhttp.Request, e *crmcompatv1.ServerError) {
			status, _, message, _, violations, cause := e.Parts()
			http.WriteContractError(w, r, status, message, violations, cause)
		}),
	); err != nil {
		panic(fmt.Errorf("register crm compat service: %w", err))
	}
}
