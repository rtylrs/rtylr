package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/quotes/internal/application"
	"github.com/voxire/onion/internal/modules/quotes/internal/routes/handlers"
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
	http.RegisterDELETEWithValidation[handlers.DeleteRequest, handlers.DeleteResponse](
		ctx, router, handlers.DeleteRoutes(), &handlers.DeleteHandler{Service: service},
	)
	http.RegisterPUTWithValidation[handlers.UpdateQuoteByIDRequest, handlers.UpdateQuoteByIDResponse](
		ctx, router, handlers.UpdateQuoteByIDRoutes(), &handlers.UpdateQuoteByIDHandler{Service: service},
	)
	http.RegisterPUTWithValidation[handlers.UpdateQuoteStatusRequest, handlers.UpdateQuoteStatusResponse](
		ctx, router, handlers.UpdateQuoteStatusRoutes(), &handlers.UpdateQuoteStatusHandler{Service: service},
	)
	http.RegisterDELETEWithValidation[handlers.DeleteQuoteByIDRequest, handlers.DeleteQuoteByIDResponse](
		ctx, router, handlers.DeleteQuoteByIDRoutes(), &handlers.DeleteQuoteByIDHandler{Service: service},
	)
	http.RegisterPOSTWithValidation[handlers.ConvertQuoteRequest, handlers.ConvertQuoteResponse](
		ctx, router, handlers.ConvertQuoteRoutes(), &handlers.ConvertQuoteHandler{Service: service},
	)
}
