package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/journal/internal/application"
	"github.com/voxire/onion/internal/modules/journal/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	http.RegisterGET[handlers.ListRequest, handlers.ListResponse](
		ctx, router, handlers.ListRoutes(), &handlers.ListHandler{Service: service},
	)
	http.RegisterPOSTWithValidation[handlers.CreateRequest, handlers.CreateResponse](
		ctx, router, handlers.CreateRoutes(), &handlers.CreateHandler{Service: service},
	)
	http.RegisterPOSTWithValidation[handlers.PostRequest, handlers.PostResponse](
		ctx, router, handlers.PostRoutes(), &handlers.PostHandler{Service: service},
	)
}
