package routes

import (
	"context"
	"time"

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
	serverHttp.RegisterGET[handlers.ExportSubmissionsRequest, handlers.ExportSubmissionsResponse](
		ctx, router, handlers.ExportSubmissionsRoutes(),
		&handlers.ExportSubmissionsHandler{CareersService: careersService},
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
	// Public apply — 5/min/IP. This is the single biggest spam
	// target on the public surface (resume-blast bots and submission
	// fuzzing); global 100/min isn't tight enough to stop an abuser
	// from burying an inbox with junk.
	serverHttp.RegisterPOSTWithValidationAndMiddleware[handlers.SubmitApplicationRequest, handlers.SubmitApplicationResponse](
		ctx, router, handlers.SubmitApplicationRoutes(),
		&handlers.SubmitApplicationHandler{CareersService: careersService},
		serverHttp.PerRouteRateLimit(5, time.Minute),
	)

	// Resume upload flow — candidate presigns, browser PUTs direct to
	// R2, HR fetches signed reads on demand. The presign endpoint is
	// anonymous; the read endpoint requires auth (see module.go).
	serverHttp.RegisterPOSTWithValidation[handlers.ResumeUploadURLRequest, handlers.ResumeUploadURLResponse](
		ctx, router, handlers.ResumeUploadURLRoutes(),
		&handlers.ResumeUploadURLHandler{CareersService: careersService},
	)
	serverHttp.RegisterGETWithValidation[handlers.ResumeDownloadURLRequest, handlers.ResumeDownloadURLResponse](
		ctx, router, handlers.ResumeDownloadURLRoutes(),
		&handlers.ResumeDownloadURLHandler{CareersService: careersService},
	)

	// OG card renders PNG bytes, not JSON — it doesn't fit the typed
	// Request/Response helpers, so we register it as a raw chi handler.
	// The auth middleware still respects GetAnonymousRoutes below.
	ogHandler := handlers.PublicOGHandler(careersService)
	for _, pattern := range handlers.PublicOGRoutes() {
		router.Get(pattern, ogHandler)
	}

	// Content sections (perks, culture, testimonials, etc.). All admin
	// endpoints — public pages receive sections inline via
	// ListPublicCareersResponse.
	serverHttp.RegisterGET[handlers.ListSectionsRequest, handlers.ListSectionsResponse](
		ctx, router, handlers.ListSectionsRoutes(),
		&handlers.ListSectionsHandler{CareersService: careersService},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.CreateSectionRequest, handlers.CreateSectionResponse](
		ctx, router, handlers.CreateSectionRoutes(),
		&handlers.CreateSectionHandler{CareersService: careersService},
	)
	serverHttp.RegisterPUTWithValidation[handlers.UpdateSectionRequest, handlers.UpdateSectionResponse](
		ctx, router, handlers.UpdateSectionRoutes(),
		&handlers.UpdateSectionHandler{CareersService: careersService},
	)
	serverHttp.RegisterDELETEWithValidation[handlers.DeleteSectionRequest, handlers.DeleteSectionResponse](
		ctx, router, handlers.DeleteSectionRoutes(),
		&handlers.DeleteSectionHandler{CareersService: careersService},
	)
	serverHttp.RegisterPUTWithValidation[handlers.ReorderSectionsRequest, handlers.ReorderSectionsResponse](
		ctx, router, handlers.ReorderSectionsRoutes(),
		&handlers.ReorderSectionsHandler{CareersService: careersService},
	)

	// Per-career form fields. Admin-only endpoints — public apply
	// receives the field list inline in GetPublicCareerResponse.
	serverHttp.RegisterGETWithValidation[handlers.ListFormFieldsRequest, handlers.ListFormFieldsResponse](
		ctx, router, handlers.ListFormFieldsRoutes(),
		&handlers.ListFormFieldsHandler{CareersService: careersService},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.CreateFormFieldRequest, handlers.CreateFormFieldResponse](
		ctx, router, handlers.CreateFormFieldRoutes(),
		&handlers.CreateFormFieldHandler{CareersService: careersService},
	)
	serverHttp.RegisterPUTWithValidation[handlers.UpdateFormFieldRequest, handlers.UpdateFormFieldResponse](
		ctx, router, handlers.UpdateFormFieldRoutes(),
		&handlers.UpdateFormFieldHandler{CareersService: careersService},
	)
	serverHttp.RegisterDELETEWithValidation[handlers.DeleteFormFieldRequest, handlers.DeleteFormFieldResponse](
		ctx, router, handlers.DeleteFormFieldRoutes(),
		&handlers.DeleteFormFieldHandler{CareersService: careersService},
	)

	// Voxire corporate website apply — multipart/form-data with CV upload.
	// Separate from the tenant-scoped recruit product flow above.
	for _, path := range handlers.ApplyRoutes() {
		router.With(serverHttp.PerRouteRateLimit(5, time.Minute)).
			Post(path, handlers.ApplyHandler(ctx, careersService))
	}
}
