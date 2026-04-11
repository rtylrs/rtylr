package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/pos_terminals/internal/application"
	"github.com/voxire/onion/internal/modules/pos_terminals/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, svc *application.Service) {
	http.RegisterGET[handlers.ListTerminalsRequest, handlers.ListTerminalsResponse](
		ctx, router, handlers.TerminalListRoutes(),
		&handlers.ListTerminalsHandler{Service: svc},
	)

	http.RegisterPOSTWithValidation[handlers.CreateTerminalRequest, handlers.CreateTerminalResponse](
		ctx, router, handlers.TerminalCreateRoutes(),
		&handlers.CreateTerminalHandler{Service: svc},
	)

	http.RegisterPUTWithValidation[handlers.UpdateTerminalRequest, handlers.UpdateTerminalResponse](
		ctx, router, handlers.TerminalUpdateRoutes(),
		&handlers.UpdateTerminalHandler{Service: svc},
	)

	http.RegisterPOSTWithValidation[handlers.HeartbeatTerminalRequest, handlers.HeartbeatTerminalResponse](
		ctx, router, handlers.TerminalHeartbeatRoutes(),
		&handlers.HeartbeatTerminalHandler{Service: svc},
	)
}
