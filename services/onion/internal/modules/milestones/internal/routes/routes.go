package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/milestones/internal/application"
	"github.com/voxire/onion/internal/modules/milestones/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, svc *application.Service) {
	http.RegisterGETWithValidation[handlers.ListMilestonesRequest, handlers.ListMilestonesResponse](
		ctx, router, handlers.MilestoneListRoutes(),
		&handlers.ListMilestonesHandler{Service: svc},
	)

	http.RegisterPOSTWithValidation[handlers.CreateMilestoneRequest, handlers.CreateMilestoneResponse](
		ctx, router, handlers.MilestoneCreateRoutes(),
		&handlers.CreateMilestoneHandler{Service: svc},
	)

	http.RegisterPUTWithValidation[handlers.UpdateMilestoneRequest, handlers.UpdateMilestoneResponse](
		ctx, router, handlers.MilestoneUpdateRoutes(),
		&handlers.UpdateMilestoneHandler{Service: svc},
	)
}
