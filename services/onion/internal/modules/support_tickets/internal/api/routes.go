package api

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/support_tickets/internal/application"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	http.RegisterGET[ListTicketsRequest, ListTicketsResponse](
		ctx, router, ListTicketsRoutes(),
		&ListTicketsHandler{Service: service},
	)

	http.RegisterGETWithValidation[GetTicketRequest, GetTicketResponse](
		ctx, router, GetTicketRoutes(),
		&GetTicketHandler{Service: service},
	)

	http.RegisterPOSTWithValidation[CreateTicketRequest, CreateTicketResponse](
		ctx, router, CreateTicketRoutes(),
		&CreateTicketHandler{Service: service},
	)

	http.RegisterPUTWithValidation[UpdateTicketRequest, UpdateTicketResponse](
		ctx, router, UpdateTicketRoutes(),
		&UpdateTicketHandler{Service: service},
	)
}
