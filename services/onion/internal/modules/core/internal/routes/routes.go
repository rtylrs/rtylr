package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/core/internal/application"
	"github.com/voxire/onion/internal/modules/core/internal/routes/handlers"
	corev1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/shared/core/v1"
	httpserver "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, r *chi.Mux, service *application.Service) {
	httpserver.RegisterPOSTWithValidation[corev1.SlackErrorLogRequest, corev1.SlackErrorLogResponse](
		ctx,
		r,
		handlers.SlackDashboardSendRoutes(),
		&handlers.SlackDashboardSendHandler{Service: service},
	)
}
