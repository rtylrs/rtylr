package api

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/crm/customers/internal/application"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, crmService *application.Service) {
	// Customers
	http.RegisterGETWithValidation[ListCustomersRequest, ListCustomersResponse](
		ctx,
		router,
		CustomerListRoutes(),
		&ListCustomersHandler{CRMService: crmService},
	)

	http.RegisterPOSTWithValidation[CreateCustomerRequest, CreateCustomerResponse](
		ctx,
		router,
		CustomerCreateRoutes(),
		&CreateCustomerHandler{CRMService: crmService},
	)

	http.RegisterGETWithValidation[GetCustomerRequest, GetCustomerResponse](
		ctx,
		router,
		CustomerDetailRoutes(),
		&GetCustomerHandler{CRMService: crmService},
	)

	http.RegisterPUTWithValidation[UpdateCustomerRequest, UpdateCustomerResponse](
		ctx,
		router,
		CustomerUpdateRoutes(),
		&UpdateCustomerHandler{CRMService: crmService},
	)

	http.RegisterDELETEWithValidation[DeleteCustomerRequest, DeleteCustomerResponse](
		ctx,
		router,
		CustomerDeleteRoutes(),
		&DeleteCustomerHandler{CRMService: crmService},
	)

	http.RegisterPOSTWithValidation[RecordVisitRequest, RecordVisitResponse](
		ctx,
		router,
		CustomerRecordVisitRoutes(),
		&RecordVisitHandler{CRMService: crmService},
	)

	// Loyalty
	http.RegisterGETWithValidation[ListLoyaltyTransactionsRequest, ListLoyaltyTransactionsResponse](
		ctx,
		router,
		LoyaltyTransactionsRoutes(),
		&ListLoyaltyTransactionsHandler{CRMService: crmService},
	)

	http.RegisterPOSTWithValidation[AwardPointsRequest, AwardPointsResponse](
		ctx,
		router,
		LoyaltyAwardRoutes(),
		&AwardPointsHandler{CRMService: crmService},
	)

	http.RegisterPOSTWithValidation[RedeemPointsRequest, RedeemPointsResponse](
		ctx,
		router,
		LoyaltyRedeemRoutes(),
		&RedeemPointsHandler{CRMService: crmService},
	)

	http.RegisterGET[GetLoyaltyConfigRequest, GetLoyaltyConfigResponse](
		ctx,
		router,
		LoyaltyConfigGetRoutes(),
		&GetLoyaltyConfigHandler{CRMService: crmService},
	)

	http.RegisterPUTWithValidation[UpdateLoyaltyConfigRequest, UpdateLoyaltyConfigResponse](
		ctx,
		router,
		LoyaltyConfigUpdateRoutes(),
		&UpdateLoyaltyConfigHandler{CRMService: crmService},
	)

	// Segments
	http.RegisterGET[ListSegmentsRequest, ListSegmentsResponse](
		ctx,
		router,
		SegmentListRoutes(),
		&ListSegmentsHandler{CRMService: crmService},
	)

	http.RegisterPOSTWithValidation[CreateSegmentRequest, CreateSegmentResponse](
		ctx,
		router,
		SegmentCreateRoutes(),
		&CreateSegmentHandler{CRMService: crmService},
	)

	http.RegisterPUTWithValidation[UpdateSegmentRequest, UpdateSegmentResponse](
		ctx,
		router,
		SegmentUpdateRoutes(),
		&UpdateSegmentHandler{CRMService: crmService},
	)

	http.RegisterDELETEWithValidation[DeleteSegmentRequest, DeleteSegmentResponse](
		ctx,
		router,
		SegmentDeleteRoutes(),
		&DeleteSegmentHandler{CRMService: crmService},
	)

	// Dashboard
	http.RegisterGET[DashboardRequest, DashboardResponse](
		ctx,
		router,
		DashboardRoutes(),
		&DashboardHandler{CRMService: crmService},
	)

	// Customer Timeline
	http.RegisterGETWithValidation[GetTimelineRequest, GetTimelineResponse](
		ctx,
		router,
		CustomerTimelineRoutes(),
		&GetTimelineHandler{CRMService: crmService},
	)

	// Customer Detail (full with loyalty + visits)
	http.RegisterGETWithValidation[GetCustomerDetailFullRequest, GetCustomerDetailFullResponse](
		ctx,
		router,
		CustomerDetailFullRoutes(),
		&GetCustomerDetailFullHandler{CRMService: crmService},
	)

	// Customer Import
	http.RegisterPOSTWithValidation[ImportCustomersRequest, ImportCustomersResponse](
		ctx,
		router,
		CustomerImportRoutes(),
		&ImportCustomersHandler{CRMService: crmService},
	)

	// Customer Export (GDPR)
	http.RegisterGETWithValidation[ExportCustomerRequest, ExportCustomerResponse](
		ctx,
		router,
		CustomerExportRoutes(),
		&ExportCustomerHandler{CRMService: crmService},
	)

	// Customer Merge
	http.RegisterPOSTWithValidation[MergeCustomersRequest, MergeCustomersResponse](
		ctx,
		router,
		CustomerMergeRoutes(),
		&MergeCustomersHandler{CRMService: crmService},
	)

	// Loyalty Tiers
	http.RegisterGETWithValidation[ListLoyaltyTiersRequest, ListLoyaltyTiersResponse](
		ctx,
		router,
		LoyaltyTiersListRoutes(),
		&ListLoyaltyTiersHandler{CRMService: crmService},
	)

	http.RegisterPOSTWithValidation[CreateLoyaltyTierRequest, CreateLoyaltyTierResponse](
		ctx,
		router,
		LoyaltyTiersCreateRoutes(),
		&CreateLoyaltyTierHandler{CRMService: crmService},
	)

	http.RegisterPUTWithValidation[UpdateLoyaltyTierRequest, UpdateLoyaltyTierResponse](
		ctx,
		router,
		LoyaltyTiersUpdateRoutes(),
		&UpdateLoyaltyTierHandler{CRMService: crmService},
	)

	http.RegisterDELETEWithValidation[DeleteLoyaltyTierRequest, DeleteLoyaltyTierResponse](
		ctx,
		router,
		LoyaltyTiersDeleteRoutes(),
		&DeleteLoyaltyTierHandler{CRMService: crmService},
	)

	// Campaign Stats
	http.RegisterGETWithValidation[GetCampaignStatsRequest, GetCampaignStatsResponse](
		ctx,
		router,
		CampaignStatsRoutes(),
		&GetCampaignStatsHandler{CRMService: crmService},
	)

	// Campaign Detail
	http.RegisterGETWithValidation[GetCampaignDetailRequest, GetCampaignDetailResponse](
		ctx,
		router,
		CampaignDetailFullRoutes(),
		&GetCampaignDetailHandler{CRMService: crmService},
	)

	// Segment Detail
	http.RegisterGETWithValidation[GetSegmentDetailRequest, GetSegmentDetailResponse](
		ctx,
		router,
		SegmentDetailRoutes(),
		&GetSegmentDetailHandler{CRMService: crmService},
	)

	// Segment Customers
	http.RegisterGETWithValidation[GetSegmentCustomersRequest, GetSegmentCustomersResponse](
		ctx,
		router,
		SegmentCustomersRoutes(),
		&GetSegmentCustomersHandler{CRMService: crmService},
	)
}
