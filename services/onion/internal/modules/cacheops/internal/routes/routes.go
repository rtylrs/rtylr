package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/cacheops/internal/routes/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, r *chi.Mux) {
	serverHttp.RegisterPOSTWithValidation[handlers.InvalidateRequest, handlers.InvalidateResponse](
		ctx,
		r,
		handlers.InvalidateRoutes(),
		&handlers.InvalidateHandler{},
	)
}
