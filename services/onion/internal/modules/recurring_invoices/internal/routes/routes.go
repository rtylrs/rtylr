package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/recurring_invoices/internal/application"
	"github.com/voxire/onion/internal/modules/recurring_invoices/internal/routes/handlers"
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

	http.RegisterPOSTWithValidation[handlers.PauseRequest, handlers.PauseResponse](
		ctx, router, handlers.PauseRoutes(), &handlers.PauseHandler{Service: service},
	)

	http.RegisterPOSTWithValidation[handlers.ResumeRequest, handlers.ResumeResponse](
		ctx, router, handlers.ResumeRoutes(), &handlers.ResumeHandler{Service: service},
	)
}
