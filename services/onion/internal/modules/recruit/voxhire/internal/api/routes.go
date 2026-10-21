package api

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/recruit/voxhire/internal/application"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router serverHttp.LegacyRouter, service *application.Service) {
	// Public company-side hiring inquiry — JSON, validated.
	// Tighter per-route rate limit because this is a public surface
	// that bots will try to spam. 10/min/IP is generous for real use.
	serverHttp.RegisterPOSTWithValidationAndMiddleware[ContactRequest, ContactResponse](
		ctx, router, ContactRoutes(),
		&ContactHandler{Service: service},
		serverHttp.PerRouteRateLimit(10, time.Minute),
	)

	// Public talent-side application — multipart/form-data with CV
	// upload. The generic JSON wrapHandler can't parse multipart, so
	// this handler is wired directly to chi the same way careers/apply
	// is. We register it once per route path returned by ApplicationsRoutes().
	for _, path := range ApplicationsRoutes() {
		router.With(serverHttp.PerRouteRateLimit(5, time.Minute)).
			Post(path, ApplicationsHandler(service))
	}

	// ── Recruitment service: public careers apply + recruiter admin ──
	// Public application against a specific opening (multipart + CV).
	router.With(serverHttp.PerRouteRateLimit(5, time.Minute)).
		Post(CareerApplyRoute(), CareerApplyHandler(service))

	// Admin login — anonymous, tighter limit (login spam target).
	serverHttp.RegisterPOSTWithValidationAndMiddleware[AdminLoginRequest, AdminLoginResponse](
		ctx, router, []string{AdminLoginRoute()},
		&AdminLoginValidatedHandler{Service: service},
		serverHttp.PerRouteRateLimit(10, time.Minute),
	)

	// Admin panel — self-guarded by the admin JWT middleware.
	{
		gr := router.With(AdminAuthMiddleware(service))

		gr.Get(AdminMeRoute(), AdminMeHandler(service))
		gr.Get(AdminAdminsRoute(), ListAdminsHandler(service))
		serverHttp.RegisterPOSTWithValidation[CreateAdminRequest, CreateAdminResponse](
			ctx, gr, []string{AdminAdminsRoute()},
			&CreateAdminValidatedHandler{Service: service},
		)
		gr.Get(AdminAnalyticsRoute(), AnalyticsHandler(service))
		gr.Get(AdminOpeningsRoute(), ListOpeningsHandler(service))
		serverHttp.RegisterPOSTWithValidation[AdminCareerRequest, CareerMutationResponse](
			ctx, gr, []string{AdminCareersRoute()},
			&CreateCareerValidatedHandler{Service: service},
		)
		serverHttp.RegisterPUTWithValidation[AdminCareerRequest, CareerMutationResponse](
			ctx, gr, []string{AdminCareerRoute()},
			&UpdateCareerValidatedHandler{Service: service},
		)
		serverHttp.RegisterPUTWithValidation[SetCareerActiveRequest, CareerMutationResponse](
			ctx, gr, []string{AdminCareerActiveRoute()},
			&SetCareerActiveValidatedHandler{Service: service},
		)
		gr.Get(AdminLeadsRoute(), ListLeadsHandler(service))
		gr.Get(AdminCandidatesRoute(), ListCandidatesHandler(service))
		gr.Get(AdminCVJobsRoute(), ListCVJobsHandler(service))
		serverHttp.RegisterPUTWithValidation[RetryCVJobRequest, StatusResponse](
			ctx, gr, []string{AdminCVJobRetryRoute()},
			&RetryCVJobValidatedHandler{Service: service},
		)
		gr.Get(AdminTasksRoute(), ListTasksHandler(service))
		serverHttp.RegisterPOSTWithValidation[AdminTaskRequest, TaskMutationResponse](
			ctx, gr, []string{AdminTasksRoute()},
			&CreateTaskValidatedHandler{Service: service},
		)
		serverHttp.RegisterPUTWithValidation[AdminTaskRequest, TaskMutationResponse](
			ctx, gr, []string{AdminTaskRoute()},
			&UpdateTaskValidatedHandler{Service: service},
		)
		gr.Get(AdminShortlistsRoute(), ListShortlistsHandler(service))
		serverHttp.RegisterPOSTWithValidation[AdminShortlistRequest, ShortlistMutationResponse](
			ctx, gr, []string{AdminShortlistsRoute()},
			&CreateShortlistValidatedHandler{Service: service},
		)
		serverHttp.RegisterPOSTWithValidation[AddShortlistItemRequest, AddShortlistItemResponse](
			ctx, gr, []string{AdminShortlistItemsRoute()},
			&AddShortlistItemValidatedHandler{Service: service},
		)
		gr.Get(AdminApplicationsRoute(), ListApplicationsHandler(service))
		serverHttp.RegisterPUTWithValidation[UpdateApplicationStatusRequest, StatusResponse](
			ctx, gr, []string{AdminAppStatusRoute()},
			&UpdateApplicationStatusValidatedHandler{Service: service},
		)
		serverHttp.RegisterPUTWithValidation[UpdateApplicationReviewRequest, StatusResponse](
			ctx, gr, []string{AdminAppReviewRoute()},
			&UpdateApplicationReviewValidatedHandler{Service: service},
		)
		gr.Get(AdminAppCVRoute(), ApplicationCVHandler(service))
		gr.Get(AdminAppTimelineRoute(), ApplicationTimelineHandler(service))
	}
}
