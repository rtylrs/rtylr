package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/billing/internal/application"
	"github.com/voxire/onion/internal/modules/billing/internal/routes/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	serverHttp.RegisterGET[handlers.SummaryRequest, handlers.SummaryResponse](
		ctx, router, handlers.SummaryRoutes(),
		&handlers.SummaryHandler{Service: service},
	)

	serverHttp.RegisterGET[handlers.InvoicesRequest, handlers.InvoicesResponse](
		ctx, router, handlers.InvoicesRoutes(),
		&handlers.InvoicesHandler{Service: service},
	)
}
