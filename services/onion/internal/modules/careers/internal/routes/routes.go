package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/careers/internal/application"
	"github.com/voxire/onion/internal/modules/careers/internal/routes/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

// RegisterRoutes wires the legacy /careers endpoint (global, voxire-
// corporate list) plus the recruit product endpoints (tenant-scoped,
// powering recruit.rtylr.com).
func RegisterRoutes(ctx context.Context, router *chi.Mux, careersService *application.Service) {
	// Legacy: voxire-corporate global careers list. Kept for back-compat
	// while the landing site is migrated to the tenant-scoped recruit
	// flow; once the voxire careers page redirects to recruit.rtylr.com
	// this handler can be retired.
	serverHttp.RegisterGET[handlers.GetActiveCareersRequest, handlers.GetActiveCareersResponse](
		ctx, router, handlers.CareersRoutes(),
		&handlers.CareersHandler{CareersService: careersService},
	)

	// Recruit admin — authenticated, tenant-scoped.
	serverHttp.RegisterGET[handlers.ListRecruitCareersRequest, handlers.ListRecruitCareersResponse](
		ctx, router, handlers.ListRecruitCareersRoutes(),
		&handlers.ListRecruitCareersHandler{CareersService: careersService},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.CreateRecruitCareerRequest, handlers.CreateRecruitCareerResponse](
		ctx, router, handlers.CreateRecruitCareerRoutes(),
		&handlers.CreateRecruitCareerHandler{CareersService: careersService},
	)
	serverHttp.RegisterPUTWithValidation[handlers.UpdateRecruitCareerRequest, handlers.UpdateRecruitCareerResponse](
		ctx, router, handlers.UpdateRecruitCareerRoutes(),
		&handlers.UpdateRecruitCareerHandler{CareersService: careersService},
	)
	serverHttp.RegisterDELETEWithValidation[handlers.DeleteRecruitCareerRequest, handlers.DeleteRecruitCareerResponse](
		ctx, router, handlers.DeleteRecruitCareerRoutes(),
		&handlers.DeleteRecruitCareerHandler{CareersService: careersService},
	)

	serverHttp.RegisterGET[handlers.GetCustomizationRequest, handlers.GetCustomizationResponse](
		ctx, router, handlers.GetRecruitCustomizationRoutes(),
		&handlers.GetCustomizationHandler{CareersService: careersService},
	)
	serverHttp.RegisterPUTWithValidation[handlers.PutCustomizationRequest, handlers.PutCustomizationResponse](
		ctx, router, handlers.PutRecruitCustomizationRoutes(),
		&handlers.PutCustomizationHandler{CareersService: careersService},
	)

	serverHttp.RegisterGET[handlers.ListSubmissionsRequest, handlers.ListSubmissionsResponse](
		ctx, router, handlers.ListSubmissionsRoutes(),
		&handlers.ListSubmissionsHandler{CareersService: careersService},
	)
	serverHttp.RegisterPUTWithValidation[handlers.UpdateSubmissionStatusRequest, handlers.UpdateSubmissionStatusResponse](
		ctx, router, handlers.UpdateSubmissionStatusRoutes(),
		&handlers.UpdateSubmissionStatusHandler{CareersService: careersService},
	)

	// Recruit public — anonymous, powers the candidate-facing site.
	serverHttp.RegisterGETWithValidation[handlers.GetPublicBrandingRequest, handlers.GetPublicBrandingResponse](
		ctx, router, handlers.GetPublicBrandingRoutes(),
		&handlers.GetPublicBrandingHandler{CareersService: careersService},
	)
	serverHttp.RegisterGETWithValidation[handlers.ListPublicCareersRequest, handlers.ListPublicCareersResponse](
		ctx, router, handlers.ListPublicCareersRoutes(),
		&handlers.ListPublicCareersHandler{CareersService: careersService},
	)
	serverHttp.RegisterGETWithValidation[handlers.GetPublicCareerRequest, handlers.GetPublicCareerResponse](
		ctx, router, handlers.GetPublicCareerRoutes(),
		&handlers.GetPublicCareerHandler{CareersService: careersService},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.SubmitApplicationRequest, handlers.SubmitApplicationResponse](
		ctx, router, handlers.SubmitApplicationRoutes(),
		&handlers.SubmitApplicationHandler{CareersService: careersService},
	)
}
