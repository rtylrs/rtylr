package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/support_tickets/internal/application"
	"github.com/voxire/onion/internal/modules/support_tickets/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	http.RegisterGET[handlers.ListTicketsRequest, handlers.ListTicketsResponse](
		ctx, router, handlers.ListTicketsRoutes(),
		&handlers.ListTicketsHandler{Service: service},
	)

	http.RegisterGETWithValidation[handlers.GetTicketRequest, handlers.GetTicketResponse](
		ctx, router, handlers.GetTicketRoutes(),
		&handlers.GetTicketHandler{Service: service},
	)

	http.RegisterPOSTWithValidation[handlers.CreateTicketRequest, handlers.CreateTicketResponse](
		ctx, router, handlers.CreateTicketRoutes(),
		&handlers.CreateTicketHandler{Service: service},
	)

	http.RegisterPUTWithValidation[handlers.UpdateTicketRequest, handlers.UpdateTicketResponse](
		ctx, router, handlers.UpdateTicketRoutes(),
		&handlers.UpdateTicketHandler{Service: service},
	)
}
