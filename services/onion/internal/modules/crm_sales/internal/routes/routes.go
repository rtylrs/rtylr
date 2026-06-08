package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/crm_sales/internal/application"
	"github.com/voxire/onion/internal/modules/crm_sales/internal/routes/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	// Pipeline summary
	serverHttp.RegisterGET[handlers.GetPipelineRequest, handlers.GetPipelineResponse](
		ctx, router, handlers.GetPipelineRoutes(), &handlers.GetPipelineHandler{Service: service},
	)

	// Leads
	serverHttp.RegisterGET[handlers.ListLeadsRequest, handlers.ListLeadsResponse](
		ctx, router, handlers.ListLeadsRoutes(), &handlers.ListLeadsHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.CreateLeadRequest, handlers.CreateLeadResponse](
		ctx, router, handlers.CreateLeadRoutes(), &handlers.CreateLeadHandler{Service: service},
	)
	serverHttp.RegisterGET[handlers.GetLeadRequest, handlers.GetLeadResponse](
		ctx, router, handlers.GetLeadRoutes(), &handlers.GetLeadHandler{Service: service},
	)
	serverHttp.RegisterPUTWithValidation[handlers.UpdateLeadRequest, handlers.UpdateLeadResponse](
		ctx, router, handlers.UpdateLeadRoutes(), &handlers.UpdateLeadHandler{Service: service},
	)
	serverHttp.RegisterDELETEWithValidation[handlers.DeleteLeadRequest, handlers.DeleteLeadResponse](
		ctx, router, handlers.DeleteLeadRoutes(), &handlers.DeleteLeadHandler{Service: service},
	)

	// Touchpoints (lead-scoped list lives under /leads/{id}/touchpoints)
	serverHttp.RegisterGET[handlers.ListLeadTouchpointsRequest, handlers.ListLeadTouchpointsResponse](
		ctx, router, handlers.ListLeadTouchpointsRoutes(), &handlers.ListLeadTouchpointsHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.CreateTouchpointRequest, handlers.CreateTouchpointResponse](
		ctx, router, handlers.CreateTouchpointRoutes(), &handlers.CreateTouchpointHandler{Service: service},
	)
	serverHttp.RegisterDELETEWithValidation[handlers.DeleteTouchpointRequest, handlers.DeleteTouchpointResponse](
		ctx, router, handlers.DeleteTouchpointRoutes(), &handlers.DeleteTouchpointHandler{Service: service},
	)

	// Contacts
	serverHttp.RegisterGET[handlers.ListContactsRequest, handlers.ListContactsResponse](
		ctx, router, handlers.ListContactsRoutes(), &handlers.ListContactsHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.CreateContactRequest, handlers.CreateContactResponse](
		ctx, router, handlers.CreateContactRoutes(), &handlers.CreateContactHandler{Service: service},
	)
	serverHttp.RegisterPUTWithValidation[handlers.UpdateContactRequest, handlers.UpdateContactResponse](
		ctx, router, handlers.UpdateContactRoutes(), &handlers.UpdateContactHandler{Service: service},
	)
	serverHttp.RegisterDELETEWithValidation[handlers.DeleteContactRequest, handlers.DeleteContactResponse](
		ctx, router, handlers.DeleteContactRoutes(), &handlers.DeleteContactHandler{Service: service},
	)

	// Contracts
	serverHttp.RegisterGET[handlers.ListContractsRequest, handlers.ListContractsResponse](
		ctx, router, handlers.ListContractsRoutes(), &handlers.ListContractsHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.CreateContractRequest, handlers.CreateContractResponse](
		ctx, router, handlers.CreateContractRoutes(), &handlers.CreateContractHandler{Service: service},
	)
	serverHttp.RegisterGET[handlers.GetContractRequest, handlers.GetContractResponse](
		ctx, router, handlers.GetContractRoutes(), &handlers.GetContractHandler{Service: service},
	)
	serverHttp.RegisterPUTWithValidation[handlers.UpdateContractRequest, handlers.UpdateContractResponse](
		ctx, router, handlers.UpdateContractRoutes(), &handlers.UpdateContractHandler{Service: service},
	)
}
