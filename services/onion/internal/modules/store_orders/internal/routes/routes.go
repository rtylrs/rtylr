package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/store_orders/internal/application"
	"github.com/voxire/onion/internal/modules/store_orders/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	http.RegisterGET[handlers.ListRequest, handlers.ListResponse](ctx, router, handlers.ListRoutes(), &handlers.ListHandler{Service: service})
	http.RegisterGETWithValidation[handlers.DetailRequest, handlers.DetailResponse](ctx, router, handlers.DetailRoutes(), &handlers.DetailHandler{Service: service})
	http.RegisterPOSTWithValidation[handlers.CreateRequest, handlers.CreateResponse](ctx, router, handlers.CreateRoutes(), &handlers.CreateHandler{Service: service})
	http.RegisterPUTWithValidation[handlers.UpdateRequest, handlers.UpdateResponse](ctx, router, handlers.UpdateRoutes(), &handlers.UpdateHandler{Service: service})
}
