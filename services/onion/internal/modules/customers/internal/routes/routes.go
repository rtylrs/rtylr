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

	http.RegisterPUTWithValidation[handlers.UpdateLoyaltyConfigRequest, handlers.UpdateLoyaltyConfigResponse](
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

	// Dashboard
	http.RegisterGET[handlers.DashboardRequest, handlers.DashboardResponse](
		ctx,
		router,
		handlers.DashboardRoutes(),
		&handlers.DashboardHandler{CRMService: crmService},
	)

	// Customer Timeline
	http.RegisterGETWithValidation[handlers.GetTimelineRequest, handlers.GetTimelineResponse](
		ctx,
		router,
		handlers.CustomerTimelineRoutes(),
		&handlers.GetTimelineHandler{CRMService: crmService},
	)

	// Customer Detail (full with loyalty + visits)
	http.RegisterGETWithValidation[handlers.GetCustomerDetailFullRequest, handlers.GetCustomerDetailFullResponse](
		ctx,
		router,
		handlers.CustomerDetailFullRoutes(),
		&handlers.GetCustomerDetailFullHandler{CRMService: crmService},
	)

	// Customer Import
	http.RegisterPOSTWithValidation[handlers.ImportCustomersRequest, handlers.ImportCustomersResponse](
		ctx,
		router,
		handlers.CustomerImportRoutes(),
		&handlers.ImportCustomersHandler{CRMService: crmService},
	)

	// Customer Export (GDPR)
	http.RegisterGETWithValidation[handlers.ExportCustomerRequest, handlers.ExportCustomerResponse](
		ctx,
		router,
		handlers.CustomerExportRoutes(),
		&handlers.ExportCustomerHandler{CRMService: crmService},
	)

	// Customer Merge
	http.RegisterPOSTWithValidation[handlers.MergeCustomersRequest, handlers.MergeCustomersResponse](
		ctx,
		router,
		handlers.CustomerMergeRoutes(),
		&handlers.MergeCustomersHandler{CRMService: crmService},
	)

	// Loyalty Tiers
	http.RegisterGET[handlers.ListLoyaltyTiersRequest, handlers.ListLoyaltyTiersResponse](
		ctx,
		router,
		handlers.LoyaltyTiersListRoutes(),
		&handlers.ListLoyaltyTiersHandler{CRMService: crmService},
	)

	http.RegisterPOSTWithValidation[handlers.CreateLoyaltyTierRequest, handlers.CreateLoyaltyTierResponse](
		ctx,
		router,
		handlers.LoyaltyTiersCreateRoutes(),
		&handlers.CreateLoyaltyTierHandler{CRMService: crmService},
	)

	http.RegisterPUTWithValidation[handlers.UpdateLoyaltyTierRequest, handlers.UpdateLoyaltyTierResponse](
		ctx,
		router,
		handlers.LoyaltyTiersUpdateRoutes(),
		&handlers.UpdateLoyaltyTierHandler{CRMService: crmService},
	)

	http.RegisterDELETEWithValidation[handlers.DeleteLoyaltyTierRequest, handlers.DeleteLoyaltyTierResponse](
		ctx,
		router,
		handlers.LoyaltyTiersDeleteRoutes(),
		&handlers.DeleteLoyaltyTierHandler{CRMService: crmService},
	)

	// Campaign Stats
	http.RegisterGETWithValidation[handlers.GetCampaignStatsRequest, handlers.GetCampaignStatsResponse](
		ctx,
		router,
		handlers.CampaignStatsRoutes(),
		&handlers.GetCampaignStatsHandler{CRMService: crmService},
	)

	// Campaign Detail
	http.RegisterGETWithValidation[handlers.GetCampaignDetailRequest, handlers.GetCampaignDetailResponse](
		ctx,
		router,
		handlers.CampaignDetailFullRoutes(),
		&handlers.GetCampaignDetailHandler{CRMService: crmService},
	)

	// Segment Detail
	http.RegisterGETWithValidation[handlers.GetSegmentDetailRequest, handlers.GetSegmentDetailResponse](
		ctx,
		router,
		handlers.SegmentDetailRoutes(),
		&handlers.GetSegmentDetailHandler{CRMService: crmService},
	)

	// Segment Customers
	http.RegisterGETWithValidation[handlers.GetSegmentCustomersRequest, handlers.GetSegmentCustomersResponse](
		ctx,
		router,
		handlers.SegmentCustomersRoutes(),
		&handlers.GetSegmentCustomersHandler{CRMService: crmService},
	)
}
