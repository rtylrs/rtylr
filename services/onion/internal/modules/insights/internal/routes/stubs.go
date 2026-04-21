package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/insights/internal/application"
)

// These register-* functions are no-ops in the initial scaffold. Each one
// is fleshed out by its own commit that adds the matching handler file.
// Keeping them pre-declared means the route-registration entry point in
// routes.go compiles cleanly at every step of the rollout, so the module
// can be toggled off by reverting a single commit without breaking main.

func registerOperationsRoutes(_ context.Context, _ *chi.Mux, _ *application.Service)  {}
func registerReportsRoutes(_ context.Context, _ *chi.Mux, _ *application.Service)     {}
func registerSettingsRoutes(_ context.Context, _ *chi.Mux, _ *application.Service)    {}
