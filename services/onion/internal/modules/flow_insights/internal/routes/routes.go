package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/flow_insights/internal/application"
	"github.com/voxire/onion/internal/modules/flow_insights/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	// GET /flow/insights
	http.RegisterGET[handlers.GetInsightsRequest, handlers.GetInsightsResponse](
		ctx, router, handlers.GetInsightsRoutes(),
		&handlers.GetInsightsHandler{Service: service},
	)

	// GET /flow/team/workload
	http.RegisterGET[handlers.GetTeamWorkloadRequest, handlers.GetTeamWorkloadResponse](
		ctx, router, handlers.GetTeamWorkloadRoutes(),
		&handlers.GetTeamWorkloadHandler{Service: service},
	)

	// GET /flow/roadmap
	http.RegisterGET[handlers.GetRoadmapRequest, handlers.GetRoadmapResponse](
		ctx, router, handlers.GetRoadmapRoutes(),
		&handlers.GetRoadmapHandler{Service: service},
	)
}
