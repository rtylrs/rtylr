package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/store_checkout/internal/application"
	"github.com/voxire/onion/internal/modules/store_checkout/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	http.RegisterPOSTWithValidation[handlers.CreateRequest, handlers.CreateResponse](ctx, router, handlers.CreateRoutes(), &handlers.CreateHandler{Service: service})
	http.RegisterPOSTWithValidation[handlers.CompleteRequest, handlers.CompleteResponse](ctx, router, handlers.CompleteRoutes(), &handlers.CompleteHandler{Service: service})
}
