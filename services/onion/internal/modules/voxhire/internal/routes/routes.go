package routes

import (
	"context"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/voxhire/internal/application"
	"github.com/voxire/onion/internal/modules/voxhire/internal/routes/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	// Public company-side hiring inquiry — JSON, validated.
	// Tighter per-route rate limit because this is a public surface
	// that bots will try to spam. 10/min/IP is generous for real use.
	serverHttp.RegisterPOSTWithValidationAndMiddleware[handlers.ContactRequest, handlers.ContactResponse](
		ctx, router, handlers.ContactRoutes(),
		&handlers.ContactHandler{Service: service},
		serverHttp.PerRouteRateLimit(10, time.Minute),
	)

	// Public talent-side application — multipart/form-data with CV
	// upload. The generic JSON wrapHandler can't parse multipart, so
	// this handler is wired directly to chi the same way careers/apply
	// is. We register it once per route path returned by ApplicationsRoutes().
	for _, path := range handlers.ApplicationsRoutes() {
		router.With(serverHttp.PerRouteRateLimit(5, time.Minute)).
			Post(path, handlers.ApplicationsHandler(service))
	}
}
