package routes

import (
	"context"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/voxhire/internal/application"
	"github.com/voxire/onion/internal/modules/voxhire/internal/routes/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	// Public company-side hiring inquiry — JSON, validated.
	// Tighter per-route rate limit because this is a public surface
	// that bots will try to spam. 10/min/IP is generous for real use.
	serverHttp.RegisterPOSTWithValidationAndMiddleware[handlers.ContactRequest, handlers.ContactResponse](
		ctx, router, handlers.ContactRoutes(),
		&handlers.ContactHandler{Service: service},
		serverHttp.PerRouteRateLimit(10, time.Minute),
	)

	// Public talent-side application — multipart/form-data with CV
	// upload. The generic JSON wrapHandler can't parse multipart, so
	// this handler is wired directly to chi the same way careers/apply
	// is. We register it once per route path returned by ApplicationsRoutes().
	for _, path := range handlers.ApplicationsRoutes() {
		router.With(serverHttp.PerRouteRateLimit(5, time.Minute)).
			Post(path, handlers.ApplicationsHandler(service))
	}

	// ── Recruitment service: public careers apply + recruiter admin ──
	// Public application against a specific opening (multipart + CV).
	router.With(serverHttp.PerRouteRateLimit(5, time.Minute)).
		Post(handlers.CareerApplyRoute(), handlers.CareerApplyHandler(service))

	// Admin login — anonymous, tighter limit (login spam target).
	serverHttp.RegisterPOSTWithValidationAndMiddleware[handlers.AdminLoginRequest, handlers.AdminLoginResponse](
		ctx, router, []string{handlers.AdminLoginRoute()},
		&handlers.AdminLoginValidatedHandler{Service: service},
		serverHttp.PerRouteRateLimit(10, time.Minute),
	)

	// Admin panel — self-guarded by the admin JWT middleware.
	router.Group(func(gr chi.Router) {
		gr.Use(handlers.AdminAuthMiddleware(service))

		gr.Get(handlers.AdminMeRoute(), handlers.AdminMeHandler(service))
		gr.Get(handlers.AdminAdminsRoute(), handlers.ListAdminsHandler(service))
		serverHttp.RegisterPOSTWithValidation[handlers.CreateAdminRequest, handlers.CreateAdminResponse](
			ctx, gr, []string{handlers.AdminAdminsRoute()},
			&handlers.CreateAdminValidatedHandler{Service: service},
		)
		gr.Get(handlers.AdminAnalyticsRoute(), handlers.AnalyticsHandler(service))
		gr.Get(handlers.AdminOpeningsRoute(), handlers.ListOpeningsHandler(service))
		serverHttp.RegisterPOSTWithValidation[handlers.AdminCareerRequest, handlers.CareerMutationResponse](
			ctx, gr, []string{handlers.AdminCareersRoute()},
			&handlers.CreateCareerValidatedHandler{Service: service},
		)
		serverHttp.RegisterPUTWithValidation[handlers.AdminCareerRequest, handlers.CareerMutationResponse](
			ctx, gr, []string{handlers.AdminCareerRoute()},
			&handlers.UpdateCareerValidatedHandler{Service: service},
		)
		serverHttp.RegisterPUTWithValidation[handlers.SetCareerActiveRequest, handlers.CareerMutationResponse](
			ctx, gr, []string{handlers.AdminCareerActiveRoute()},
			&handlers.SetCareerActiveValidatedHandler{Service: service},
		)
		gr.Get(handlers.AdminLeadsRoute(), handlers.ListLeadsHandler(service))
		gr.Get(handlers.AdminCandidatesRoute(), handlers.ListCandidatesHandler(service))
		gr.Get(handlers.AdminCVJobsRoute(), handlers.ListCVJobsHandler(service))
		serverHttp.RegisterPUTWithValidation[handlers.RetryCVJobRequest, handlers.StatusResponse](
			ctx, gr, []string{handlers.AdminCVJobRetryRoute()},
			&handlers.RetryCVJobValidatedHandler{Service: service},
		)
		gr.Get(handlers.AdminTasksRoute(), handlers.ListTasksHandler(service))
		serverHttp.RegisterPOSTWithValidation[handlers.AdminTaskRequest, handlers.TaskMutationResponse](
			ctx, gr, []string{handlers.AdminTasksRoute()},
			&handlers.CreateTaskValidatedHandler{Service: service},
		)
		serverHttp.RegisterPUTWithValidation[handlers.AdminTaskRequest, handlers.TaskMutationResponse](
			ctx, gr, []string{handlers.AdminTaskRoute()},
			&handlers.UpdateTaskValidatedHandler{Service: service},
		)
		gr.Get(handlers.AdminShortlistsRoute(), handlers.ListShortlistsHandler(service))
		serverHttp.RegisterPOSTWithValidation[handlers.AdminShortlistRequest, handlers.ShortlistMutationResponse](
			ctx, gr, []string{handlers.AdminShortlistsRoute()},
			&handlers.CreateShortlistValidatedHandler{Service: service},
		)
		serverHttp.RegisterPOSTWithValidation[handlers.AddShortlistItemRequest, handlers.AddShortlistItemResponse](
			ctx, gr, []string{handlers.AdminShortlistItemsRoute()},
			&handlers.AddShortlistItemValidatedHandler{Service: service},
		)
		gr.Get(handlers.AdminApplicationsRoute(), handlers.ListApplicationsHandler(service))
		serverHttp.RegisterPUTWithValidation[handlers.UpdateApplicationStatusRequest, handlers.StatusResponse](
			ctx, gr, []string{handlers.AdminAppStatusRoute()},
			&handlers.UpdateApplicationStatusValidatedHandler{Service: service},
		)
		serverHttp.RegisterPUTWithValidation[handlers.UpdateApplicationReviewRequest, handlers.StatusResponse](
			ctx, gr, []string{handlers.AdminAppReviewRoute()},
			&handlers.UpdateApplicationReviewValidatedHandler{Service: service},
		)
		gr.Get(handlers.AdminAppCVRoute(), handlers.ApplicationCVHandler(service))
		gr.Get(handlers.AdminAppTimelineRoute(), handlers.ApplicationTimelineHandler(service))
	})
}
