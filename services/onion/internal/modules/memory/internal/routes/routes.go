package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	http "github.com/voxire/onion/internal/platform/server/http"
	"github.com/voxire/onion/internal/modules/memory/internal/application"
	"github.com/voxire/onion/internal/modules/memory/internal/routes/handlers"
)

func RegisterRoutes(ctx context.Context, r *chi.Mux, memoryService *application.Service) {
	http.RegisterPOST[handlers.SessionRequest, handlers.SessionResponse](
		ctx,
		r,
		handlers.SessionRoutes(),
		&handlers.SessionHandler{MemoryService: memoryService},
	)

	http.RegisterPOST[handlers.MemoryRequest, handlers.MemoryResponse](
		ctx,
		r,
		handlers.MemoryRoutes(),
		&handlers.MemoryHandler{MemoryService: memoryService},
	)

	http.RegisterPOST[handlers.ContextRequest, handlers.ContextResponse](
		ctx,
		r,
		handlers.ContextRoutes(),
		&handlers.ContextHandler{MemoryService: memoryService},
	)

	http.RegisterPOST[handlers.SearchRequest, handlers.SearchResponse](
		ctx,
		r,
		handlers.SearchRoutes(),
		&handlers.SearchHandler{MemoryService: memoryService},
	)

	http.RegisterPOST[handlers.StatsRequest, handlers.StatsResponse](
		ctx,
		r,
		handlers.StatsRoutes(),
		&handlers.StatsHandler{MemoryService: memoryService},
	)

	http.RegisterPOST[handlers.SummarizeRequest, handlers.SummarizeResponse](
		ctx,
		r,
		handlers.SummarizeRoutes(),
		&handlers.SummarizeHandler{MemoryService: memoryService},
	)

	http.RegisterPOST[handlers.ChatsRequest, handlers.ChatsResponse](
		ctx,
		r,
		handlers.ChatsRoutes(),
		&handlers.ChatHistoryHandler{MemoryService: memoryService},
	)

	http.RegisterPOST[handlers.ChatMessagesRequest, handlers.ChatMessagesResponse](
		ctx,
		r,
		handlers.ChatMessagesRoutes(),
		&handlers.ChatMessagesHandler{MemoryService: memoryService},
	)
}
