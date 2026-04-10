package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/categories/internal/application"
	"github.com/voxire/onion/internal/modules/categories/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	http.RegisterGET[handlers.ListCategoriesRequest, handlers.ListCategoriesResponse](
		ctx,
		router,
		handlers.ListCategoriesRoutes(),
		&handlers.ListCategoriesHandler{Service: service},
	)

	http.RegisterPOSTWithValidation[handlers.CreateCategoryRequest, handlers.CreateCategoryResponse](
		ctx,
		router,
		handlers.CreateCategoryRoutes(),
		&handlers.CreateCategoryHandler{Service: service},
	)

	http.RegisterPUTWithValidation[handlers.UpdateCategoryRequest, handlers.UpdateCategoryResponse](
		ctx,
		router,
		handlers.UpdateCategoryRoutes(),
		&handlers.UpdateCategoryHandler{Service: service},
	)

	http.RegisterDELETEWithValidation[handlers.DeleteCategoryRequest, handlers.DeleteCategoryResponse](
		ctx,
		router,
		handlers.DeleteCategoryRoutes(),
		&handlers.DeleteCategoryHandler{Service: service},
	)
}
