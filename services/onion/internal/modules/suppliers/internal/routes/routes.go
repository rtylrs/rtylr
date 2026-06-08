package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/suppliers/internal/application"
	"github.com/voxire/onion/internal/modules/suppliers/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	http.RegisterGET[handlers.ListSuppliersRequest, handlers.ListSuppliersResponse](
		ctx,
		router,
		handlers.ListSuppliersRoutes(),
		&handlers.ListSuppliersHandler{Service: service},
	)

	http.RegisterPOSTWithValidation[handlers.CreateSupplierRequest, handlers.CreateSupplierResponse](
		ctx,
		router,
		handlers.CreateSupplierRoutes(),
		&handlers.CreateSupplierHandler{Service: service},
	)

	http.RegisterPUTWithValidation[handlers.UpdateSupplierRequest, handlers.UpdateSupplierResponse](
		ctx,
		router,
		handlers.UpdateSupplierRoutes(),
		&handlers.UpdateSupplierHandler{Service: service},
	)

	// PUT /erp/suppliers/{id} — REST path variant; coexists with the body-ID route above.
	http.RegisterPUTWithValidation[handlers.UpdateSupplierByIDRequest, handlers.UpdateSupplierByIDResponse](
		ctx,
		router,
		handlers.UpdateSupplierByIDRoutes(),
		&handlers.UpdateSupplierByIDHandler{Service: service},
	)

	// DELETE /erp/suppliers/{id} — new REST path variant.
	http.RegisterDELETEWithValidation[handlers.DeleteSupplierByIDRequest, handlers.DeleteSupplierByIDResponse](
		ctx,
		router,
		handlers.DeleteSupplierByIDRoutes(),
		&handlers.DeleteSupplierByIDHandler{Service: service},
	)
}
