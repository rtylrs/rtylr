package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/insights/internal/application"
	"github.com/voxire/onion/internal/modules/insights/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, svc *application.Service) {
	http.RegisterGET[handlers.GetInsightsRequest, handlers.GetInsightsResponse](
		ctx, router, handlers.InsightsRoutes(),
		&handlers.GetInsightsHandler{Service: svc},
	)

	http.RegisterGET[handlers.GetWorkloadRequest, handlers.GetWorkloadResponse](
		ctx, router, handlers.WorkloadRoutes(),
		&handlers.GetWorkloadHandler{Service: svc},
	)

	http.RegisterGET[handlers.GetRoadmapRequest, handlers.GetRoadmapResponse](
		ctx, router, handlers.RoadmapRoutes(),
		&handlers.GetRoadmapHandler{Service: svc},
	)
}
