package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/pos_sessions/internal/application"
	"github.com/voxire/onion/internal/modules/pos_sessions/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, svc *application.Service) {
	http.RegisterGET[handlers.ListSessionsRequest, handlers.ListSessionsResponse](
		ctx, router, handlers.SessionListRoutes(),
		&handlers.ListSessionsHandler{Service: svc},
	)

	http.RegisterGET[handlers.GetSessionRequest, handlers.GetSessionResponse](
		ctx, router, handlers.SessionGetRoutes(),
		&handlers.GetSessionHandler{Service: svc},
	)

	http.RegisterPOSTWithValidation[handlers.OpenSessionRequest, handlers.OpenSessionResponse](
		ctx, router, handlers.SessionOpenRoutes(),
		&handlers.OpenSessionHandler{Service: svc},
	)

	http.RegisterPOSTWithValidation[handlers.CloseSessionRequest, handlers.CloseSessionResponse](
		ctx, router, handlers.SessionCloseRoutes(),
		&handlers.CloseSessionHandler{Service: svc},
	)
}
