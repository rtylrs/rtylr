package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr_performance/internal/application"
	"github.com/voxire/onion/internal/modules/hr_performance/internal/routes/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	// Config
	serverHttp.RegisterGET[handlers.GetConfigRequest, handlers.GetConfigResponse](
		ctx, router, handlers.GetConfigRoutes(), &handlers.GetConfigHandler{Service: service},
	)
	serverHttp.RegisterPUTWithValidation[handlers.UpsertConfigRequest, handlers.UpsertConfigResponse](
		ctx, router, handlers.UpsertConfigRoutes(), &handlers.UpsertConfigHandler{Service: service},
	)

	// Review Cycles
	serverHttp.RegisterGET[handlers.ListCyclesRequest, handlers.ListCyclesResponse](
		ctx, router, handlers.ListCyclesRoutes(), &handlers.ListCyclesHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.CreateCycleRequest, handlers.CreateCycleResponse](
		ctx, router, handlers.CreateCycleRoutes(), &handlers.CreateCycleHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.OpenCycleRequest, handlers.OpenCycleResponse](
		ctx, router, handlers.OpenCycleRoutes(), &handlers.OpenCycleHandler{Service: service},
	)

	// Goals
	serverHttp.RegisterGET[handlers.ListGoalsRequest, handlers.ListGoalsResponse](
		ctx, router, handlers.ListGoalsRoutes(), &handlers.ListGoalsHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.CreateGoalRequest, handlers.CreateGoalResponse](
		ctx, router, handlers.CreateGoalRoutes(), &handlers.CreateGoalHandler{Service: service},
	)
	serverHttp.RegisterPUTWithValidation[handlers.UpdateGoalRequest, handlers.UpdateGoalResponse](
		ctx, router, handlers.UpdateGoalRoutes(), &handlers.UpdateGoalHandler{Service: service},
	)

	// Reviews
	serverHttp.RegisterGET[handlers.ListReviewsRequest, handlers.ListReviewsResponse](
		ctx, router, handlers.ListReviewsRoutes(), &handlers.ListReviewsHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.SubmitReviewRequest, handlers.SubmitReviewResponse](
		ctx, router, handlers.SubmitReviewRoutes(), &handlers.SubmitReviewHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.ShareReviewRequest, handlers.ShareReviewResponse](
		ctx, router, handlers.ShareReviewRoutes(), &handlers.ShareReviewHandler{Service: service},
	)
}
