package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/bills/internal/application"
	"github.com/voxire/onion/internal/modules/bills/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	http.RegisterGET[handlers.ListRequest, handlers.ListResponse](
		ctx, router, handlers.ListRoutes(), &handlers.ListHandler{Service: service},
	)
	http.RegisterPOSTWithValidation[handlers.CreateRequest, handlers.CreateResponse](
		ctx, router, handlers.CreateRoutes(), &handlers.CreateHandler{Service: service},
	)
	http.RegisterPUTWithValidation[handlers.UpdateRequest, handlers.UpdateResponse](
		ctx, router, handlers.UpdateRoutes(), &handlers.UpdateHandler{Service: service},
	)
	http.RegisterGET[handlers.GetBillRequest, handlers.GetBillResponse](
		ctx, router, handlers.GetBillRoutes(), &handlers.GetBillHandler{Service: service},
	)
	http.RegisterPUTWithValidation[handlers.UpdateBillByIDRequest, handlers.UpdateBillByIDResponse](
		ctx, router, handlers.UpdateBillByIDRoutes(), &handlers.UpdateBillByIDHandler{Service: service},
	)
	http.RegisterDELETEWithValidation[handlers.DeleteBillRequest, handlers.DeleteBillResponse](
		ctx, router, handlers.DeleteBillRoutes(), &handlers.DeleteBillHandler{Service: service},
	)
}
