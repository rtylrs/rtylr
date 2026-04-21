package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/finance_reconciliation/internal/application"
	"github.com/voxire/onion/internal/modules/finance_reconciliation/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	http.RegisterGET[handlers.OverviewRequest, handlers.OverviewResponse](
		ctx, router, handlers.OverviewRoutes(), &handlers.OverviewHandler{Service: service},
	)
	http.RegisterPOSTWithValidation[handlers.MatchRequest, handlers.MatchResponse](
		ctx, router, handlers.MatchRoutes(), &handlers.MatchHandler{Service: service},
	)
}
