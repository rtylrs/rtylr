package api

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/notifications/internal/application"
)

func RegisterRoutes(_ context.Context, router *chi.Mux, _ *application.Service) {
	router.Get("/dash/ws", ServeDashWS)
}
