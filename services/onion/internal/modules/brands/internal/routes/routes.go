package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/brands/internal/application"
	"github.com/voxire/onion/internal/modules/brands/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	http.RegisterGET[handlers.ListBrandsRequest, handlers.ListBrandsResponse](
		ctx,
		router,
		handlers.ListBrandsRoutes(),
		&handlers.ListBrandsHandler{Service: service},
	)

	http.RegisterPOSTWithValidation[handlers.CreateBrandRequest, handlers.CreateBrandResponse](
		ctx,
		router,
		handlers.CreateBrandRoutes(),
		&handlers.CreateBrandHandler{Service: service},
	)

	http.RegisterPUTWithValidation[handlers.UpdateBrandRequest, handlers.UpdateBrandResponse](
		ctx,
		router,
		handlers.UpdateBrandRoutes(),
		&handlers.UpdateBrandHandler{Service: service},
	)

	http.RegisterDELETEWithValidation[handlers.DeleteBrandRequest, handlers.DeleteBrandResponse](
		ctx,
		router,
		handlers.DeleteBrandRoutes(),
		&handlers.DeleteBrandHandler{Service: service},
	)
}
