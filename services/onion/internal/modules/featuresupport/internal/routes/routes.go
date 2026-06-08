package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/featuresupport/internal/application"
	"github.com/voxire/onion/internal/modules/featuresupport/internal/routes/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	// Reads.
	serverHttp.RegisterPOSTWithValidation[handlers.GetFeaturesRequest, handlers.FeaturesResponse](
		ctx, router, handlers.GetFeaturesRoutes(), &handlers.GetFeaturesHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.GetMyFeaturesRequest, handlers.FeaturesResponse](
		ctx, router, handlers.GetMyFeaturesRoutes(), &handlers.GetMyFeaturesHandler{Service: service},
	)

	// Admin: flags.
	serverHttp.RegisterGET[handlers.ListFlagsRequest, handlers.ListFlagsResponse](
		ctx, router, handlers.ListFlagsRoutes(), &handlers.ListFlagsHandler{Service: service},
	)
	serverHttp.RegisterPUTWithValidation[handlers.UpsertFlagRequest, handlers.UpsertFlagResponse](
		ctx, router, handlers.UpsertFlagRoutes(), &handlers.UpsertFlagHandler{Service: service},
	)
	serverHttp.RegisterDELETEWithValidation[handlers.DeleteFlagRequest, handlers.DeleteFlagResponse](
		ctx, router, handlers.DeleteFlagRoutes(), &handlers.DeleteFlagHandler{Service: service},
	)

	// Admin: per-user targets.
	serverHttp.RegisterGET[handlers.ListTargetsRequest, handlers.ListTargetsResponse](
		ctx, router, handlers.ListTargetsRoutes(), &handlers.ListTargetsHandler{Service: service},
	)
	serverHttp.RegisterPUTWithValidation[handlers.UpsertTargetRequest, handlers.UpsertTargetResponse](
		ctx, router, handlers.UpsertTargetRoutes(), &handlers.UpsertTargetHandler{Service: service},
	)
	serverHttp.RegisterDELETEWithValidation[handlers.DeleteTargetRequest, handlers.DeleteTargetResponse](
		ctx, router, handlers.DeleteTargetRoutes(), &handlers.DeleteTargetHandler{Service: service},
	)

	// Admin: targeting rules + conditions.
	serverHttp.RegisterGET[handlers.ListRulesRequest, handlers.ListRulesResponse](
		ctx, router, handlers.ListRulesRoutes(), &handlers.ListRulesHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.CreateRuleRequest, handlers.CreateRuleResponse](
		ctx, router, handlers.CreateRuleRoutes(), &handlers.CreateRuleHandler{Service: service},
	)
	serverHttp.RegisterPUTWithValidation[handlers.UpdateRuleRequest, handlers.UpdateRuleResponse](
		ctx, router, handlers.UpdateRuleRoutes(), &handlers.UpdateRuleHandler{Service: service},
	)
	serverHttp.RegisterDELETEWithValidation[handlers.DeleteRuleRequest, handlers.DeleteRuleResponse](
		ctx, router, handlers.DeleteRuleRoutes(), &handlers.DeleteRuleHandler{Service: service},
	)
}
