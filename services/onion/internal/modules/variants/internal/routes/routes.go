package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/variants/internal/application"
	"github.com/voxire/onion/internal/modules/variants/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	http.RegisterGET[handlers.ListVariantsRequest, handlers.ListVariantsResponse](
		ctx,
		router,
		handlers.ListVariantsRoutes(),
		&handlers.ListVariantsHandler{Service: service},
	)

	http.RegisterPOSTWithValidation[handlers.CreateVariantRequest, handlers.CreateVariantResponse](
		ctx,
		router,
		handlers.CreateVariantRoutes(),
		&handlers.CreateVariantHandler{Service: service},
	)

	http.RegisterPUTWithValidation[handlers.UpdateVariantRequest, handlers.UpdateVariantResponse](
		ctx,
		router,
		handlers.UpdateVariantRoutes(),
		&handlers.UpdateVariantHandler{Service: service},
	)

	http.RegisterDELETEWithValidation[handlers.DeleteVariantRequest, handlers.DeleteVariantResponse](
		ctx,
		router,
		handlers.DeleteVariantRoutes(),
		&handlers.DeleteVariantHandler{Service: service},
	)
}
