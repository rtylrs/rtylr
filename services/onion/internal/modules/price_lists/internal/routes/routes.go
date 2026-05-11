package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/price_lists/internal/application"
	"github.com/voxire/onion/internal/modules/price_lists/internal/routes/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	serverHttp.RegisterGET[handlers.ListPriceListsRequest, handlers.ListPriceListsResponse](
		ctx,
		router,
		handlers.ListPriceListsRoutes(),
		&handlers.ListPriceListsHandler{Service: service},
	)

	serverHttp.RegisterPOSTWithValidation[handlers.CreatePriceListRequest, handlers.CreatePriceListResponse](
		ctx,
		router,
		handlers.CreatePriceListRoutes(),
		&handlers.CreatePriceListHandler{Service: service},
	)
}
