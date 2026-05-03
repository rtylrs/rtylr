package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/oauth/internal/application"
	"github.com/voxire/onion/internal/modules/oauth/internal/routes/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	// Authorize: raw handler that redirects browser to auth.rtylr.com/oauth
	authorizeHandler := &handlers.AuthorizeHandler{Service: service}
	router.Get("/oauth/authorize", authorizeHandler.ServeHTTP)

	// Issue: called by auth.rtylr.com after user logs in + grants consent (requires Bearer token)
	serverHttp.RegisterPOSTWithValidation[handlers.IssueRequest, handlers.IssueResponse](
		ctx, router, handlers.IssueRoutes(), &handlers.IssueHandler{Service: service},
	)

	// Token exchange (anonymous — no JWT required)
	serverHttp.RegisterPOSTWithValidation[handlers.TokenRequest, handlers.TokenResponse](
		ctx, router, handlers.TokenRoutes(), &handlers.TokenHandler{Service: service},
	)

	// Revoke (anonymous — token may already be expired)
	serverHttp.RegisterPOSTWithValidation[handlers.RevokeRequest, handlers.RevokeResponse](
		ctx, router, handlers.RevokeRoutes(), &handlers.RevokeHandler{Service: service},
	)

	// Me — requires valid Bearer token (auth middleware validates)
	serverHttp.RegisterGET[handlers.MeRequest, handlers.MeResponse](
		ctx, router, handlers.MeRoutes(), &handlers.MeHandler{},
	)
}

// AnonymousRoutes lists the paths that must bypass JWT auth middleware.
func AnonymousRoutes() []string {
	return []string{
		"/oauth/authorize",
		"/oauth/token",
		"/oauth/revoke",
	}
}
