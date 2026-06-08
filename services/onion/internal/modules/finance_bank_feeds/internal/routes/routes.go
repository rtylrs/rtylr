package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/finance_bank_feeds/internal/application"
	"github.com/voxire/onion/internal/modules/finance_bank_feeds/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	http.RegisterGET[handlers.ListRequest, handlers.ListResponse](
		ctx, router, handlers.ListRoutes(), &handlers.ListHandler{Service: service},
	)
	http.RegisterPOSTWithValidation[handlers.SyncRequest, handlers.SyncResponse](
		ctx, router, handlers.SyncRoutes(), &handlers.SyncHandler{Service: service},
	)
}
