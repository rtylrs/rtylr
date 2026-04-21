package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/products/internal/application"
	"github.com/voxire/onion/internal/modules/products/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	http.RegisterGET[handlers.ListProductsRequest, handlers.ListProductsResponse](
		ctx,
		router,
		handlers.ListProductsRoutes(),
		&handlers.ListProductsHandler{Service: service},
	)

	http.RegisterPOSTWithValidation[handlers.CreateProductRequest, handlers.CreateProductResponse](
		ctx,
		router,
		handlers.CreateProductRoutes(),
		&handlers.CreateProductHandler{Service: service},
	)

	http.RegisterPUTWithValidation[handlers.UpdateProductRequest, handlers.UpdateProductResponse](
		ctx,
		router,
		handlers.UpdateProductRoutes(),
		&handlers.UpdateProductHandler{Service: service},
	)

	http.RegisterDELETEWithValidation[handlers.DeleteProductRequest, handlers.DeleteProductResponse](
		ctx,
		router,
		handlers.DeleteProductRoutes(),
		&handlers.DeleteProductHandler{Service: service},
	)

	http.RegisterGET[handlers.DashboardRequest, handlers.DashboardResponse](
		ctx,
		router,
		handlers.DashboardRoutes(),
		&handlers.DashboardHandler{Service: service},
	)

	http.RegisterGET[handlers.ReorderSuggestionsRequest, handlers.ReorderSuggestionsResponse](
		ctx,
		router,
		handlers.ReorderSuggestionsRoutes(),
		&handlers.ReorderSuggestionsHandler{Service: service},
	)

	http.RegisterGET[handlers.ExpiryRequest, handlers.ExpiryResponse](
		ctx,
		router,
		handlers.ExpiryRoutes(),
		&handlers.ExpiryHandler{Service: service},
	)

	// Path-param fetch — the FE-expected REST shape. Register helpers
	// don't parse chi path params, so these are wired directly.
	// `barcode` before `{id}` so the specific segment matches first.
	router.Get("/erp/products/barcode/{code}", handlers.BarcodeLookupHandlerFunc(service))
	router.Get("/erp/products/{id}", handlers.GetProductHandlerFunc(service))
}
