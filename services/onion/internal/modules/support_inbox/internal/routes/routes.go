package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/support_inbox/internal/application"
	"github.com/voxire/onion/internal/modules/support_inbox/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	http.RegisterGET[handlers.ListInboxRequest, handlers.ListInboxResponse](
		ctx, router, handlers.ListInboxRoutes(),
		&handlers.ListInboxHandler{Service: service},
	)

	http.RegisterGET[handlers.UnreadCountRequest, handlers.UnreadCountResponse](
		ctx, router, handlers.UnreadCountRoutes(),
		&handlers.UnreadCountHandler{Service: service},
	)
}
