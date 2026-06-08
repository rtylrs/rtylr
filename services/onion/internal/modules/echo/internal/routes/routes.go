package routes

import (
	"context"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/echo/internal/application"
	"github.com/voxire/onion/internal/modules/echo/internal/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

// aiStreamRateLimit is 5 requests per minute per authenticated user.
// The limiter is package-level so all paths returned by
// GenerateTextStreamRoutes() share the same bucket state.
var aiStreamRateLimit = serverHttp.AIUserRateLimit(5, time.Minute)

func RegisterRoutes(ctx context.Context, r *chi.Mux, snipsService *application.Service, memoryService handlers.MemoryService, quotaService handlers.QuotaService, analystService handlers.AnalystService) {
	serverHttp.RegisterGET[handlers.GetProviderStatusRequest, handlers.GetProviderStatusResponse](
		ctx,
		r,
		handlers.ProviderStatusRoutes(),
		&handlers.ProviderStatusHandler{SnipsService: snipsService},
	)

	serverHttp.RegisterPOSTWithValidation[handlers.ListToolsRequest, handlers.ListToolsResponse](
		ctx,
		r,
		handlers.ListToolsRoutes(),
		&handlers.ListToolsHandler{},
	)

	// SSE endpoints are registered manually — they stream tokens rather than
	// following the standard request/response pattern used by RegisterPOST.
	// A per-user rate limit (5 req/min) is applied to control AI generation cost.
	for _, path := range handlers.GenerateTextStreamRoutes() {
		streamHandler := &handlers.GenerateTextStreamHandler{
			AiService:      snipsService,
			MemoryService:  memoryService,
			QuotaService:   quotaService,
			AnalystService: analystService,
		}
		r.With(aiStreamRateLimit).Post(path, streamHandler.ServeHTTP)
	}
}
