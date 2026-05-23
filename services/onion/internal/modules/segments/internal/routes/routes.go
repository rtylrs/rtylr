package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/segments/internal/application"
	"github.com/voxire/onion/internal/modules/segments/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, svc *application.Service) {
	http.RegisterGET[handlers.ListSegmentsRequest, handlers.ListSegmentsResponse](
		ctx, router, handlers.SegmentListRoutes(),
		&handlers.ListSegmentsHandler{Service: svc},
	)

	http.RegisterPOSTWithValidation[handlers.CreateSegmentRequest, handlers.CreateSegmentResponse](
		ctx, router, handlers.SegmentCreateRoutes(),
		&handlers.CreateSegmentHandler{Service: svc},
	)

	http.RegisterPUTWithValidation[handlers.UpdateSegmentRequest, handlers.UpdateSegmentResponse](
		ctx, router, handlers.SegmentUpdateRoutes(),
		&handlers.UpdateSegmentHandler{Service: svc},
	)

	http.RegisterDELETEWithValidation[handlers.DeleteSegmentRequest, handlers.DeleteSegmentResponse](
		ctx, router, handlers.SegmentDeleteRoutes(),
		&handlers.DeleteSegmentHandler{Service: svc},
	)
}
