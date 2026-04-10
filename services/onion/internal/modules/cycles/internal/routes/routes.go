package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/cycles/internal/application"
	"github.com/voxire/onion/internal/modules/cycles/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, svc *application.Service) {
	http.RegisterGETWithValidation[handlers.ListCyclesRequest, handlers.ListCyclesResponse](
		ctx, router, handlers.CycleListRoutes(),
		&handlers.ListCyclesHandler{Service: svc},
	)

	http.RegisterPOSTWithValidation[handlers.CreateCycleRequest, handlers.CreateCycleResponse](
		ctx, router, handlers.CycleCreateRoutes(),
		&handlers.CreateCycleHandler{Service: svc},
	)

	http.RegisterPUTWithValidation[handlers.UpdateCycleRequest, handlers.UpdateCycleResponse](
		ctx, router, handlers.CycleUpdateRoutes(),
		&handlers.UpdateCycleHandler{Service: svc},
	)
}
