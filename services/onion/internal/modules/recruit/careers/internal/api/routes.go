package api

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/recruit/careers/internal/application"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

// RegisterRoutes wires the legacy /careers endpoint (global, voxire-
// corporate list) plus the recruit product endpoints (tenant-scoped,
// powering recruit.rtylr.com).
func RegisterRoutes(ctx context.Context, router serverHttp.LegacyRouter, careersService *application.Service) {
	// Legacy: voxire-corporate global careers list. Kept for back-compat
	// while the landing site is migrated to the tenant-scoped recruit
	// flow; once the voxire careers page redirects to recruit.rtylr.com
	// this handler can be retired.
	serverHttp.RegisterGET[GetActiveCareersRequest, GetActiveCareersResponse](
		ctx, router, CareersRoutes(),
		&CareersHandler{CareersService: careersService},
	)

	// Recruit admin — authenticated, tenant-scoped.
	serverHttp.RegisterGET[ListRecruitCareersRequest, ListRecruitCareersResponse](
		ctx, router, ListRecruitCareersRoutes(),
		&ListRecruitCareersHandler{CareersService: careersService},
	)
	serverHttp.RegisterPOSTWithValidation[CreateRecruitCareerRequest, CreateRecruitCareerResponse](
		ctx, router, CreateRecruitCareerRoutes(),
		&CreateRecruitCareerHandler{CareersService: careersService},
	)
	serverHttp.RegisterPUTWithValidation[UpdateRecruitCareerRequest, UpdateRecruitCareerResponse](
		ctx, router, UpdateRecruitCareerRoutes(),
		&UpdateRecruitCareerHandler{CareersService: careersService},
	)
	serverHttp.RegisterDELETEWithValidation[DeleteRecruitCareerRequest, DeleteRecruitCareerResponse](
		ctx, router, DeleteRecruitCareerRoutes(),
		&DeleteRecruitCareerHandler{CareersService: careersService},
	)

	serverHttp.RegisterGET[GetCustomizationRequest, GetCustomizationResponse](
		ctx, router, GetRecruitCustomizationRoutes(),
		&GetCustomizationHandler{CareersService: careersService},
	)
	serverHttp.RegisterPUTWithValidation[PutCustomizationRequest, PutCustomizationResponse](
		ctx, router, PutRecruitCustomizationRoutes(),
		&PutCustomizationHandler{CareersService: careersService},
	)

	serverHttp.RegisterGET[ListSubmissionsRequest, ListSubmissionsResponse](
		ctx, router, ListSubmissionsRoutes(),
		&ListSubmissionsHandler{CareersService: careersService},
	)
	serverHttp.RegisterPUTWithValidation[UpdateSubmissionStatusRequest, UpdateSubmissionStatusResponse](
		ctx, router, UpdateSubmissionStatusRoutes(),
		&UpdateSubmissionStatusHandler{CareersService: careersService},
	)
	serverHttp.RegisterGET[ExportSubmissionsRequest, ExportSubmissionsResponse](
		ctx, router, ExportSubmissionsRoutes(),
		&ExportSubmissionsHandler{CareersService: careersService},
	)

	// Recruit public — anonymous, powers the candidate-facing site.
	serverHttp.RegisterGETWithValidation[GetPublicBrandingRequest, GetPublicBrandingResponse](
		ctx, router, GetPublicBrandingRoutes(),
		&GetPublicBrandingHandler{CareersService: careersService},
	)
	serverHttp.RegisterGETWithValidation[ListPublicCareersRequest, ListPublicCareersResponse](
		ctx, router, ListPublicCareersRoutes(),
		&ListPublicCareersHandler{CareersService: careersService},
	)
	serverHttp.RegisterGETWithValidation[GetPublicCareerRequest, GetPublicCareerResponse](
		ctx, router, GetPublicCareerRoutes(),
		&GetPublicCareerHandler{CareersService: careersService},
	)
	// Public apply — 5/min/IP. This is the single biggest spam
	// target on the public surface (resume-blast bots and submission
	// fuzzing); global 100/min isn't tight enough to stop an abuser
	// from burying an inbox with junk.
	serverHttp.RegisterPOSTWithValidationAndMiddleware[SubmitApplicationRequest, SubmitApplicationResponse](
		ctx, router, SubmitApplicationRoutes(),
		&SubmitApplicationHandler{CareersService: careersService},
		serverHttp.PerRouteRateLimit(5, time.Minute),
	)

	// Resume upload flow — candidate presigns, browser PUTs direct to
	// R2, HR fetches signed reads on demand. The presign endpoint is
	// anonymous; the read endpoint requires auth (see module.go).
	serverHttp.RegisterPOSTWithValidationAndMiddleware[ResumeUploadURLRequest, ResumeUploadURLResponse](
		ctx, router, ResumeUploadURLRoutes(),
		&ResumeUploadURLHandler{CareersService: careersService},
		serverHttp.PerRouteRateLimit(10, time.Minute),
	)
	serverHttp.RegisterGETWithValidation[ResumeDownloadURLRequest, ResumeDownloadURLResponse](
		ctx, router, ResumeDownloadURLRoutes(),
		&ResumeDownloadURLHandler{CareersService: careersService},
	)

	// OG card renders PNG bytes, not JSON — it doesn't fit the typed
	// Request/Response helpers, so we register it as a raw chi handler.
	// The auth middleware still respects GetAnonymousRoutes below.
	ogHandler := PublicOGHandler(careersService)
	for _, pattern := range PublicOGRoutes() {
		router.Get(pattern, ogHandler)
	}

	// Content sections (perks, culture, testimonials, etc.). All admin
	// endpoints — public pages receive sections inline via
	// ListPublicCareersResponse.
	serverHttp.RegisterGET[ListSectionsRequest, ListSectionsResponse](
		ctx, router, ListSectionsRoutes(),
		&ListSectionsHandler{CareersService: careersService},
	)
	serverHttp.RegisterPOSTWithValidation[CreateSectionRequest, CreateSectionResponse](
		ctx, router, CreateSectionRoutes(),
		&CreateSectionHandler{CareersService: careersService},
	)
	serverHttp.RegisterPUTWithValidation[UpdateSectionRequest, UpdateSectionResponse](
		ctx, router, UpdateSectionRoutes(),
		&UpdateSectionHandler{CareersService: careersService},
	)
	serverHttp.RegisterDELETEWithValidation[DeleteSectionRequest, DeleteSectionResponse](
		ctx, router, DeleteSectionRoutes(),
		&DeleteSectionHandler{CareersService: careersService},
	)
	serverHttp.RegisterPUTWithValidation[ReorderSectionsRequest, ReorderSectionsResponse](
		ctx, router, ReorderSectionsRoutes(),
		&ReorderSectionsHandler{CareersService: careersService},
	)

	// Per-career form fields. Admin-only endpoints — public apply
	// receives the field list inline in GetPublicCareerResponse.
	serverHttp.RegisterGETWithValidation[ListFormFieldsRequest, ListFormFieldsResponse](
		ctx, router, ListFormFieldsRoutes(),
		&ListFormFieldsHandler{CareersService: careersService},
	)
	serverHttp.RegisterPOSTWithValidation[CreateFormFieldRequest, CreateFormFieldResponse](
		ctx, router, CreateFormFieldRoutes(),
		&CreateFormFieldHandler{CareersService: careersService},
	)
	serverHttp.RegisterPUTWithValidation[UpdateFormFieldRequest, UpdateFormFieldResponse](
		ctx, router, UpdateFormFieldRoutes(),
		&UpdateFormFieldHandler{CareersService: careersService},
	)
	serverHttp.RegisterDELETEWithValidation[DeleteFormFieldRequest, DeleteFormFieldResponse](
		ctx, router, DeleteFormFieldRoutes(),
		&DeleteFormFieldHandler{CareersService: careersService},
	)

	// Voxire corporate website apply — multipart/form-data with CV upload.
	// Separate from the tenant-scoped recruit product flow above.
	for _, path := range ApplyRoutes() {
		router.With(serverHttp.PerRouteRateLimit(5, time.Minute)).
			Post(path, ApplyHandler(ctx, careersService))
	}
}
