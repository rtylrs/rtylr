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
	router.With(serverHttp.PerRouteRateLimit(10, time.Minute)).
		Post(handlers.AdminLoginRoute(), handlers.AdminLoginHandler(service))

	// Admin panel — self-guarded by the admin JWT middleware.
	router.Group(func(gr chi.Router) {
		gr.Use(handlers.AdminAuthMiddleware(service))

		gr.Get(handlers.AdminMeRoute(), handlers.AdminMeHandler(service))
		gr.Get(handlers.AdminAdminsRoute(), handlers.ListAdminsHandler(service))
		gr.Post(handlers.AdminAdminsRoute(), handlers.CreateAdminHandler(service))
		gr.Get(handlers.AdminAnalyticsRoute(), handlers.AnalyticsHandler(service))
		gr.Get(handlers.AdminOpeningsRoute(), handlers.ListOpeningsHandler(service))
		gr.Post(handlers.AdminCareersRoute(), handlers.CreateCareerHandler(service))
		gr.Put(handlers.AdminCareerRoute(), handlers.UpdateCareerHandler(service))
		gr.Put(handlers.AdminCareerActiveRoute(), handlers.SetCareerActiveHandler(service))
		gr.Get(handlers.AdminLeadsRoute(), handlers.ListLeadsHandler(service))
		gr.Get(handlers.AdminApplicationsRoute(), handlers.ListApplicationsHandler(service))
		gr.Put(handlers.AdminAppStatusRoute(), handlers.UpdateApplicationStatusHandler(service))
		gr.Get(handlers.AdminAppCVRoute(), handlers.ApplicationCVHandler(service))
	})
}
