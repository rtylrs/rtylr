package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/customers/internal/application"
	"github.com/voxire/onion/internal/modules/customers/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, crmService *application.Service) {
	// Customers
	http.RegisterGETWithValidation[handlers.ListCustomersRequest, handlers.ListCustomersResponse](
		ctx,
		router,
		handlers.CustomerListRoutes(),
		&handlers.ListCustomersHandler{CRMService: crmService},
	)

	http.RegisterPOSTWithValidation[handlers.CreateCustomerRequest, handlers.CreateCustomerResponse](
		ctx,
		router,
		handlers.CustomerCreateRoutes(),
		&handlers.CreateCustomerHandler{CRMService: crmService},
	)

	http.RegisterGETWithValidation[handlers.GetCustomerRequest, handlers.GetCustomerResponse](
		ctx,
		router,
		handlers.CustomerDetailRoutes(),
		&handlers.GetCustomerHandler{CRMService: crmService},
	)

	http.RegisterPUTWithValidation[handlers.UpdateCustomerRequest, handlers.UpdateCustomerResponse](
		ctx,
		router,
		handlers.CustomerUpdateRoutes(),
		&handlers.UpdateCustomerHandler{CRMService: crmService},
	)

	http.RegisterDELETEWithValidation[handlers.DeleteCustomerRequest, handlers.DeleteCustomerResponse](
		ctx,
		router,
		handlers.CustomerDeleteRoutes(),
		&handlers.DeleteCustomerHandler{CRMService: crmService},
	)

	http.RegisterPOSTWithValidation[handlers.RecordVisitRequest, handlers.RecordVisitResponse](
		ctx,
		router,
		handlers.CustomerRecordVisitRoutes(),
		&handlers.RecordVisitHandler{CRMService: crmService},
	)

	// Loyalty
	http.RegisterGETWithValidation[handlers.ListLoyaltyTransactionsRequest, handlers.ListLoyaltyTransactionsResponse](
		ctx,
		router,
		handlers.LoyaltyTransactionsRoutes(),
		&handlers.ListLoyaltyTransactionsHandler{CRMService: crmService},
	)

	http.RegisterPOSTWithValidation[handlers.AwardPointsRequest, handlers.AwardPointsResponse](
		ctx,
		router,
		handlers.LoyaltyAwardRoutes(),
		&handlers.AwardPointsHandler{CRMService: crmService},
	)

	http.RegisterPOSTWithValidation[handlers.RedeemPointsRequest, handlers.RedeemPointsResponse](
		ctx,
		router,
		handlers.LoyaltyRedeemRoutes(),
		&handlers.RedeemPointsHandler{CRMService: crmService},
	)

	http.RegisterGET[handlers.GetLoyaltyConfigRequest, handlers.GetLoyaltyConfigResponse](
		ctx,
		router,
		handlers.LoyaltyConfigGetRoutes(),
		&handlers.GetLoyaltyConfigHandler{CRMService: crmService},
	)

	http.RegisterPUT[handlers.UpdateLoyaltyConfigRequest, handlers.UpdateLoyaltyConfigResponse](
		ctx,
		router,
		handlers.LoyaltyConfigUpdateRoutes(),
		&handlers.UpdateLoyaltyConfigHandler{CRMService: crmService},
	)

	// Segments
	http.RegisterGET[handlers.ListSegmentsRequest, handlers.ListSegmentsResponse](
		ctx,
		router,
		handlers.SegmentListRoutes(),
		&handlers.ListSegmentsHandler{CRMService: crmService},
	)

	http.RegisterPOSTWithValidation[handlers.CreateSegmentRequest, handlers.CreateSegmentResponse](
		ctx,
		router,
		handlers.SegmentCreateRoutes(),
		&handlers.CreateSegmentHandler{CRMService: crmService},
	)

	http.RegisterPUTWithValidation[handlers.UpdateSegmentRequest, handlers.UpdateSegmentResponse](
		ctx,
		router,
		handlers.SegmentUpdateRoutes(),
		&handlers.UpdateSegmentHandler{CRMService: crmService},
	)

	http.RegisterDELETEWithValidation[handlers.DeleteSegmentRequest, handlers.DeleteSegmentResponse](
		ctx,
		router,
		handlers.SegmentDeleteRoutes(),
		&handlers.DeleteSegmentHandler{CRMService: crmService},
	)

	// Campaigns
	http.RegisterGET[handlers.ListCampaignsRequest, handlers.ListCampaignsResponse](
		ctx,
		router,
		handlers.CampaignListRoutes(),
		&handlers.ListCampaignsHandler{CRMService: crmService},
	)

	http.RegisterPOSTWithValidation[handlers.CreateCampaignRequest, handlers.CreateCampaignResponse](
		ctx,
		router,
		handlers.CampaignCreateRoutes(),
		&handlers.CreateCampaignHandler{CRMService: crmService},
	)

	http.RegisterPUTWithValidation[handlers.UpdateCampaignRequest, handlers.UpdateCampaignResponse](
		ctx,
		router,
		handlers.CampaignUpdateRoutes(),
		&handlers.UpdateCampaignHandler{CRMService: crmService},
	)

	http.RegisterPUTWithValidation[handlers.UpdateCampaignStatusRequest, handlers.UpdateCampaignStatusResponse](
		ctx,
		router,
		handlers.CampaignStatusRoutes(),
		&handlers.UpdateCampaignStatusHandler{CRMService: crmService},
	)

	// Dashboard
	http.RegisterGET[handlers.DashboardRequest, handlers.DashboardResponse](
		ctx,
		router,
		handlers.DashboardRoutes(),
		&handlers.DashboardHandler{CRMService: crmService},
	)
}
