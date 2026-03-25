package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	echov1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/rtylr/echo/v1"
	http "github.com/voxire/onion/internal/platform/server/http"
	"github.com/voxire/onion/internal/platforms/rtylr/echo/internal/application"
	"github.com/voxire/onion/internal/platforms/rtylr/echo/internal/handlers"
	"github.com/voxire/onion/internal/platforms/rtylr/memory"
)

func RegisterRoutes(ctx context.Context, r *chi.Mux, snipsService *application.Service, memoryService *memory.Module) {
	http.RegisterGET[echov1.GetProviderStatusRequest, echov1.GetProviderStatusResponse](
		ctx,
		r,
		handlers.ProviderStatusRoutes(),
		&handlers.ProviderStatusHandler{SnipsService: snipsService},
	)

	// For SSE endpoints, we need to register them manually since they don't follow the standard request/response pattern
	for _, path := range handlers.GenerateTextStreamRoutes() {
		streamHandler := &handlers.GenerateTextStreamHandler{
			AiService:     snipsService,
			MemoryService: memoryService,
		}
		r.Post(path, streamHandler.ServeHTTP)
	}
}
