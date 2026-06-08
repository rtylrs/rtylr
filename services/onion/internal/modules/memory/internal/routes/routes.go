package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/memory/internal/application"
	"github.com/voxire/onion/internal/modules/memory/internal/routes/handlers"
	http "github.com/voxire/onion/internal/platform/server/http"
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

	http.RegisterGET[handlers.ChatsRequest, handlers.ChatsResponse](
		ctx,
		r,
		handlers.ChatsRoutes(),
		&handlers.ChatHistoryHandler{MemoryService: memoryService},
	)

	http.RegisterGET[handlers.ChatMessagesRequest, handlers.ChatMessagesResponse](
		ctx,
		r,
		handlers.ChatMessagesRoutes(),
		&handlers.ChatMessagesHandler{MemoryService: memoryService},
	)

	http.RegisterDELETE[handlers.DeleteChatRequest, handlers.DeleteChatResponse](
		ctx,
		r,
		handlers.DeleteChatRoutes(),
		&handlers.DeleteChatHandler{Service: memoryService},
	)

	http.RegisterPUTWithValidation[handlers.RenameChatRequest, handlers.RenameChatResponse](
		ctx,
		r,
		handlers.RenameChatRoutes(),
		&handlers.RenameChatHandler{Service: memoryService},
	)

	http.RegisterPUT[handlers.PinChatRequest, handlers.PinChatResponse](
		ctx,
		r,
		handlers.PinChatRoutes(),
		&handlers.PinChatHandler{Service: memoryService},
	)
}
