package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr_equipment/internal/application"
	"github.com/voxire/onion/internal/modules/hr_equipment/internal/routes/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	serverHttp.RegisterGET[handlers.ListEquipmentRequestsRequest, handlers.ListEquipmentRequestsResponse](
		ctx, router, handlers.ListEquipmentRequestsRoutes(), &handlers.ListEquipmentRequestsHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.CreateEquipmentRequestRequest, handlers.CreateEquipmentRequestResponse](
		ctx, router, handlers.CreateEquipmentRequestRoutes(), &handlers.CreateEquipmentRequestHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.ApproveEquipmentRequestRequest, handlers.ApproveEquipmentRequestResponse](
		ctx, router, handlers.ApproveEquipmentRequestRoutes(), &handlers.ApproveEquipmentRequestHandler{Service: service},
	)
	serverHttp.RegisterGET[handlers.ListAssetsRequest, handlers.ListAssetsResponse](
		ctx, router, handlers.ListAssetsRoutes(), &handlers.ListAssetsHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.CreateAssetRequest, handlers.CreateAssetResponse](
		ctx, router, handlers.CreateAssetRoutes(), &handlers.CreateAssetHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.AssignAssetRequest, handlers.AssignAssetResponse](
		ctx, router, handlers.AssignAssetRoutes(), &handlers.AssignAssetHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.UnassignAssetRequest, handlers.UnassignAssetResponse](
		ctx, router, handlers.UnassignAssetRoutes(), &handlers.UnassignAssetHandler{Service: service},
	)
}
