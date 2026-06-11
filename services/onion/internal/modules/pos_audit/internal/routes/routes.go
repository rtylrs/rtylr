package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/pos_audit/internal/application"
	"github.com/voxire/onion/internal/modules/pos_audit/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, svc *application.Service) {
	http.RegisterGETWithValidation[handlers.ListAuditRequest, handlers.ListAuditResponse](
		ctx, router, handlers.AuditListRoutes(),
		&handlers.ListAuditHandler{Service: svc},
	)

	http.RegisterPOSTWithValidation[handlers.CreateAuditRequest, handlers.CreateAuditResponse](
		ctx, router, handlers.AuditCreateRoutes(),
		&handlers.CreateAuditHandler{Service: svc},
	)
}
