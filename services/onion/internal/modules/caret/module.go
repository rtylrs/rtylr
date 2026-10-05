package echo

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/caret/internal/api"
	"github.com/voxire/onion/internal/modules/caret/internal/application"
	"github.com/voxire/onion/internal/modules/caret/internal/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
	hubcaretv1 "github.com/voxire/voxie/gen/go/hub/caret/v1"
)

var aiStreamRateLimit = serverHttp.AIUserRateLimit(5, time.Minute)

type Module struct {
	snipsService   *application.Service
	memoryService  application.MemoryService
	quotaService   handlers.QuotaService
	analystService application.AnalystService
	toolExecutors  []application.ToolExecutor
}

func NewModule(ctx context.Context, memoryService application.MemoryService, quotaService handlers.QuotaService, analystService application.AnalystService, toolExecutors ...application.ToolExecutor) *Module {
	if err := application.ValidateToolRegistry(); err != nil {
		panic(fmt.Errorf("invalid Caret tool registry: %w", err))
	}
	snipsService, err := application.NewService(ctx)
	if err != nil {
		snipsService = nil
	}
	if analystService != nil {
		toolExecutors = append([]application.ToolExecutor{
			application.NewBusinessSnapshotToolExecutor(analystService),
			application.NewAnalyticsQueryToolExecutor(snipsService, analystService),
		}, toolExecutors...)
	}
	if profile, ok := memoryService.(application.AssistantProfileReader); ok {
		toolExecutors = append([]application.ToolExecutor{application.NewAssistantToolExecutor(profile, nil)}, toolExecutors...)
	}

	return &Module{
		snipsService:   snipsService,
		memoryService:  memoryService,
		quotaService:   quotaService,
		analystService: analystService,
		toolExecutors:  toolExecutors,
	}
}

func (m *Module) GetSnipsService() *application.Service {
	return m.snipsService
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := hubcaretv1.RegisterCaretServiceServer(
		api.NewCaretAPI(),
		hubcaretv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register hub caret service: %w", err))
	}
}

func (m *Module) RegisterRoutes(ctx context.Context, r *chi.Mux, _ ...interface{}) {
	serverHttp.RegisterGET[handlers.ListToolsRequest, handlers.ListToolsResponse](
		ctx,
		r,
		handlers.ListToolsRoutes(),
		&handlers.ListToolsHandler{},
	)

	serverHttp.RegisterGET[handlers.GetProviderStatusRequest, handlers.GetProviderStatusResponse](
		ctx,
		r,
		handlers.ProviderStatusRoutes(),
		&handlers.ProviderStatusHandler{SnipsService: m.snipsService},
	)

	serverHttp.RegisterPOSTWithValidation[handlers.ToolDecisionRequest, handlers.ToolDecisionResponse](
		ctx,
		r,
		handlers.ToolDecisionRoutes(),
		&handlers.ToolDecisionHandler{Executors: m.toolExecutors, RequirePendingApproval: true},
	)

	for _, path := range handlers.GenerateTextStreamRoutes() {
		streamHandler := &handlers.GenerateTextStreamHandler{
			AiService:      m.snipsService,
			MemoryService:  m.memoryService,
			QuotaService:   m.quotaService,
			AnalystService: m.analystService,
			ToolExecutors:  m.toolExecutors,
		}
		r.With(aiStreamRateLimit).Post(path, streamHandler.ServeHTTP)
	}
}
