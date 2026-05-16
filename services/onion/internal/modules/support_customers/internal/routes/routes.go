package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/support_customers/internal/application"
	"github.com/voxire/onion/internal/modules/support_customers/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	http.RegisterGET[handlers.ListCustomersRequest, handlers.ListCustomersResponse](
		ctx, router, handlers.ListCustomersRoutes(),
		&handlers.ListCustomersHandler{Service: service},
	)

	http.RegisterGETWithValidation[handlers.GetCustomerRequest, handlers.GetCustomerResponse](
		ctx, router, handlers.GetCustomerRoutes(),
		&handlers.GetCustomerHandler{Service: service},
	)
}
