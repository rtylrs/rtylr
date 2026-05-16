package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/documents/internal/application"
	"github.com/voxire/onion/internal/modules/documents/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	http.RegisterGET[handlers.ListRequest, handlers.ListResponse](
		ctx, router, handlers.ListRoutes(), &handlers.ListHandler{Service: service},
	)
	http.RegisterPOSTWithValidation[handlers.CreateRequest, handlers.CreateResponse](
		ctx, router, handlers.CreateRoutes(), &handlers.CreateHandler{Service: service},
	)
	http.RegisterDELETEWithValidation[handlers.DeleteRequest, handlers.DeleteResponse](
		ctx, router, handlers.DeleteRoutes(), &handlers.DeleteHandler{Service: service},
	)
}
