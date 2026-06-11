package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/pos_day_summary/internal/application"
	"github.com/voxire/onion/internal/modules/pos_day_summary/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, svc *application.Service) {
	http.RegisterGETWithValidation[handlers.DaySummaryRequest, handlers.DaySummaryResponse](
		ctx, router, handlers.DaySummaryRoutes(),
		&handlers.DaySummaryHandler{Service: svc},
	)
}
