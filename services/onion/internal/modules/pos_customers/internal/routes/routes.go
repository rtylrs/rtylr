package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/pos_customers/internal/application"
	"github.com/voxire/onion/internal/modules/pos_customers/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, svc *application.Service) {
	http.RegisterGET[handlers.SearchCustomersRequest, handlers.SearchCustomersResponse](
		ctx, router, handlers.SearchCustomersRoutes(),
		&handlers.SearchCustomersHandler{Service: svc},
	)
}
