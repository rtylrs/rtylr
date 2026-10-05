package api

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/application"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	// Pipeline summary
	serverHttp.RegisterGET[GetPipelineRequest, GetPipelineResponse](
		ctx, router, GetPipelineRoutes(), &GetPipelineHandler{Service: service},
	)

	// Leads
	serverHttp.RegisterGET[ListLeadsRequest, ListLeadsResponse](
		ctx, router, ListLeadsRoutes(), &ListLeadsHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[CreateLeadRequest, CreateLeadResponse](
		ctx, router, CreateLeadRoutes(), &CreateLeadHandler{Service: service},
	)
	serverHttp.RegisterGET[GetLeadRequest, GetLeadResponse](
		ctx, router, GetLeadRoutes(), &GetLeadHandler{Service: service},
	)
	serverHttp.RegisterPUTWithValidation[UpdateLeadRequest, UpdateLeadResponse](
		ctx, router, UpdateLeadRoutes(), &UpdateLeadHandler{Service: service},
	)
	serverHttp.RegisterDELETEWithValidation[DeleteLeadRequest, DeleteLeadResponse](
		ctx, router, DeleteLeadRoutes(), &DeleteLeadHandler{Service: service},
	)

	// Touchpoints (lead-scoped list lives under /leads/{id}/touchpoints)
	serverHttp.RegisterGET[ListLeadTouchpointsRequest, ListLeadTouchpointsResponse](
		ctx, router, ListLeadTouchpointsRoutes(), &ListLeadTouchpointsHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[CreateTouchpointRequest, CreateTouchpointResponse](
		ctx, router, CreateTouchpointRoutes(), &CreateTouchpointHandler{Service: service},
	)
	serverHttp.RegisterDELETEWithValidation[DeleteTouchpointRequest, DeleteTouchpointResponse](
		ctx, router, DeleteTouchpointRoutes(), &DeleteTouchpointHandler{Service: service},
	)

	// Contacts
	serverHttp.RegisterGET[ListContactsRequest, ListContactsResponse](
		ctx, router, ListContactsRoutes(), &ListContactsHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[CreateContactRequest, CreateContactResponse](
		ctx, router, CreateContactRoutes(), &CreateContactHandler{Service: service},
	)
	serverHttp.RegisterPUTWithValidation[UpdateContactRequest, UpdateContactResponse](
		ctx, router, UpdateContactRoutes(), &UpdateContactHandler{Service: service},
	)
	serverHttp.RegisterDELETEWithValidation[DeleteContactRequest, DeleteContactResponse](
		ctx, router, DeleteContactRoutes(), &DeleteContactHandler{Service: service},
	)

	// Contracts
	serverHttp.RegisterGET[ListContractsRequest, ListContractsResponse](
		ctx, router, ListContractsRoutes(), &ListContractsHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[CreateContractRequest, CreateContractResponse](
		ctx, router, CreateContractRoutes(), &CreateContractHandler{Service: service},
	)
	serverHttp.RegisterGET[GetContractRequest, GetContractResponse](
		ctx, router, GetContractRoutes(), &GetContractHandler{Service: service},
	)
	serverHttp.RegisterPUTWithValidation[UpdateContractRequest, UpdateContractResponse](
		ctx, router, UpdateContractRoutes(), &UpdateContractHandler{Service: service},
	)
}
